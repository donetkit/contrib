package queue_reliable

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/db/queue/queue_delay"
	"github.com/donetkit/contrib/utils/cache"
	"github.com/donetkit/contrib/utils/gjson"
	"github.com/donetkit/contrib/utils/grand"
	chost "github.com/donetkit/contrib/utils/host"
	"github.com/shirou/gopsutil/v3/host"
	"os"
	"strings"
	"time"
)

var _Key string
var _StatusKey string
var _Status RedisQueueStatus

type RedisQueueStatus struct {
	Key         string // 标识消费者的唯一Key
	MachineName string // 机器名
	UserName    string // 用户名
	ProcessId   int    // 进程Id
	Ip          string // Ip地址
	CreateTime  int64  // 开始时间
	LastActive  int64  // 最后活跃时间
	Consumes    int64  // 消费消息数
	Acks        int64  // 确认消息数
}

type RedisReliableQueue struct {
	key                         string
	DB                          int
	ThrowOnFailure              bool             // 失败时抛出异常。默认false
	RetryTimesWhenSendFailed    int              //发送消息失败时的重试次数。默认3次
	RetryIntervalWhenSendFailed int              // 重试间隔。默认1000ms
	AckKey                      string           // 用于确认的列表
	RetryInterval               int64            // 重新处理确认队列中死信的间隔。默认60s
	MinPipeline                 int64            // 最小管道阈值，达到该值时使用管道，默认3
	count                       int64            // 个数
	IsEmpty                     bool             // 是否为空
	Status                      RedisQueueStatus // 消费状态
	logger                      glog.ILoggerEntry
	l                           glog.ILogger
	client                      cache.ICache    // Client
	ctx                         context.Context // Context
}

func CreateStatus() RedisQueueStatus {
	info, _ := host.Info()
	return RedisQueueStatus{
		Key:         grand.RandAllString(8),
		MachineName: info.Hostname,
		UserName:    "",
		ProcessId:   os.Getpid(),
		Ip:          chost.GetOutBoundIp(),
		CreateTime:  time.Now().UnixMilli(),
		LastActive:  time.Now().UnixMilli(),
	}
}

func New(client cache.ICache, key string, logger glog.ILogger) *RedisReliableQueue {
	_Key = key
	_Status = CreateStatus()
	_StatusKey = fmt.Sprintf("%s:Status:%s", key, _Status.Key)
	return &RedisReliableQueue{
		key:                         key,
		RetryTimesWhenSendFailed:    3,
		RetryIntervalWhenSendFailed: 1000,
		RetryInterval:               60,
		MinPipeline:                 3,
		logger:                      logger.WithField("mq_redis_reliable", "mq_redis_reliable"),
		l:                           logger,
		AckKey:                      fmt.Sprintf("%s:Ack:%s", key, _Status.Key),
		client:                      client,
		ctx:                         context.Background(),
	}
}

// Add 批量生产添加
func (r *RedisReliableQueue) Add(values ...interface{}) int64 {
	if values == nil || len(values) == 0 {
		return 0
	}

	var rs int64
	for i := 0; i < r.RetryTimesWhenSendFailed; i++ {
		// 返回插入后的LIST长度。Redis执行命令不会失败，因此正常插入不应该返回0，如果返回了0或者空，可能是中间代理出了问题
		rs = r.client.WithDB(r.DB).WithContext(r.ctx).LPush(r.key, values...)
		if rs > 0 {
			return rs
		}
		r.logger.Debug(fmt.Sprintf("发布到队列[%s]失败！", r.key))

		if i < r.RetryTimesWhenSendFailed {
			time.Sleep(time.Millisecond * time.Duration(r.RetryIntervalWhenSendFailed))
		}
	}
	return rs

}

// TakeOne 消费获取，从Key弹出并备份到AckKey，支持阻塞
// 假定前面获取的消息已经确认，因该方法内部可能回滚确认队列，避免误杀
// timeout 超时时间，默认0秒永远阻塞；负数表示直接返回，不阻塞。
func (r *RedisReliableQueue) TakeOne(timeout ...int64) string {
	r.RetryAck()
	var timeOut int64
	if len(timeout) > 0 {
		timeOut = timeout[0]
	}
	var rs string
	if timeOut >= 0 {
		rs = r.client.WithDB(r.DB).WithContext(r.ctx).BRPopLPush(r.key, r.AckKey, time.Second*time.Duration(timeOut))
	} else {
		rs = r.client.WithDB(r.DB).WithContext(r.ctx).RPopLPush(r.key, r.AckKey)
	}
	if len(rs) > 0 {
		_Status.Consumes++
	}
	return rs
}

