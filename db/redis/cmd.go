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

func (c *Cache) WithDB(db int) cache.ICache {
	if db < 0 || db > 15 {
		db = 0
	}
	cache := &Cache{
		db:     db,
		ctx:    c.ctx,
		client: allClient[db],
		config: c.config,
	}
	cache.ctx = context.WithValue(cache.ctx, redisClientDBKey, db)
	return cache
}

func (c *Cache) WithContext(ctx context.Context) cache.ICache {
	if ctx != nil {
		c.ctx = ctx
	} else {
		c.ctx = c.config.ctx
	}
	c.ctx = context.WithValue(c.ctx, redisClientDBKey, c.db)
	return c
}

func (c *Cache) Get(key string) interface{} {

	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var reply interface{}
	if err = Unmarshal(data, &reply); err != nil {
		return nil
	}
	return reply
}

func (c *Cache) GetString(key string) (string, error) {

	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Cache) Set(key string, value interface{}, timeout time.Duration) error {
	data, err := Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(c.ctx, key, data, timeout).Err()
}

func (c *Cache) SetEX(key string, value interface{}, timeout time.Duration) error {
	data, err := Marshal(value)
	if err != nil {
		return err
	}
	return c.client.SetEX(c.ctx, key, data, timeout).Err()
}

// IsExist 判断key是否存在
func (c *Cache) IsExist(key string) bool {

	i := c.client.Exists(c.ctx, key).Val()
	return i > 0
}

