package queue_stream

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/utils/cache"
)

type MQRedisStream struct {
	client cache.ICache
	logger glog.ILogger
}

func NewMQRedisStream(client cache.ICache, logger glog.ILogger) *MQRedisStream {
	return &MQRedisStream{
		client: client,
		logger: logger,
	}
}

func (r *MQRedisStream) GetStream(topic string) *RedisStream {
	return NewRedisStream(r.client, topic, r.logger)
}
