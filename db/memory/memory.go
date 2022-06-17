package memory

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	icache "github.com/donetkit/contrib/utils/cache"
	"io"
	"os"
	"sync"
	"time"
)

type Cache struct {
	sync.Mutex
	ctxCache context.Context
	config   *config

	defaultExpiration time.Duration
	items             map[string]*item
	janitor           *janitor
}

func (c *Cache) WithDB(db int) icache.ICache {
	return c
}

func (c *Cache) WithContext(ctx context.Context) icache.ICache {
	return c
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *Cache) Get(k string) interface{} {
	c.Lock()
	x, found := c.get(k)
	c.Unlock()
	if found {
		return x
	}
	return nil
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *Cache) GetString(k string) (string, error) {
	c.Lock()
	x, found := c.get(k)
	c.Unlock()
	if found {
		return x.(string), nil
	}
	return "", nil
}

// Add an item to the cache, replacing any existing item. If the duration is 0,
// the cache's default expiration time is used. If it is -1, the item never
// expires.
func (c *Cache) Set(key string, value interface{}, timeout time.Duration) error {
	c.Lock()
	c.set(key, value, timeout)
	// TODO: Calls to mu.Unlock are currently not deferred because defer
	// adds ~200 ns (as of go1.)
	c.Unlock()
	return nil
}

func (c *Cache) set(k string, x interface{}, d time.Duration) {
	var e *time.Time
	if d == 0 {
		d = c.defaultExpiration
	}
	if d > 0 {
		t := time.Now().Add(d)
		e = &t
	}
	c.items[k] = &item{
		Object:     x,
		Expiration: e,
	}
}

func (c *Cache) get(k string) (interface{}, bool) {
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.IsExist() {
		c.delete(k)
		return nil, false
	}
	return item.Object, true
}

func (c *Cache) IsExist(k string) bool {
	_, found := c.items[k]
	if !found {
		return false
	}
	return true
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (c *Cache) Delete(k string) (int64, error) {
	c.Lock()
	defer c.Unlock()
	_, found := c.get(k)
	if !found {
		return 0, errors.New("not found")
	}
	c.delete(k)
	return 1, nil
}

// Increment an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to increment it by n. Pass a negative number to decrement the value.
func (c *Cache) IncrementFloat(k string, n float64) (float64, error) {
	c.Lock()
	v, found := c.items[k]
	if !found || v.IsExist() {
		c.Unlock()
		return n, fmt.Errorf("item not found")
	}
	switch v.Object.(type) {
	case float32:
		v.Object = v.Object.(float32) + float32(n)
	case float64:
		v.Object = v.Object.(float64) + n
	default:
		c.Unlock()
		return n, fmt.Errorf("The value for %s does not have type float32 or float64", k)
	}
	c.Unlock()
	return v.Object.(float64), nil
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or int64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n.
// Wraps around on overlow.
func (c *Cache) Increment(k string, n int64) (int64, error) {
	c.Lock()
	defer c.Unlock()
	v, found := c.items[k]
	if !found || v.IsExist() {
		return 0, ErrCacheMiss
	}
	switch v.Object.(type) {
	case int:
		v.Object = v.Object.(int) + int(n)
		return int64(v.Object.(int)), nil
	case int8:
		v.Object = v.Object.(int8) + int8(n)
		return int64(v.Object.(int8)), nil
	case int16:
		v.Object = v.Object.(int16) + int16(n)
		return int64(v.Object.(int16)), nil
	case int32:
		v.Object = v.Object.(int32) + int32(n)
		return int64(v.Object.(int32)), nil
	case int64:
		v.Object = v.Object.(int64) + int64(n)
		return int64(v.Object.(int64)), nil
	case uint:
		v.Object = v.Object.(uint) + uint(n)
		return int64(v.Object.(uint)), nil
	case uintptr:
		v.Object = v.Object.(uintptr) + uintptr(n)
		return int64(v.Object.(uintptr)), nil
	case uint8:
		v.Object = v.Object.(uint8) + uint8(n)
		return int64(v.Object.(uint8)), nil
	case uint16:
		v.Object = v.Object.(uint16) + uint16(n)
		return int64(v.Object.(uint16)), nil
	case uint32:
		v.Object = v.Object.(uint32) + uint32(n)
		return int64(v.Object.(uint32)), nil
	case uint64:
		v.Object = v.Object.(uint64) + uint64(n)
		return int64(v.Object.(uint64)), nil
	}
	return 0, fmt.Errorf("The value for %s is not an integer", k)
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or int64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n.
// Stops at 0 on underflow.
func (c *Cache) Decrement(k string, n int64) (int64, error) {
	// TODO: Implement Increment and Decrement more cleanly.
	// (Cannot do Increment(k, n*-1) for uints.)
	c.Lock()
	defer c.Unlock()
	v, found := c.items[k]
	if !found || v.IsExist() {
		return 0, ErrCacheMiss
	}
	switch v.Object.(type) {
	case int:
		vi := v.Object.(int)
		if vi > int(n) {
			v.Object = vi - int(n)
		} else {
			v.Object = int(0)
		}
		return int64(v.Object.(int)), nil
	case int8:
		vi := v.Object.(int8)
		if vi > int8(n) {
			v.Object = vi - int8(n)
		} else {
			v.Object = int8(0)
		}
		return int64(v.Object.(int8)), nil
	case int16:
		vi := v.Object.(int16)
		if vi > int16(n) {
			v.Object = vi - int16(n)
		} else {
			v.Object = int16(0)
		}
		return int64(v.Object.(int16)), nil
	case int32:
		vi := v.Object.(int32)
		if vi > int32(n) {
			v.Object = vi - int32(n)
		} else {
			v.Object = int32(0)
		}
		return int64(v.Object.(int32)), nil
	case int64:
		vi := v.Object.(int64)
		if vi > int64(n) {
			v.Object = vi - int64(n)
		} else {
			v.Object = int64(0)
		}
		return int64(v.Object.(int64)), nil
	case uint:
		vi := v.Object.(uint)
		if vi > uint(n) {
			v.Object = vi - uint(n)
		} else {
			v.Object = uint(0)
		}
		return int64(v.Object.(uint)), nil
	case uintptr:
		vi := v.Object.(uintptr)
		if vi > uintptr(n) {
			v.Object = vi - uintptr(n)
		} else {
			v.Object = uintptr(0)
		}
		return int64(v.Object.(uintptr)), nil
	case uint8:
		vi := v.Object.(uint8)
		if vi > uint8(n) {
			v.Object = vi - uint8(n)
		} else {
			v.Object = uint8(0)
		}
		return int64(v.Object.(uint8)), nil
	case uint16:
		vi := v.Object.(uint16)
		if vi > uint16(n) {
			v.Object = vi - uint16(n)
		} else {
			v.Object = uint16(0)
		}
		return int64(v.Object.(uint16)), nil
	case uint32:
		vi := v.Object.(uint32)
		if vi > uint32(n) {
			v.Object = vi - uint32(n)
		} else {
			v.Object = uint32(0)
		}
		return int64(v.Object.(uint32)), nil
	case uint64:
		vi := v.Object.(int64)
		if vi > int64(n) {
			v.Object = vi - int64(n)
		} else {
			v.Object = int64(0)
		}
		return v.Object.(int64), nil
	}
	return 0, fmt.Errorf("The value for %s is not an integer", k)
}

func (c *Cache) delete(k string) {
	delete(c.items, k)
}

// Delete all expired items from the cache.
func (c *Cache) DeleteExpired() {
	c.Lock()
	for k, v := range c.items {
		if v.IsExist() {
			c.delete(k)
		}
	}
	c.Unlock()
}

// Write the cache's items (using Gob) to an io.Writer.
func (c *Cache) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)

	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Error registering item types with Gob library")
		}
	}()
	for _, v := range c.items {
		gob.Register(v.Object)
	}
	err = enc.Encode(&c.items)
	return
}