// Delete 删除
func (c *Cache) Delete(key ...string) int64 {

	cmd := c.client.Del(c.ctx, key...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

// LPush 左进
func (c *Cache) LPush(key string, values ...interface{}) int64 {

	cmd := c.client.LPush(c.ctx, key, values...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

// RPop 右出
func (c *Cache) RPop(key string) string {
	cmd := c.client.RPop(c.ctx, key)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) BRPopLPush(source string, destination string, timeout time.Duration) string {
	cmd := c.client.BRPopLPush(c.ctx, source, destination, timeout)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) RPopLPush(source string, destination string) string {
	cmd := c.client.RPopLPush(c.ctx, source, destination)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) LRem(key string, count int64, value interface{}) int64 {
	cmd := c.client.LRem(c.ctx, key, count, value)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) Scan(cursor uint64, match string, count int64) ([]string, uint64) {

	cmd := c.client.Scan(c.ctx, cursor, match, count)
	if cmd.Err() != nil {
		return nil, 0
	}
	return cmd.Val()
}

func (c *Cache) SetNX(key string, value interface{}, expiration time.Duration) bool {
	data, err := Marshal(value)
	if err != nil {
		return false
	}
	cmd := c.client.SetNX(c.ctx, key, data, expiration)
	if cmd.Err() != nil {
		return false
	}
	return cmd.Val()
}

func (c *Cache) LRange(key string, start int64, stop int64) []string {

	cmd := c.client.LRange(c.ctx, key, start, stop)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

// XRead default type []redis.XStream
func (c *Cache) XRead(key string, startId string, count int64, block int64) []redis.XMessage {
	// startId 开始编号 特殊的$，表示接收从阻塞那一刻开始添加到流的消息
	if len(startId) == 0 {
		startId = "$"
	}

	arg := &redis.XReadArgs{
		Streams: []string{key, startId},
		Count:   count,
		//Block:   1 * time.Millisecond,
	}
	if block > 0 {
		arg.Block = time.Millisecond * time.Duration(block)
	}
	val := c.client.XRead(c.ctx, arg)
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

	id, err := c.client.XAdd(c.ctx, arg).Result()
	if err != nil {
		return ""
	}
	return id
}

func (c *Cache) XAddKey(key, msgId string, trim bool, maxLength int64, vKey string, value interface{}) string {
	val := interfaceToStr(value)
	arg := &redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{vKey: val},
	}
	if trim {
		arg.MaxLenApprox = maxLength
	}
	if msgId != "" {
		arg.ID = msgId
	}

	id, err := c.client.XAdd(c.ctx, arg).Result()
	if err != nil {
		return ""
	}
	return id
}

func (c *Cache) XDel(key string, id ...string) int64 {

	n, err := c.client.XDel(c.ctx, key, id...).Result()
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

		if success, err := c.client.SetNX(c.ctx, lockName, code, lockTimeOut).Result(); err != nil && err != redis.Nil {
			return "", err
		} else if success {
			return code, nil
		} else if c.client.TTL(c.ctx, lockName).Val() == -1 {
			c.client.Expire(c.ctx, lockName, lockTimeOut)
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
		if err := c.client.Watch(c.ctx, txf, lockName); err == nil {
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

	cmd := c.client.IncrBy(c.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) IncrementFloat(key string, value float64) (float64, error) {

	cmd := c.client.IncrByFloat(c.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Decrement(key string, value int64) (int64, error) {

	cmd := c.client.DecrBy(c.ctx, key, value)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *Cache) Flush() {

	c.client.FlushAll(c.ctx)
}

func (c *Cache) XLen(key string) int64 {

	cmd := c.client.XLen(c.ctx, key)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) Exists(keys ...string) int64 {

	cmd := c.client.Exists(c.ctx, keys...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XInfoGroups(key string) []redis.XInfoGroup {

	cmd := c.client.XInfoGroups(c.ctx, key)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XGroupCreateMkStream(key string, group string, start string) string {

	cmd := c.client.XGroupCreateMkStream(c.ctx, key, group, start)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) XGroupDestroy(key string, group string) int64 {

	cmd := c.client.XGroupDestroy(c.ctx, key, group)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XPendingExt(key string, group string, startId string, endId string, count int64, consumer ...string) []redis.XPendingExt {

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
	cmd := c.client.XPendingExt(c.ctx, arg)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XPending(key string, group string) *redis.XPending {

	cmd := c.client.XPending(c.ctx, key, group)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XGroupDelConsumer(key string, group string, consumer string) int64 {

	cmd := c.client.XGroupDelConsumer(c.ctx, key, group, consumer)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XGroupSetID(key string, group string, start string) string {

	cmd := c.client.XGroupSetID(c.ctx, key, group, start)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) XReadGroup(key string, group string, consumer string, count int64, block int64, id ...string) []redis.XMessage {

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
	cmd := c.client.XReadGroup(c.ctx, arg)
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

	cmd := c.client.XInfoStream(c.ctx, key)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XInfoConsumers(key string, group string) []redis.XInfoConsumer {

	cmd := c.client.XInfoConsumers(c.ctx, key, group)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

func (c *Cache) XClaim(key string, group string, consumer string, id string, msIdle int64) []redis.XMessage {

	arg := &redis.XClaimArgs{
		Stream:   key,
		Group:    group,
		Consumer: consumer,
		MinIdle:  time.Millisecond * time.Duration(msIdle),
		Messages: []string{id},
	}
	cmd := c.client.XClaim(c.ctx, arg)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XAck(key string, group string, ids ...string) int64 {

	cmd := c.client.XAck(c.ctx, key, group, ids...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XTrimMaxLen(key string, maxLen int64) int64 {

	cmd := c.client.XTrimMaxLen(c.ctx, key, maxLen)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) XRangeN(key string, start string, stop string, count int64) []redis.XMessage {

	cmd := c.client.XRangeN(c.ctx, key, start, stop, count)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) XRange(key string, start string, stop string) []redis.XMessage {

	cmd := c.client.XRange(c.ctx, key, start, stop)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) ZAdd(key string, score float64, value ...interface{}) int64 {
	if len(value) <= 0 {
		return 0
	}

	var member []*redis.Z
	for _, val := range value {
		member = append(member, &redis.Z{Score: score, Member: val})
	}

	cmd := c.client.ZAdd(c.ctx, key, member...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) ZRem(key string, value ...interface{}) int64 {

	cmd := c.client.ZRem(c.ctx, key, value...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) ZRangeByScore(key string, min int64, max int64, offset int64, count int64) []string {

	cmd := c.client.ZRangeByScore(c.ctx, key, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%d", min),
		Max:    fmt.Sprintf("%d", max),
		Offset: offset,
		Count:  count,
	})
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) HashGet(key, value string) string {
	cmd := c.client.HGet(c.ctx, key, value)
	if cmd.Err() != nil {
		return ""
	}
	return cmd.Val()
}

func (c *Cache) HashGets(key string, value ...string) []interface{} {
	cmd := c.client.HMGet(c.ctx, key, value...)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) HashAll(key string) map[string]string {
	cmd := c.client.HGetAll(c.ctx, key)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) HashSet(key string, values ...interface{}) int64 {
	cmd := c.client.HSet(c.ctx, key, values...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) HashExist(key, values string) bool {
	cmd := c.client.HExists(c.ctx, key, values)
	if cmd.Err() != nil {
		return false
	}
	return cmd.Val()
}

func (c *Cache) HashDel(key string, values ...string) int64 {
	cmd := c.client.HDel(c.ctx, key, values...)
	if cmd.Err() != nil {
		return 0
	}
	return cmd.Val()
}

func (c *Cache) HashKeys(key string) []string {
	cmd := c.client.HKeys(c.ctx, key)
	if cmd.Err() != nil {
		return nil
	}
	return cmd.Val()
}

func (c *Cache) HashLen(key string) int64 {
	cmd := c.client.HLen(c.ctx, key)
	if cmd.Err() != nil {
		return 0
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
