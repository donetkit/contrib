package queue_reliable

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/utils/grand"
	"github.com/go-redis/redis/v8"
	"github.com/shirou/gopsutil/v3/host"
	"os"
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
	client                      *redis.Client   // Client
	ctx                         context.Context // Context
}

func CreateStatus() RedisQueueStatus {
	info, _ := host.Info()
	return RedisQueueStatus{
		Key:         grand.RandAllString(8),
		MachineName: info.Hostname,
		UserName:    "",
		ProcessId:   os.Getpid(),
		Ip:          "",
		CreateTime:  time.Now().UnixMilli(),
		LastActive:  time.Now().UnixMilli(),
	}
}

func NewRedisReliable(client *redis.Client, key string, logger glog.ILogger) *RedisReliableQueue {
	_Key = key
	_Status = CreateStatus()
	_StatusKey = fmt.Sprintf("%s:Status:%s", key, _Status.Key)
	return &RedisReliableQueue{
		key:                         key,
		RetryTimesWhenSendFailed:    3,
		RetryIntervalWhenSendFailed: 1000,
		RetryInterval:               60,
		MinPipeline:                 3,
		logger:                      logger.WithField("MQ_REDIS_DELAY", "MQ_REDIS_DELAY"),
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
		rs, _ = r.client.LPush(r.ctx, r.key, values...).Result()
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

func (r *RedisReliableQueue) TakeOne(timeout ...int64) string {
	r.RetryAck()
	var timeOut int64
	if len(timeout) > 0 {
		timeOut = timeout[0]
	}
	var rs string
	if timeOut >= 0 {
		rs = r.client.BRPopLPush(r.ctx, r.key, r.AckKey, time.Second*time.Duration(timeOut)).Val()
	} else {
		rs = r.client.RPopLPush(r.ctx, r.key, r.AckKey).Val()
	}
	if len(rs) > 0 {
		_Status.Consumes++
	}
	return rs
}

func (r *RedisReliableQueue) Take(count ...int64) {
	var ccount int64 = 1
	if len(count) > 0 {
		ccount = count[0]
	}
	if ccount > r.MinPipeline {

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
		result := r.client.SetNX(r.ctx, fmt.Sprintf("%s:AllStatus", _Key), _Status, time.Duration(r.RetryInterval)*time.Second).Val()
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

		result := r.client.RPopLPush(r.ctx, ackKey, key).Val()
		if result != "" {
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
	r.client.Set(r.ctx, _StatusKey, _Status, 7*24*3600)
}

// RollbackAllAck 全局回滚死信，一般由单一线程执行，避免干扰处理中数据
func (r *RedisReliableQueue) RollbackAllAck() int64 {
	// TODO 待完善
	// 先找到所有Key
	var count int
	//var acks []string
	//
	//keys, cursor := r.client.Scan(r.ctx, 0, fmt.Sprintf("%s:Status:*", _Key), 1000).Val()
	//fmt.Println(cursor)
	//for _, key := range keys {
	//	var ackKey = fmt.Sprintf("%s:Ack:%s", _Key, strings.TrimLeft(key, fmt.Sprintf("%s:Status:", _Key)))
	//	acks = append(acks, ackKey)
	//
	//	var st = r.client.Get(r.ctx, key).Val()
	//	if st != nil {
	//		s, ok := st.(RedisQueueStatus)
	//		if ok {
	//			if s.LastActive+(r.RetryInterval+10)*1000 < time.Now().UnixMilli() {
	//				if r.client.Exists(r.ctx, ackKey).Val() > 0 {
	//					//r.logger.Debug(fmt.Sprintf("发现死信队列：%s", ackKey))
	//					r.logger.Debugf("发现死信队列：%v", ackKey)
	//
	//					var list = r.RollbackAck(_Key, ackKey)
	//					for _, item := range list {
	//						r.logger.Debugf("全局回滚死信：%v", item)
	//					}
	//					count += len(list)
	//				}
	//
	//				// 删除状态
	//				r.client.Del(r.ctx, key).Result()
	//				r.logger.Debugf("删除队列状态：%v %v", st.ToJson())
	//			}
	//		}
	//	}
	//}
	//
	//keys, cursor = r.client.Scan(r.ctx, 0, fmt.Sprintf("%s:Ack:*", _Key), 1000).Val()
	//fmt.Println(cursor)
	//for _, key := range keys {
	//
	//	if !stringArray(&acks, key) {
	//		var msgs = r.client.LRange(r.ctx, key, 0, -1).Val()
	//		r.logger.Debugf("全局清理死信：%v %v", key, msgs.ToJson())
	//		r.client.Del(r.ctx, key).Result()
	//
	//	}
	//
	//}

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
