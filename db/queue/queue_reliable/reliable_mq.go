package queue_reliable

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/utils/cache"
)

type MQRedisReliable struct {
	client cache.ICache
	logger glog.ILogger
}

func NewMQRedisReliable(client cache.ICache, logger glog.ILogger) *MQRedisReliable {
	return &MQRedisReliable{
		client: client,
		logger: logger,
	}
}

func (r *MQRedisReliable) GetReliableQueue(topic string) *RedisReliableQueue {
	return NewRedisReliable(r.client, topic, r.logger)
}
