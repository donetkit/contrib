package queue

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/go-redis/redis/v8"
)

type FullRedis struct {
	client *redis.Client
	logger glog.ILogger
}

func NewFullRedis(client *redis.Client, logger glog.ILogger) *FullRedis {
	return &FullRedis{
		client: client,
		logger: logger,
	}
}

func (r *FullRedis) GetStream(topic string) *RedisStream {
	return NewRedisStream(r.client, topic, r.logger)
}
