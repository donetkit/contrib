package queue

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/go-redis/redis/v8"
)

type MQRedisStream struct {
	client *redis.Client
	logger glog.ILogger
}

func NewMQRedisStream(client *redis.Client, logger glog.ILogger) *MQRedisStream {
	return &MQRedisStream{
		client: client,
		logger: logger,
	}
}

func (r *MQRedisStream) GetStream(topic string) *RedisStream {
	return NewRedisStream(r.client, topic, r.logger)
}
