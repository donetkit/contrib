package queue_delay

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/utils/cache"
)

type DelayQueue struct {
	client cache.ICache
	logger glog.ILogger
}

func NewDelayQueue(client cache.ICache, logger glog.ILogger) *DelayQueue {
	return &DelayQueue{
		client: client,
		logger: logger,
	}
}

func (r *DelayQueue) GetDelayQueue(topic string) *RedisDelayQueue {
	return New(r.client, topic, r.logger)
}
