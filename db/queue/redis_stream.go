package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type RedisStream struct {
	// 失败时抛出异常。默认false
	ThrowOnFailure bool

	//发送消息失败时的重试次数。默认3次
	RetryTimesWhenSendFailed int

	// 重试间隔。默认1000ms
	RetryIntervalWhenSendFailed int

	//消息队列主题
	Topic string

	ctx context.Context

	Key string

	_count int64

	// 重新处理确认队列中死信的间隔。默认60s
	RetryInterval int64

	// 基元类型数据添加该key构成集合。默认__data
	PrimitiveKey string

	// 最大队列长度。要保留的消息个数，超过则移除较老消息，非精确，实际上略大于该值，默认100万
	MaxLength int64

	// 最大重试次数。超过该次数后，消息将被抛弃，默认10次
	MaxRetry int64

	// 异步消费时的阻塞时间。默认15秒
	BlockTime int64

	// 开始编号。独立消费时使用，消费组消费时不使用，默认0-0
	StartId string

	// 消费者组。指定消费组后，不再使用独立消费。通过SetGroup可自动创建消费组
	Group string

	// 消费者
	Consumer string

	client *redis.Client
}

func NewRedisStream(client *redis.Client, key string) *RedisStream {
	return &RedisStream{
		Key:           key,
		RetryInterval: 60,
		PrimitiveKey:  "__data",
		MaxLength:     1_000_000,
		MaxRetry:      10,
		BlockTime:     15,
		StartId:       "0-0",
		client:        client,
		ctx:           context.Background(),

		RetryTimesWhenSendFailed: 3,

		RetryIntervalWhenSendFailed: 1000,
	}
}

// Count 个数
func (r *RedisStream) Count() int64 {
	result, err := r.client.XLen(r.ctx, r.Key).Result()
	if err != nil {
		return 0
	}
	return result

}

// IsEmpty 集合是否为空
func (r *RedisStream) IsEmpty() bool {
	return r.Count() == 0

}

// SetGroup 设置消费组。如果消费组不存在则创建
func (r *RedisStream) SetGroup(group string) bool {
	if len(group) == 0 {
		return false
	}
	r.Group = group

	keyCount, _ := r.client.Exists(r.ctx, r.Key).Result()
	// 如果Stream不存在，则直接创建消费组，此时会创建Stream
	if keyCount > 0 {
		return r.GroupCreate(group)
	}

	groups := r.GetGroups()
	if groups == nil {
		return r.GroupCreate(group)
	}
	groupCreate := false
	for _, g := range groups {
		if g.Name == group {
			groupCreate = true
			break
		}
	}
	if !groupCreate {
		return r.GroupCreate(group)
	}

	return false
}

func (r *RedisStream) GetGroups() []redis.XInfoGroup {
	xInfoGroup, err := r.client.XInfoGroups(r.ctx, r.Key).Result()
	if err != nil {
		return nil
	}
	return xInfoGroup
}

// GroupCreate 创建消费组
//group 消费组名称开始编号。
//startIds 0表示从开头，$表示从末尾，收到下一条生产消息才开始消费 stream不存在，则会报错，所以在后面 加上 MkStream
func (r *RedisStream) GroupCreate(group string, startIds ...string) bool {
	if len(group) == 0 {
		return false
	}
	startId := "0"
	if len(startIds) > 0 {
		startId = startIds[0]
	}

	result, err := r.client.XGroupCreateMkStream(r.ctx, r.Key, group, startId).Result()
	if err != nil {
		return false
	}
	return result == "OK"
}

// GroupDestroy 销毁消费组
// group 消费组名称
func (r *RedisStream) GroupDestroy(group string) int64 {
	result, err := r.client.XGroupDestroy(r.ctx, r.Key, group).Result()
	if err != nil {
		return 0
	}
	return result

}

// Pending 获取等待列表消息
// group 消费组名称
func (r *RedisStream) Pending(group string, startId string, endId string, count ...int64) []redis.XPendingExt {
	if len(group) == 0 {
		return nil
	}
	if len(startId) == 0 {
		startId = "-"
	}
	if len(endId) == 0 {
		endId = "+"
	}
	var pendingCount int64 = 100
	if len(count) > 0 {
		pendingCount = count[0]
	}

	args := &redis.XPendingExtArgs{
		Stream: r.Key,
		Group:  group,
		Start:  startId,
		End:    endId,
		Count:  pendingCount,
		//Consumer: "consumer",
	}
	infoExt, err := r.client.XPendingExt(r.ctx, args).Result()
	if err != nil {
		return nil
	}
	return infoExt
}

// GetPending 获取等待列表
// group 消费组名称信息
func (r *RedisStream) GetPending(group string) *redis.XPending {
	if len(group) == 0 {
		return nil
	}
	result, err := r.client.XPending(r.ctx, r.Key, group).Result()
	if err != nil {
		return nil
	}
	return result
}

