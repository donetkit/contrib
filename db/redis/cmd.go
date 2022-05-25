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

func (c *Cache) getInstance() *Cache {
	if c.db < 0 || c.db > 15 {
		c.db = 0
	}
	return &Cache{
		ctx:    c.config.ctx,
		client: allClient[c.db],
		config: c.config,
	}
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
	data, err := c.getInstance().client.Get(c.ctx, key).Bytes()
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
	data, err := c.getInstance().client.Get(c.ctx, key).Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Cache) Set(key string, val interface{}, timeout time.Duration) error {
	return c.getInstance().client.Set(c.ctx, key, val, timeout).Err()
}

//IsExist 判断key是否存在
func (c *Cache) IsExist(key string) bool {
	i := c.getInstance().client.Exists(c.ctx, key).Val()
	return i > 0
}

//Delete 删除
func (c *Cache) Delete(key string) (int64, error) {
	cmd := c.getInstance().client.Del(c.ctx, key)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

// LPush 左进
func (c *Cache) LPush(key string, values interface{}) (int64, error) {
	cmd := c.getInstance().client.LPush(c.ctx, key, values)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

// RPop 右出
func (c *Cache) RPop(key string) interface{} {
	cmd := c.getInstance().client.RPop(c.ctx, key)
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
	msg, err := c.getInstance().client.XRead(c.ctx, &redis.XReadArgs{
		Streams: []string{key, "0"},
		Count:   count,
		Block:   10 * time.Millisecond,
	}).Result()
	//msg, err := c.getInstance().client.XReadStreams(c.ctx, key, fmt.Sprintf("%d", count)).Result()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Cache) XAdd(key, id string, values []string) (string, error) {
	id, err := c.getInstance().client.XAdd(c.ctx, &redis.XAddArgs{
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
	n, err := c.getInstance().client.XDel(c.ctx, key, id).Result()
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
		if success, err := c.getInstance().client.SetNX(c.ctx, lockName, code, lockTimeOut).Result(); err != nil && err != redis.Nil {
			return "", err
		} else if success {
			return code, nil
		} else if c.getInstance().client.TTL(c.ctx, lockName).Val() == -1 {
			c.getInstance().client.Expire(c.ctx, lockName, lockTimeOut)
		}
		time.Sleep(time.Millisecond)
	}
	return "", fmt.Errorf("timeout")
}

func (c *Cache) ReleaseLock(lockName, code string) bool {
	txf := func(tx *redis.Tx) error {
		if v, err := tx.Get(c.ctx, lockName).Result(); err != nil && err != redis.Nil {
			return err
		} else if v == code {
			_, err := tx.Pipelined(c.ctx, func(pipe redis.Pipeliner) error {
				//count++
				pipe.Del(c.ctx, lockName)
				return nil
			})
			return err
		}
		return nil
	}
	for {
		if err := c.getInstance().client.Watch(c.ctx, txf, lockName); err == nil {
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
	cmd := c.getInstance().client.IncrBy(c.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) IncrementFloat(key string, value float64) (float64, error) {
	cmd := c.getInstance().client.IncrByFloat(c.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Decrement(key string, value int64) (int64, error) {
	cmd := c.getInstance().client.DecrBy(c.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Flush() {
	c.getInstance().client.FlushAll(c.ctx)
}
