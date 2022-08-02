package queue_delay

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/utils/cache"
)

type MQRedisDelay struct {
	client cache.ICache
	logger glog.ILogger
}

func NewMQRedisDelay(client cache.ICache, logger glog.ILogger) *MQRedisDelay {
	return &MQRedisDelay{
		client: client,
		logger: logger,
	}
}

func (r *MQRedisDelay) GetRedisDelay(topic string) *RedisDelayQueue {
	return NewRedisDelay(r.client, topic, r.logger)
}