// GroupDeleteConsumer 销毁消费者
// group 消费组名称
// consumer 消费者
// returns 返回消费者在被删除之前所拥有的待处理消息数量
func (r *RedisStream) GroupDeleteConsumer(group string, consumer string) int64 {
	if len(group) == 0 {
		return 0
	}
	if len(consumer) == 0 {
		return 0
	}
	result, err := r.client.XGroupDelConsumer(r.ctx, r.Key, group, consumer).Result()
	if err != nil {
		return 0
	}
	return result
}

// GroupSetId 设置消费组Id
// group 消费组名称
// startId 开始编号
func (r *RedisStream) GroupSetId(group string, startId string) bool {
	if len(group) == 0 {
		return false
	}
	if len(startId) == 0 {
		startId = "$"
	}
	result, err := r.client.XGroupSetID(r.ctx, r.Key, group, startId).Result()
	if err != nil {
		return false
	}
	return result == "OK"
}

// ReadGroup 消费组消费  group 消费组名称  consumer 消费组  count 消息个数
func (r *RedisStream) ReadGroup(group string, consumer string, count ...int64) []redis.XStream {
	if len(group) == 0 {
		return nil
	}

	arg := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{r.Key, ">"},
	}
	if len(count) > 0 {
		arg.Count = count[0]
	}
	result, err := r.client.XReadGroup(r.ctx, arg).Result()
	if err != nil {
		return nil
	}
	return result
}

// ReadGroupBlock 消费组消费
// group 消费组
// consumer 消费组
// count 消息个数
// block 阻塞毫秒数，0表示永远
// id 消息id
func (r *RedisStream) ReadGroupBlock(group string, consumer string, count int64, block int64, id ...string) []redis.XStream {
	if len(group) == 0 {
		return nil
	}
	if len(consumer) == 0 {
		return nil
	}
	arg := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Count:    count,
		Block:    time.Millisecond * time.Duration(block),
		Streams:  []string{r.Key, ">"},
	}

	if len(id) > 0 {
		arg.Streams = []string{r.Key, id[0]}
	}

	result, err := r.client.XReadGroup(r.ctx, arg).Result()
	if err != nil {
		return nil
	}
	return result
}

// GetInfo 队列信息
func (r *RedisStream) GetInfo() *redis.XInfoStream {
	result, err := r.client.XInfoStream(r.ctx, r.Key).Result()
	if err != nil {
		return nil
	}
	return result
}

// GetConsumers 获取消费者
// group 消费组
func (r *RedisStream) GetConsumers(group string) []redis.XInfoConsumer {
	if len(group) == 0 {
		return nil
	}
	result, err := r.client.XInfoConsumers(r.ctx, r.Key, group).Result()
	if err != nil {
		return nil
	}
	return result
}

// Acknowledge 消费确认
func (r *RedisStream) Acknowledge(keys ...string) int64 {
	var rs int64
	for _, key := range keys {
		rs += r.Ack(r.Group, key)
	}
	return rs
}

// Add 生产添加
func (r *RedisStream) Add(value interface{}, msgId ...string) string {
	if value == nil {
		return "" //, errors.New("argument null exception error: value is null")
	}

	// 自动修剪超长部分，每1000次生产，修剪一次
	if r._count <= 0 {
		r._count = r.Count()
	}
	atomic.AddInt64(&r._count, 1)

	var trim = false
	if r.MaxLength > 0 && r._count%1000 == 0 {
		r._count = r.Count() + 1

		trim = true
	}
	var id = ""
	if len(msgId) > 0 {
		id = msgId[0]
	}
	return r.AddInternal(value, id, trim, true)
}

func (r *RedisStream) AddInternal(value interface{}, msgId string, trim bool, retryOnFailed bool) string {
	args := &redis.XAddArgs{
		Stream: r.Key,
		Values: map[string]interface{}{r.Key: value},
	}

	if trim {
		args.MaxLenApprox = r.MaxLength
	}
	if msgId != "" {
		args.ID = msgId
	}
	for i := 0; i < r.RetryTimesWhenSendFailed; i++ {
		var id, err = r.client.XAdd(r.ctx, args).Result()
		if !retryOnFailed || err != redis.Nil {
			return id
		}
		if i < r.RetryTimesWhenSendFailed {
			time.Sleep(time.Second * time.Duration(r.RetryIntervalWhenSendFailed))
		}
	}

	return ""
}

