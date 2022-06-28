package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/donetkit/contrib/utils/cache"
	"github.com/donetkit/contrib/utils/uuid"
	"github.com/go-redis/redis/v8"
	"time"
)

func (c *Cache) getInstance() *Cache {
	if c.db < 0 || c.db > 15 {
		c.db = 0
	}
	cache := &Cache{
		ctx:    c.config.ctx,
		client: allClient[c.db],
		config: c.config,
	}
	cache.ctx = context.WithValue(cache.ctx, redisClientDBKey, c.db)
	return cache
}

func (c *Cache) WithDB(db int) cache.ICache {
	if db < 0 || db > 15 {
		db = 0
	}
	c.db = db
	return c
}

func (c *Cache) WithContext(ctx context.Context) cache.ICache {
	if ctx != nil {
		c.ctx = ctx
	} else {
		c.ctx = c.config.ctx
	}
	return c
}

func (c *Cache) Get(key string) interface{} {
	instance := c.getInstance()
	data, err := instance.client.Get(instance.ctx, key).Bytes()
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
	instance := c.getInstance()
	data, err := instance.client.Get(instance.ctx, key).Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Cache) Set(key string, val interface{}, timeout time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	instance := c.getInstance()
	return instance.client.Set(instance.ctx, key, data, timeout).Err()
}

//IsExist 判断key是否存在
func (c *Cache) IsExist(key string) bool {
	instance := c.getInstance()
	i := instance.client.Exists(instance.ctx, key).Val()
	return i > 0
}

//Delete 删除
func (c *Cache) Delete(key string) (int64, error) {
	instance := c.getInstance()
	cmd := instance.client.Del(instance.ctx, key)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

// LPush 左进
func (c *Cache) LPush(key string, values interface{}) (int64, error) {
	instance := c.getInstance()
	cmd := instance.client.LPush(instance.ctx, key, values)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

// RPop 右出
func (c *Cache) RPop(key string) interface{} {
	instance := c.getInstance()
	cmd := instance.client.RPop(instance.ctx, key)
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
	instance := c.getInstance()
	msg, err := instance.client.XRead(instance.ctx, &redis.XReadArgs{
		Streams: []string{key, "0"},
		Count:   count,
		Block:   10 * time.Millisecond,
	}).Result()
	//msg, err := instance.client.XReadStreams(instance.ctx, key, fmt.Sprintf("%d", count)).Result()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Cache) XAdd(key, id string, values []string) (string, error) {
	instance := c.getInstance()
	id, err := instance.client.XAdd(instance.ctx, &redis.XAddArgs{
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
	instance := c.getInstance()
	n, err := instance.client.XDel(instance.ctx, key, id).Result()
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
		instance := c.getInstance()
		if success, err := instance.client.SetNX(instance.ctx, lockName, code, lockTimeOut).Result(); err != nil && err != redis.Nil {
			return "", err
		} else if success {
			return code, nil
		} else if instance.client.TTL(instance.ctx, lockName).Val() == -1 {
			instance.client.Expire(instance.ctx, lockName, lockTimeOut)
		}
		time.Sleep(time.Millisecond)
	}
	return "", fmt.Errorf("timeout")
}

func (c *Cache) ReleaseLock(lockName, code string) bool {
	instance := c.getInstance()
	txf := func(tx *redis.Tx) error {
		if v, err := tx.Get(instance.ctx, lockName).Result(); err != nil && err != redis.Nil {
			return err
		} else if v == code {
			_, err := tx.Pipelined(instance.ctx, func(pipe redis.Pipeliner) error {
				//count++
				pipe.Del(instance.ctx, lockName)
				return nil
			})
			return err
		}
		return nil
	}
	for {
		if err := instance.client.Watch(instance.ctx, txf, lockName); err == nil {
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
	instance := c.getInstance()
	cmd := instance.client.IncrBy(instance.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) IncrementFloat(key string, value float64) (float64, error) {
	instance := c.getInstance()
	cmd := instance.client.IncrByFloat(instance.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Decrement(key string, value int64) (int64, error) {
	instance := c.getInstance()
	cmd := instance.client.DecrBy(instance.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Flush() {
	instance := c.getInstance()
	instance.client.FlushAll(c.ctx)
}

func (c *Cache) ZAdd(key string, score float64, member interface{}) interface{} {
	//NX: 添加元素时，如果该元素已经存在，则添加失败。
	//XX: 添加元素时，如果元素存在，执行修改，如果不存在，则失败。
	//CH：修改元素分数，后面可以接多个score member
	//INCR: 和 ZINCRBY 一样的效果，可以指定某一个元素，给它的分数进行加减操作
	instance := c.getInstance()
	cmd := instance.client.ZAdd(instance.ctx, key, &redis.Z{Score: score, Member: member})
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (c *Cache) ZRangeByScore(key string, min, max int64) ([]string, error) {
	instance := c.getInstance()
	cmd := instance.client.ZRangeByScore(instance.ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", min),
		Max: fmt.Sprintf("%d", max),
	})
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) ZRem(key string, members ...interface{}) error {
	instance := c.getInstance()
	cmd := instance.client.ZRem(instance.ctx, key, members)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}
