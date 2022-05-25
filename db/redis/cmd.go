package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/donetkit/contrib/utils/cache"
	"github.com/donetkit/contrib/utils/uuid"
	//"github.com/donetkit/contrib/utils/uuid"
	"github.com/go-redis/redis/v8"
	"time"
)

func (c *Cache) WithDB(db int) cache.ICache {
	if db < 0 || db > 15 {
		db = 0
	}
	return &Cache{
		ctxCache: c.config.ctx,
		client:   allClient[db],
		config:   c.config,
	}
}

func (c *Cache) WithContext(ctx context.Context) cache.ICache {
	if ctx != nil {
		c.ctxCache = ctx
	} else {
		c.ctxCache = c.config.ctx
	}
	return c
}

func (c *Cache) Get(key string) interface{} {
	data, err := c.client.Get(c.ctxCache, key).Bytes()
	if err != nil {
		return nil
	}
	var reply interface{}
	if err = json.Unmarshal(data, &reply); err != nil {
		return nil
	}
	return reply
}

func (c *Cache) GetString(key string) (string, error) {
	data, err := c.client.Get(c.ctxCache, key).Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Cache) Set(key string, val interface{}, timeout time.Duration) error {
	return c.client.Set(c.ctxCache, key, val, timeout).Err()
}

//IsExist 判断key是否存在
func (c *Cache) IsExist(key string) bool {
	i := c.client.Exists(c.ctxCache, key).Val()
	return i > 0
}

//Delete 删除
func (c *Cache) Delete(key string) (int64, error) {
	cmd := c.client.Del(c.ctxCache, key)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

// LPush 左进
func (c *Cache) LPush(key string, values interface{}) (int64, error) {
	cmd := c.client.LPush(c.ctxCache, key, values)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

// RPop 右出
func (c *Cache) RPop(key string) interface{} {
	cmd := c.client.RPop(c.ctxCache, key)
	if cmd.Err() != nil {
		return nil
	}
	var reply interface{}
	if err := json.Unmarshal([]byte(cmd.Val()), &reply); err != nil {
		return nil
	}
	return reply
}

// XRead default type []redis.XStream
func (c *Cache) XRead(key string, count int64) (interface{}, error) {
	if count <= 0 {
		count = 10
	}
	msg, err := c.client.XRead(c.ctxCache, &redis.XReadArgs{
		Streams: []string{key, "0"},
		Count:   count,
		Block:   10 * time.Millisecond,
	}).Result()
	//msg, err := c.client.XReadStreams(c.ctxCache, key, fmt.Sprintf("%d", count)).Result()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Cache) XAdd(key, id string, values []string) (string, error) {
	id, err := c.client.XAdd(c.ctxCache, &redis.XAddArgs{
		Stream: key,
		ID:     id,
		Values: values,
	}).Result()
	if err != nil {
		return "", err
	}
	return id, nil
}

func (c *Cache) XDel(key string, id string) (int64, error) {
	n, err := c.client.XDel(c.ctxCache, key, id).Result()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c *Cache) GetLock(lockName string, acquireTimeout, lockTimeOut time.Duration) (string, error) {
	code := uuid.NewUUID()
	//endTime := util.FwTimer.CalcMillis(time.Now().Add(acquireTimeout))
	endTime := time.Now().Add(acquireTimeout).UnixNano()
	//for util.FwTimer.CalcMillis(time.Now()) <= endTime {
	for time.Now().UnixNano() <= endTime {
		if success, err := c.client.SetNX(c.ctxCache, lockName, code, lockTimeOut).Result(); err != nil && err != redis.Nil {
			return "", err
		} else if success {
			return code, nil
		} else if c.client.TTL(c.ctxCache, lockName).Val() == -1 {
			c.client.Expire(c.ctxCache, lockName, lockTimeOut)
		}
		time.Sleep(time.Millisecond)
	}
	return "", fmt.Errorf("timeout")
}

func (c *Cache) ReleaseLock(lockName, code string) bool {
	txf := func(tx *redis.Tx) error {
		if v, err := tx.Get(c.ctxCache, lockName).Result(); err != nil && err != redis.Nil {
			return err
		} else if v == code {
			_, err := tx.Pipelined(c.ctxCache, func(pipe redis.Pipeliner) error {
				//count++
				pipe.Del(c.ctxCache, lockName)
				return nil
			})
			return err
		}
		return nil
	}
	for {
		if err := c.client.Watch(c.ctxCache, txf, lockName); err == nil {
			return true
		} else if err == redis.TxFailedErr {
			c.config.logger.Errorf("watch key is modified, retry to release lock. err: %s", err.Error())
		} else {
			c.config.logger.Errorf("err: %s", err.Error())
			return false
		}
	}
}

func (c *Cache) Increment(key string, value int64) (int64, error) {
	cmd := c.client.IncrBy(c.ctxCache, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) IncrementFloat(key string, value float64) (float64, error) {
	cmd := c.client.IncrByFloat(c.ctxCache, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Decrement(key string, value int64) (int64, error) {
	cmd := c.client.DecrBy(c.ctxCache, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Flush() {
	c.client.FlushAll(c.ctxCache)
}