// Adds 批量生产添加
func (r *RedisStream) Adds(values []interface{}) int {
	if len(values) == 0 {
		return 0
	}
	// 量少时直接插入，而不用管道
	if len(values) <= 2 {
		for _, value := range values {
			r.Add(value)
		}
		return len(values)
	}

	// 自动修剪超长部分，每1000次生产，修剪一次
	if r._count <= 0 {
		r._count = r.Count()
	}

	// 自动修剪超长部分，每1000次生产，修剪一次
	if r._count <= 0 {
		r._count = r.Count()
	}

	var trim = false
	if r.MaxLength > 0 && r._count > 0 && r._count%1000 == 0 {
		r._count = r.Count() + 1
		trim = true
	}
	// 开启管道
	pipe := r.client.Pipeline()
	for _, item := range values {
		atomic.AddInt64(&r._count, 1)
		r.AddInternal(item, "", trim, false)
		trim = false
	}
	cmder, _ := pipe.Exec(r.ctx)

	perm, err := cmder[0].(*redis.StringCmd).Result()
	if err == redis.Nil {

	}
	fmt.Println("从redis中获取到的值：", perm)

	return len(values)

}

// Take 批量消费获取，前移指针StartId
func (r *RedisStream) Take(count int64) []redis.XMessage {
	var message []redis.XMessage
	var group = r.Group
	var rs []redis.XStream
	if !(len(group) == 0) {
		r.RetryAck()
		rs = r.ReadGroup(group, r.Consumer, count)
	} else {
		rs = r.Read(r.StartId, count)
	}

	for _, item := range rs {
		if len(group) == 0 {
			for _, message := range item.Messages {
				r.SetNextId(message.ID)
			}
		}
		message = append(message, item.Messages...)
	}
	return message
}

// TakeOne 消费获取一个
func (r *RedisStream) TakeOne() []redis.XMessage {
	return r.Take(1)
}

// TakeOneAsync 异步消费获取一个
func (r *RedisStream) TakeOneAsync(ctx context.Context, timeout int64) []redis.XMessage {
	return r.TakeMessageAsync(timeout)
}

func (r *RedisStream) SetNextId(id string) {
	r.StartId = id
}

var _nextRetry time.Time

// RetryAck 处理未确认的死信，重新放入队列
func (r *RedisStream) RetryAck() {

	var now = time.Now()
	// 一定间隔处理当前key死信
	if _nextRetry.UnixMilli() < now.UnixMilli() {
		_nextRetry = now.Add(time.Duration(r.RetryInterval) * time.Second)
		var retry = time.Duration(r.RetryInterval*1000) * time.Millisecond
		// 拿到死信，重新放入队列
		id := ""
		for {
			var listXPendingExt = r.Pending(r.Group, id, "", 100)
			if len(listXPendingExt) == 0 {
				break
			}
			for _, xPendingExt := range listXPendingExt {
				if xPendingExt.Idle >= retry {
					if xPendingExt.RetryCount > r.MaxRetry {
						str, _ := json.Marshal(xPendingExt)
						fmt.Println(fmt.Sprintf("%s 删除多次失败死信：%s", r.Group, str))
						//Delete(item.Id);
						r.Claim(r.Group, r.Consumer, xPendingExt.ID, r.RetryInterval*1000)
						r.Ack(r.Group, xPendingExt.ID)

					} else {
						str, _ := json.Marshal(xPendingExt)
						fmt.Println(fmt.Sprintf("%s 定时回滚：%s", r.Group, str))
						r.Claim(r.Group, r.Consumer, xPendingExt.ID, r.RetryInterval*1000)
					}
				}

			}

			// 下一个开始id
			id = listXPendingExt[len(listXPendingExt)-1].ID
			var p = strings.Index(id, "-")
			if p > 0 {
				nid, err := strconv.ParseInt(id[:p], 10, 64)
				if err == nil {
					id = fmt.Sprintf("%d-0", nid+1)
				}
			}

			// 清理历史消费者
			consumers := r.GetConsumers(r.Group)
			for _, item := range consumers {
				if item.Pending == 0 && item.Idle > 3600_000 {
					str, _ := json.Marshal(item)
					fmt.Println(fmt.Sprintf("%s s删除空闲消费者：%s", r.Group, str))
					r.GroupDeleteConsumer(r.Group, item.Name)
				}
			}

		}
	}

}

// Read 原始独立消费
// startId 开始编号 特殊的$，表示接收从阻塞那一刻开始添加到流的消息
// count 消息个数
func (r *RedisStream) Read(startId string, count int64) []redis.XStream {
	if len(startId) == 0 {
		startId = "$"
	}
	arg := &redis.XReadArgs{
		Streams: []string{r.Key, startId},
		Count:   count,
		//Block:   1 * time.Millisecond,
	}
	result, err := r.client.XRead(r.ctx, arg).Result()
	if err != nil {
		return nil
	}
	return result
}

