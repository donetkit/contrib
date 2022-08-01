package queue

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/utils/cache"
	"time"
)

type RedisDelayQueue struct {
	DB int

	// 失败时抛出异常。默认false
	ThrowOnFailure bool

	//发送消息失败时的重试次数。默认3次
	RetryTimesWhenSendFailed int

	// 重试间隔。默认1000ms
	RetryIntervalWhenSendFailed int

	//消息队列主题
	Topic string

	key string
	ctx context.Context
	// 转移延迟消息到主队列的间隔。默认10s
	TransferInterval int64

	// 个数
	Count int64

	IsEmpty bool

	// 默认延迟时间。默认60秒
	Delay int64

	logger glog.ILoggerEntry

	client cache.ICache
}

func NewRedisDelay(client cache.ICache, key string, logger glog.ILogger) *RedisDelayQueue {
	return &RedisDelayQueue{
		RetryTimesWhenSendFailed:    3,
		RetryIntervalWhenSendFailed: 1000,
		logger:                      logger.WithField("MQ_REDIS_DELAY", "MQ_REDIS_DELAY"),
		key:                         key,
		TransferInterval:            10,
		Delay:                       60,
		//MaxLength:                   1_000_000,
		//MaxRetry:                    10,
		//BlockTime:                   15,
		//StartId:                     "0-0",
		Topic:  key,
		client: client,
		ctx:    context.Background(),
	}
}

// Add 添加延迟消息
func (r *RedisDelayQueue) Add(value interface{}, delay int64) int64 {
	if value == nil {
		return 0
	}
	var target = time.Now().Unix() + delay
	var rs int64
	for i := 0; i < r.RetryTimesWhenSendFailed; i++ {
		// 添加到有序集合的成员数量，不包括已经存在更新分数的成员
		rs = r.client.WithContext(r.ctx).WithContext(r.ctx).ZAdd(r.key, float64(target), value)
		if rs >= 0 {
			return rs
		}

		r.logger.Debug(fmt.Sprintf("发布到队列[%s]失败！", r.Topic))

		if i < r.RetryTimesWhenSendFailed {
			time.Sleep(time.Duration(r.RetryIntervalWhenSendFailed) * time.Millisecond)
		}
	}
	return rs

}

// Adds 批量生产
func (r *RedisDelayQueue) Adds(values ...interface{}) int64 {
	if values == nil || len(values) == 0 {
		return 0
	}

	var target = time.Now().Unix() + r.Delay

	var rs int64
	for i := 0; i < r.RetryTimesWhenSendFailed; i++ {
		// 添加到有序集合的成员数量，不包括已经存在更新分数的成员
		rs = r.client.WithContext(r.ctx).ZAdd(r.key, float64(target), values)
		if rs >= 0 {
			return rs
		}

		r.logger.Debug(fmt.Sprintf("发布到队列[%s]失败！", r.Topic))

		if i < r.RetryTimesWhenSendFailed {
			time.Sleep(time.Duration(r.RetryIntervalWhenSendFailed) * time.Millisecond)
		}
	}
	return rs

}

// Remove 删除项
func (r *RedisDelayQueue) Remove(value ...interface{}) int64 {
	return r.client.WithContext(r.ctx).ZRem(r.key, value...)
}

// TakeOne
// timeout 超时时间，默认0秒永远阻塞；负数表示直接返回，不阻塞。获取一个
func (r *RedisDelayQueue) TakeOne(timeout ...int64) string {
	var timeOut int64 = 60

	if len(timeout) > 0 {
		timeOut = timeout[0]
	}

	for {
		var score = time.Now().Unix()

		rs := r.client.WithContext(r.ctx).ZRangeByScore(r.key, 0, score, 0, 1)
		if len(rs) > 0 && r.TryPop(rs[0]) {
			return rs[0]
		}
		// 是否需要等待
		if timeOut <= 0 {
			break
		}
		time.Sleep(time.Second * 1)
		timeOut--
	}
	return ""
}

// TakeOneBlock
// timeout 超时时间，默认0秒永远阻塞；负数表示直接返回，不阻塞。异步获取一个
func (r *RedisDelayQueue) TakeOneBlock(ctx context.Context, timeout ...int64) string {
	var timeOut int64 = 60

	if len(timeout) > 0 {
		timeOut = timeout[0]
	}

	for {
		var score = time.Now().Unix()

		rs := r.client.WithContext(r.ctx).ZRangeByScore(r.key, 0, score, 0, 1)
		if len(rs) > 0 {
			return rs[0]
		}
		// 是否需要等待
		if timeOut <= 0 {
			break
		}
		time.Sleep(time.Second * 1)
		timeOut--
	}
	return ""
}

//func (r *RedisDelayQueue) TakeOneBlock(timeout ...int64) {
//
//}

// Take 获取一批
func (r *RedisDelayQueue) Take(count int64) []string {
	if count <= 0 {
		return nil
	}
	var score = time.Now().Unix()

	rs := r.client.WithContext(r.ctx).ZRangeByScore(r.key, 0, score, 0, 1)
	if len(rs) <= 0 {
		return nil
	}

	var arr []string
	for _, item := range rs {
		// 争夺消费
		if r.TryPop(item) {
			arr = append(arr, item)
		}
	}
	return arr
}

// TryPop 争夺消费，只有一个线程能够成功删除，
func (r *RedisDelayQueue) TryPop(value interface{}) bool {
	return r.Remove(value) > 0
}

// Acknowledge 确认删除
func (r *RedisDelayQueue) Acknowledge(keys ...string) int64 {
	return -1
}

func (r *RedisDelayQueue) TransferAsync(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // 退出了...
			default:
				var score = time.Now().Unix()
				rs := r.client.WithContext(r.ctx).ZRangeByScore(r.key, 0, score, 0, 10)
				if len(rs) > 0 {
					// 逐个删除，多线程争夺可能失败
					var arr []string
					for _, item := range rs {
						if r.Remove(item) > 0 {
							arr = append(arr, item)
						}
					}
					// 转移消息
					if len(arr) > 0 {
						r.Adds(arr)
					}

				} else {
					// 没有消息，歇一会
					time.Sleep(time.Duration(r.TransferInterval) * time.Second)
				}
			}

		}
	}()
}
