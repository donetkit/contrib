package memory

import (
	"context"
	"encoding/binary"
	icache "github.com/donetkit/contrib/utils/cache"
	"hash/fnv"
	"runtime"
	"time"
)

func (j *janitor) Run(c *Cache) {
	j.stop = make(chan bool)
	tick := time.Tick(j.Interval)
	for {
		select {
		case <-tick:
			c.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}

func runJanitor(c *Cache, ci time.Duration) {
	j := &janitor{
		Interval: ci,
	}
	c.janitor = j
	go j.Run(c)
}

func newCache(de time.Duration) *Cache {
	if de == 0 {
		de = -1
	}
	c := &Cache{
		defaultExpiration: de,
		items:             map[string]*item{},
	}
	return c
}

// Return a new cache with a given default expiration duration and cleanup
// interval. If the expiration duration is less than 1, the items in the cache
// never expire (by default), and must be deleted manually. If the cleanup
// interval is less than one, expired items are not deleted from the cache
// before their next lookup or before calling DeleteExpired.
func New(opts ...Option) *Cache {
	cfg := &config{
		ctx: context.TODO(),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	c := newCache(cfg.defaultExpiration)
	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	if cfg.cleanupInterval > 0 {
		runJanitor(c, cfg.cleanupInterval)
		runtime.SetFinalizer(c, stopJanitor)
	}
	return c
}

type ShardedCache struct {
	m       uint32
	cs      []*Cache
	janitor *shardedJanitor
}

func (sc *ShardedCache) bucket(k string) *Cache {
	h := fnv.New32()
	h.Write([]byte(k))
	n := binary.BigEndian.Uint32(h.Sum(nil))
	return sc.cs[n%sc.m]
}

func (sc *ShardedCache) WithDB(db int) icache.ICache {
	return sc.bucket("").WithDB(db)
}
func (sc *ShardedCache) WithContext(ctx context.Context) icache.ICache {
	return sc.bucket("").WithContext(ctx)
}
func (sc *ShardedCache) Get(key string) interface{} {
	return sc.bucket(key).Get(key)
}
func (sc *ShardedCache) GetString(key string) (string, error) {
	return sc.bucket(key).GetString(key)
}
func (sc *ShardedCache) Set(key string, val interface{}, timeout time.Duration) error {
	return sc.bucket(key).Set(key, val, timeout)
}
func (sc *ShardedCache) IsExist(key string) bool {
	return sc.bucket(key).IsExist(key)
}
func (sc *ShardedCache) Delete(key string) (int64, error) {
	return sc.bucket(key).Delete(key)
}

func (sc *ShardedCache) Increment(key string, val int64) (int64, error) {
	return sc.bucket(key).Increment(key, val)
}
func (sc *ShardedCache) IncrementFloat(key string, val float64) (float64, error) {
	return sc.bucket(key).IncrementFloat(key, val)
}
func (sc *ShardedCache) Decrement(key string, val int64) (int64, error) {
	return sc.bucket(key).Increment(key, val)
}

func (sc *ShardedCache) DeleteExpired() {
	for _, v := range sc.cs {
		v.DeleteExpired()
	}
}

func (sc *ShardedCache) Flush() {
	for _, v := range sc.cs {
		v.Flush()
	}
}

type shardedJanitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *shardedJanitor) Run(sc *ShardedCache) {
	j.stop = make(chan bool)
	tick := time.Tick(j.Interval)
	for {
		select {
		case <-tick:
			sc.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

func stopShardedJanitor(sc *ShardedCache) {
	sc.janitor.stop <- true
}

func runShardedJanitor(sc *ShardedCache, ci time.Duration) {
	j := &shardedJanitor{
		Interval: ci,
	}
	sc.janitor = j
	go j.Run(sc)
}

func newShardedCache(n int, de time.Duration) *ShardedCache {
	sc := &ShardedCache{
		m:  uint32(n - 1),
		cs: make([]*Cache, n),
	}
	for i := 0; i < n; i++ {
		c := &Cache{
			defaultExpiration: de,
			items:             map[string]*item{},
		}
		sc.cs[i] = c
	}
	return sc
}

func unexportedNewSharded(shards int, defaultExpiration, cleanupInterval time.Duration) *ShardedCache {
	if defaultExpiration == 0 {
		defaultExpiration = -1
	}
	sc := newShardedCache(shards, defaultExpiration)
	if cleanupInterval > 0 {
		runShardedJanitor(sc, cleanupInterval)
		runtime.SetFinalizer(sc, stopShardedJanitor)
	}
	return sc
}