// Take 批量消费获取，从Key弹出并备份到AckKey
// 假定前面获取的消息已经确认，因该方法内部可能回滚确认队列，避免误杀
// count 要消费的消息个数
func (r *RedisReliableQueue) Take(count ...int) []string {
	var cCount = 1
	if len(count) > 0 {
		cCount = count[0]
	}
	var values []string
	for i := 0; i < cCount; i++ {
		rs := r.client.WithDB(r.DB).WithContext(r.ctx).RPopLPush(r.key, r.AckKey)
		if rs != "" {
			values = append(values, rs)
		}
	}
	return values

}

// Acknowledge 确认消费，从AckKey中删除
func (r *RedisReliableQueue) Acknowledge(keys ...string) int64 {
	var rs int64
	_Status.Acks += int64(len(keys))
	for _, item := range keys {
		val := r.client.WithDB(r.DB).WithContext(r.ctx).LRem(r.AckKey, 1, item)
		if val > 0 {
			rs += val
		}
	}
	return rs

}

var _delay *queue_delay.RedisDelayQueue

// InitDelay 初始化延迟队列功能。生产者自动初始化，消费者最好能够按队列初始化一次
// 该功能是附加功能，需要消费者主动调用，每个队列的多消费者开一个即可。
//  核心工作是启动延迟队列的TransferAsync大循环，每个进程内按队列开一个最合适，多了没有用反而形成争夺。
func (r *RedisReliableQueue) InitDelay() {
	if _delay == nil {
		queue_delay.New(r.client, fmt.Sprintf("%s:Delay", r.key), r.l)
	}
	go func() {
		_delay.TransferAsync(r.ctx)
	}()
}

// AddDelay 添加延迟消息
func (r *RedisReliableQueue) AddDelay(value interface{}, delay int64) int64 {
	r.InitDelay()
	return _delay.Add(value, delay)
}

// Publish 高级生产消息。消息体和消息键分离，业务层指定消息键，可随时查看或删除，同时避免重复生产
//  Publish 必须跟 ConsumeAsync 配对使用。
// messages 消息字典，id为键，消息体为值
// expire 消息体过期时间，单位秒
func (r *RedisReliableQueue) Publish(messages map[string]interface{}, expire int64) int64 {
	if messages == nil {
		return 0
	}
	var keys []string
	if expire > 0 {
		for key, val := range messages {
			keys = append(keys, key)
			r.client.WithDB(r.DB).WithContext(r.ctx).SetEX(key, val, time.Duration(expire)*time.Second)
		}

	} else {
		for key, val := range messages {
			keys = append(keys, key)
			r.client.WithDB(r.DB).WithContext(r.ctx).Set(key, val, 0)
		}
	}
	return r.client.WithDB(r.DB).WithContext(r.ctx).LPush(r.key, keys)

}

// Consume 高级消费消息。消息处理成功后，自动确认并删除消息体
// Publish 必须跟 ConsumeAsync 配对使用。
func (r *RedisReliableQueue) Consume(fn func(value interface{}) int64, timeOut ...int64) int64 {
	var timeout int64 = 0
	if len(timeOut) > 0 {
		timeout = timeOut[0]
	}

	r.RetryAck()

	var msgId string
	if timeout < 0 {
		msgId = r.client.WithDB(r.DB).WithContext(r.ctx).RPopLPush(r.key, r.AckKey)
	} else {
		msgId = r.client.WithDB(r.DB).WithContext(r.ctx).BRPopLPush(r.key, r.AckKey, time.Duration(timeout)*time.Second)
	}
	if msgId == "" {
		return 0
	}
	_Status.Consumes++
	// 取出消息。如果重复消费，或者业务层已经删除消息，此时将拿不到
	result, _ := r.client.WithDB(r.DB).WithContext(r.ctx).GetString(msgId)
	if result == "" {
		//   拿不到消息体，直接确认消息键
		r.Acknowledge(msgId)
		return 0
	}

	var rs = fn(result)

	// 确认并删除消息
	r.client.WithDB(r.DB).WithContext(r.ctx).Delete(msgId)
	r.Acknowledge(msgId)

	return rs
}

// TakeAck 从确认列表弹出消息，用于消费中断后，重新恢复现场时获取
// 理论上Ack队列只存储极少数数据
func (r *RedisReliableQueue) TakeAck(count ...int) []string {
	var cCount = 1
	if len(count) > 0 {
		cCount = count[0]
	}
	if cCount < 0 {
		return nil
	}
	var rs []string
	for i := 0; i < cCount; i++ {
		result := r.client.WithDB(r.DB).WithContext(r.ctx).RPop(r.AckKey)
		if result != nil {
			rs = append(rs, result.(string))
		}
	}
	return rs

}

