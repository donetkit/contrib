package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/donetkit/contrib/utils/cache"
	"github.com/donetkit/contrib/utils/uuid"
	"github.com/go-redis/redis/v8"
	"reflect"
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
func (c *Cache) XRead(key string, startId string, count int64, block int64) []redis.XMessage {
	// startId 开始编号 特殊的$，表示接收从阻塞那一刻开始添加到流的消息
	if len(startId) == 0 {
		startId = "$"
	}
	instance := c.getInstance()
	arg := &redis.XReadArgs{
		Streams: []string{key, startId},
		Count:   count,
		//Block:   1 * time.Millisecond,
	}
	if block > 0 {
		arg.Block = time.Millisecond * time.Duration(block)
	}
	val := instance.client.XRead(instance.ctx, arg)
	if val.Err() != nil {
		return nil
	}
	var message []redis.XMessage
	for _, stream := range val.Val() {
		message = append(message, stream.Messages...)
	}
	return message

}

func (c *Cache) XAdd(key, msgId string, trim bool, maxLength int64, value interface{}) string {
	instance := c.getInstance()
	val := interfaceToStr(value)
	arg := &redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{key: val},
	}
	if trim {
		arg.MaxLenApprox = maxLength
	}
	if msgId != "" {
		arg.ID = msgId
	}

	id, err := instance.client.XAdd(instance.ctx, arg).Result()
	if err != nil {
		return ""
	}
	return id
}

func (c *Cache) XDel(key string, id ...string) int64 {
	instance := c.getInstance()
	n, err := instance.client.XDel(instance.ctx, key, id...).Result()
	if err != nil {
		return 0
	}
	return n
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

func (c *Cache) XLen(key string) int64 {
	instance := c.getInstance()
	cmd := instance.client.XLen(instance.ctx, key)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) Exists(keys ...string) int64 {
	instance := c.getInstance()
	cmd := instance.client.Exists(instance.ctx, keys...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XInfoGroups(key string) []redis.XInfoGroup {
	instance := c.getInstance()
	cmd := instance.client.XInfoGroups(instance.ctx, key)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XGroupCreateMkStream(key string, group string, start string) string {
	instance := c.getInstance()
	cmd := instance.client.XGroupCreateMkStream(instance.ctx, key, group, start)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) XGroupDestroy(key string, group string) int64 {
	instance := c.getInstance()
	cmd := instance.client.XGroupDestroy(instance.ctx, key, group)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XPendingExt(key string, group string, startId string, endId string, count int64, consumer ...string) []redis.XPendingExt {
	instance := c.getInstance()
	arg := &redis.XPendingExtArgs{
		Stream: key,
		Group:  group,
		Start:  startId,
		End:    endId,
		Count:  count,
		//consumer: "consumer",
	}
	if len(consumer) > 0 {
		arg.Consumer = consumer[0]
	}
	cmd := instance.client.XPendingExt(instance.ctx, arg)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XPending(key string, group string) *redis.XPending {
	instance := c.getInstance()
	cmd := instance.client.XPending(instance.ctx, key, group)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XGroupDelConsumer(key string, group string, consumer string) int64 {
	instance := c.getInstance()
	cmd := instance.client.XGroupDelConsumer(instance.ctx, key, group, consumer)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XGroupSetID(key string, group string, start string) string {
	instance := c.getInstance()
	cmd := instance.client.XGroupSetID(instance.ctx, key, group, start)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) XReadGroup(key string, group string, consumer string, count int64, block int64, id ...string) []redis.XMessage {
	instance := c.getInstance()
	arg := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Count:    count,
		Streams:  []string{key, ">"},
	}

	if block > 0 {
		arg.Block = time.Millisecond * time.Duration(block)
	}

	if len(id) > 0 {
		arg.Streams = []string{key, id[0]}
	}
	cmd := instance.client.XReadGroup(instance.ctx, arg)
	if cmd.Err() != nil {
		return nil
	}
	var message []redis.XMessage
	for _, stream := range cmd.Val() {
		message = append(message, stream.Messages...)
	}
	return message
}

func (c *Cache) XInfoStream(key string) *redis.XInfoStream {
	instance := c.getInstance()

	cmd := instance.client.XInfoStream(instance.ctx, key)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XInfoConsumers(key string, group string) []redis.XInfoConsumer {
	instance := c.getInstance()

	cmd := instance.client.XInfoConsumers(instance.ctx, key, group)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) Pipeline() redis.Pipeliner {
	return c.getInstance().client.Pipeline()
}

func (c *Cache) XClaim(key string, group string, consumer string, id string, msIdle int64) []redis.XMessage {
	instance := c.getInstance()
	arg := &redis.XClaimArgs{
		Stream:   key,
		Group:    group,
		Consumer: consumer,
		MinIdle:  time.Millisecond * time.Duration(msIdle),
		Messages: []string{id},
	}
	cmd := instance.client.XClaim(instance.ctx, arg)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XAck(key string, group string, ids ...string) int64 {
	instance := c.getInstance()
	cmd := instance.client.XAck(instance.ctx, key, group, ids...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XTrimMaxLen(key string, maxLen int64) int64 {
	instance := c.getInstance()
	cmd := instance.client.XTrimMaxLen(instance.ctx, key, maxLen)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XRangeN(key string, start string, stop string, count int64) []redis.XMessage {
	instance := c.getInstance()
	cmd := instance.client.XRangeN(instance.ctx, key, start, stop, count)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XRange(key string, start string, stop string) []redis.XMessage {
	instance := c.getInstance()
	cmd := instance.client.XRange(instance.ctx, key, start, stop)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

// interfaceToStr
func interfaceToStr(obj interface{}) string {
	if str, ok := obj.(string); ok {
		return str
	}
	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.String:
		return obj.(string)
	default:

	}
	str, _ := json.Marshal(obj)
	return string(str)
}
