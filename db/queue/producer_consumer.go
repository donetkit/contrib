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
	Add(value interface{}, msgId ...string) string

	// Take 消费获取一批
	Take(count int64) []redis.XMessage

	// TakeOne 消费获取一个
	TakeOne() []redis.XMessage

	// TakeOneBlock 异步消费获取一个
	TakeOneBlock(ctx context.Context, timeout int64) []redis.XMessage

	// Acknowledge 确认消费
	Acknowledge(keys ...string) int64

	// Delete 删除消费消息
	Delete(id ...string) int64
}