// Save the cache's items to the given filename, creating the file if it
// doesn't exist, and overwriting it if it does.
func (c *Cache) SaveFile(fname string) error {
	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	err = c.Save(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// Add (Gob-serialized) cache items from an io.Reader, excluding any items with
// keys that already exist in the current cache.
func (c *Cache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]*item{}
	err := dec.Decode(&items)
	if err == nil {
		for k, v := range items {
			_, found := c.items[k]
			if !found {
				c.items[k] = v
			}
		}
	}
	return err
}

// Load and add cache items from the given filename, excluding any items with
// keys that already exist in the current cache.
func (c *Cache) LoadFile(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = c.Load(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// Delete all items from the cache.
func (c *Cache) Flush() {
	c.Lock()
	c.items = map[string]*item{}
	c.Unlock()
}

func (c *Cache) LPush(key string, value interface{}) (int64, error) {
	//TODO implement me
	panic("implement me")
}
func (c *Cache) RPop(key string) interface{} {
	//TODO implement me
	panic("implement me")
}
func (c *Cache) XRead(key string, value int64) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}
func (c *Cache) XAdd(key, id string, values []string) (string, error) {
	//TODO implement me
	panic("implement me")
}
func (c *Cache) XDel(key string, value string) (int64, error) {
	//TODO implement me
	panic("implement me")
}
func (c *Cache) GetLock(key string, acquireTimeout, lockTimeOut time.Duration) (string, error) {
	//TODO implement me
	panic("implement me")
}
func (c *Cache) ReleaseLock(key string, value string) bool {
	//TODO implement me
	panic("implement me")
}

func (c *Cache) ZAdd(s string, f float64, i interface{}) interface{} {
	//TODO implement me
	panic("implement me")
}

func (c *Cache) ZRangeByScore(s string, i int64, i2 int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Cache) ZRem(s string, i ...interface{}) error {
	//TODO implement me
	panic("implement me")
}
