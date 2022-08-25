package queue_reliable

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/utils/cache"
)

type ReliableQueue struct {
	client cache.ICache
	logger glog.ILogger
}

func NewReliableQueue(client cache.ICache, logger glog.ILogger) *ReliableQueue {
	return &ReliableQueue{
		client: client,
		logger: logger,
	}
}

func (r *ReliableQueue) GetReliableQueue(topic string) *RedisReliableQueue {
	return NewRedisReliable(r.client, topic, r.logger)
}
