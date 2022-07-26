package queue

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type IProducerConsumer interface {
	// Count 元素个数
	Count() int64

	// IsEmpty 集合是否为空
	IsEmpty() bool

	// Add 生产添加
	Add(params interface{}, msgId ...string) string

	// Take 消费获取一批
	Take(count int64) []redis.XMessage

	TakeOne() []redis.XMessage

	// TakeOneAsync 异步消费获取一个
	TakeOneAsync(ctx context.Context, timeout int64) []redis.XMessage

	// Acknowledge 确认消费
	Acknowledge(keys ...string) int64
}