// ReadAsync 原始独立消费
// startId 开始编号 特殊的$，表示接收从阻塞那一刻开始添加到流的消息
// count 消息个数
func (r *RedisStream) ReadAsync(startId string, count int64, block int64) []redis.XStream {
	if len(startId) == 0 {
		startId = "$"
	}
	arg := &redis.XReadArgs{
		Streams: []string{r.Key, startId},
		Count:   count,
		//Block:   1 * time.Millisecond,
	}
	if block > 0 {
		arg.Block = time.Duration(block) * time.Millisecond
	}
	result, err := r.client.XRead(r.ctx, arg).Result()
	if err != nil {
		return nil
	}
	return result
}

// Claim 改变待处理消息的所有权，抢夺他人未确认消息
// group 消费组名称
// consumer 目标消费者
// id 消息Id
// msIdle 空闲时间。默认3600_000
func (r *RedisStream) Claim(group string, consumer string, id string, msIdle int64) []redis.XMessage {
	if len(group) == 0 || len(consumer) == 0 {
		return nil
	}
	arg := &redis.XClaimArgs{
		Stream:   r.Key,
		Group:    group,
		Consumer: consumer,
		MinIdle:  time.Millisecond * time.Duration(msIdle),
		Messages: []string{id},
	}
	result, err := r.client.XClaim(r.ctx, arg).Result()
	if err != nil {
		return nil
	}
	return result

}

// Ack 确认消息
// group 消费组名称
// id 消息Id
func (r *RedisStream) Ack(group string, id string) int64 {
	if len(group) == 0 {
		return 0
	}
	if len(id) == 0 {
		return 0
	}
	result, err := r.client.XAck(r.ctx, r.Key, group, id).Result()
	if err != nil {
		return 0
	}
	return result
}

// TakeMessageAsync 异步消费获取一个 timeout 超时时间，默认0秒永远阻塞
func (r *RedisStream) TakeMessageAsync(timeout int64) []redis.XMessage {
	var group = r.Group
	if len(group) == 0 {
		r.RetryAck()
	}
	var rs []redis.XStream
	var t = timeout * 1000
	if len(group) > 0 {
		rs = r.ReadGroupBlock(group, r.Consumer, 1, t, ">")
	} else {
		rs = r.ReadAsync(r.StartId, 1, t)
	}

	if len(rs) == 0 {
		rs = r.ReadGroupBlock(group, r.Consumer, 1, 3_000, "0")
		if len(rs) == 0 {
			return nil
		}

		fmt.Println(fmt.Sprintf("%s处理历史：%s", r.Group, rs[0].Messages[0].ID))
		return rs[0].Messages
	}
	return nil
}

// ConsumeAsync 队列消费大循环，处理消息后自动确认
func (r *RedisStream) ConsumeAsync(ctx context.Context, OnMessage func(msg []redis.XMessage)) {
	// 自动创建消费组
	r.SetGroup(r.Group)
	// 主题
	//var topic = r.Key
	// 超时时间，用于阻塞等待
	var timeout = r.BlockTime

	for {
		select {
		case <-ctx.Done():
			return // 退出了...
		}
		// 异步阻塞消费
		mqMsg := r.TakeMessageAsync(timeout)
		if len(mqMsg) == 0 {
			// 没有消息，歇一会
			time.Sleep(time.Millisecond * 1000)
			continue
		}
		// 处理消息
		OnMessage(mqMsg)
		// 确认消息
		for _, msg := range mqMsg {
			r.Acknowledge(msg.ID)
		}
	}

}

// Delete 删除指定消息
// id 消息Id
func (r *RedisStream) Delete(id ...string) int64 {
	result, err := r.client.XDel(r.ctx, r.Key, id...).Result()
	if err != nil {
		return 0
	}
	return result
}

// Trim 裁剪队列到指定大小
// maxLen 最大长度。为了提高效率，最大长度并没有那么精准
func (r *RedisStream) Trim(maxLen int64) int64 {
	result, err := r.client.XTrimMaxLen(r.ctx, r.Key, maxLen).Result()
	if err != nil {
		return 0
	}
	return result
}

// Range  获取区间消息
func (r *RedisStream) Range(startId string, endId string, count ...int64) []redis.XMessage {
	if len(startId) == 0 {
		startId = "-"
	}
	if len(endId) == 0 {
		endId = "+"
	}

	if len(count) > 0 {
		result, err := r.client.XRangeN(r.ctx, r.Key, startId, endId, count[0]).Result()
		if err != nil {
			return nil
		}
		return result

	} else {
		result, err := r.client.XRange(r.ctx, r.Key, startId, endId).Result()
		if err != nil {
			return nil
		}
		return result
	}
	return nil
}

// RangeTimeSpan 获取区间消息
func (r *RedisStream) RangeTimeSpan(start int64, end int64, count ...int64) []redis.XMessage {
	return r.Range(fmt.Sprintf("%d-0", start), fmt.Sprintf("%d-0", end), count...)
}