// ClearAllAck 清空所有Ack队列。危险操作！！！
func (r *RedisReliableQueue) ClearAllAck() {
	// 先找到所有Key

	keys, _ := r.client.WithDB(r.DB).WithContext(r.ctx).Scan(0, fmt.Sprintf("%s:Ack:*", _Key), 1000)
	if len(keys) > 0 {
		r.client.WithDB(r.DB).WithContext(r.ctx).Delete(keys...)
	}

}

var _nextRetry int64

// RetryAck 消费获取，从Key弹出并备份到AckKey，支持阻塞 假定前面获取的消息已经确认，因该方法内部可能回滚确认队列，避免误杀 超时时间，默认0秒永远阻塞；负数表示直接返回，不阻塞。
func (r *RedisReliableQueue) RetryAck() {
	var now = time.Now()
	if _nextRetry < now.UnixMilli() {
		_nextRetry = now.Add(time.Second * time.Duration(r.RetryInterval)).UnixMilli()
		// 拿到死信，重新放入队列
		data := r.RollbackAck(_Key, r.AckKey)
		for _, item := range data {
			r.logger.Debug(fmt.Sprintf("定时回滚死信：%s", item))
		}
		// 更新状态
		r.UpdateStatus()
		// 处理其它消费者遗留下来的死信，需要抢夺全局清理权，减少全局扫描次数
		result := r.client.WithDB(r.DB).WithContext(r.ctx).SetNX(fmt.Sprintf("%s:AllStatus", _Key), _Status, time.Duration(r.RetryInterval)*time.Second)
		if result {
			r.RollbackAllAck()
		}

	}
}

// RollbackAck 回滚指定AckKey内的消息到Key
func (r *RedisReliableQueue) RollbackAck(key, ackKey string) []string {
	// 消费所有数据
	var data []string
	for {

		result := r.client.WithDB(r.DB).WithContext(r.ctx).RPopLPush(ackKey, key)
		if result == "" {
			break
		}
		data = append(data, result)
	}
	return data
}

// UpdateStatus 更新状态
func (r *RedisReliableQueue) UpdateStatus() {
	// 更新状态，7天过期
	_Status.LastActive = time.Now().UnixMilli()
	r.client.WithDB(r.DB).WithContext(r.ctx).Set(_StatusKey, _Status, 7*24*3600)
}

// RollbackAllAck 全局回滚死信，一般由单一线程执行，避免干扰处理中数据
func (r *RedisReliableQueue) RollbackAllAck() int64 {
	// 先找到所有Key
	var count int
	var ackKeys []string

	keys, cursor := r.client.WithDB(r.DB).WithContext(r.ctx).Scan(0, fmt.Sprintf("%s:Status:*", _Key), 1000)
	fmt.Println(cursor)
	for _, key := range keys {
		var ackKey = fmt.Sprintf("%s:Ack:%s", _Key, strings.TrimLeft(key, fmt.Sprintf("%s:Status:", _Key)))
		ackKeys = append(ackKeys, ackKey)

		var st = r.client.WithDB(r.DB).WithContext(r.ctx).Get(key)
		if st != nil {
			s, ok := st.(RedisQueueStatus)
			if ok {
				if s.LastActive+(r.RetryInterval+10)*1000 < time.Now().UnixMilli() {
					if r.client.WithDB(r.DB).WithContext(r.ctx).Exists(ackKey) > 0 {
						//r.logger.Debug(fmt.Sprintf("发现死信队列：%s", ackKey))
						r.logger.Debugf("发现死信队列：%v", ackKey)

						var list = r.RollbackAck(_Key, ackKey)
						for _, item := range list {
							r.logger.Debugf("全局回滚死信：%v", item)
						}
						count += len(list)
					}

					// 删除状态
					r.client.WithDB(r.DB).WithContext(r.ctx).Delete(key)
					r.logger.Debugf("删除队列状态：%v %v", key, gjson.Marshal(st))
				}
			}
		}
	}

	keys, cursor = r.client.WithDB(r.DB).WithContext(r.ctx).Scan(0, fmt.Sprintf("%s:Ack:*", _Key), 1000)
	fmt.Println(cursor)
	for _, key := range keys {

		if !stringArray(&ackKeys, key) {
			var msgs = r.client.WithDB(r.DB).WithContext(r.ctx).LRange(key, 0, -1)
			r.logger.Debugf("全局清理死信：%v %v", key, gjson.Marshal(msgs))
			r.client.WithDB(r.DB).WithContext(r.ctx).Delete(key)

		}

	}

	return int64(count)

}

func stringArray(keys *[]string, key string) bool {
	for _, val := range *keys {
		if val == key {
			return true
		}
	}
	return false
}
