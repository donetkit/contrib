package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

const redisClientDBKey = "RedisClientDBKey"

var allClient []*redis.Client

type Cache struct {
	db     int
	ctx    context.Context
	client *redis.Client
	config *config
}

func New(opts ...Option) *Cache {
	c := &config{
		ctx:      context.TODO(),
		addr:     "127.0.0.1",
		port:     6379,
		password: "",
		db:       0,
	}
	for _, opt := range opts {
		opt(c)
	}
	allClient = c.newRedisClient()
	return &Cache{config: c, ctx: c.ctx, client: allClient[c.db]}
}

func (c *config) newRedisClient() []*redis.Client {
	redisClients := make([]*redis.Client, 0)
	redisClients = append(redisClients, c.newClient(0))
	redisClients = append(redisClients, c.newClient(1))
	redisClients = append(redisClients, c.newClient(2))
	redisClients = append(redisClients, c.newClient(3))
	redisClients = append(redisClients, c.newClient(4))
	redisClients = append(redisClients, c.newClient(5))
	redisClients = append(redisClients, c.newClient(6))
	redisClients = append(redisClients, c.newClient(7))
	redisClients = append(redisClients, c.newClient(8))
	redisClients = append(redisClients, c.newClient(9))
	redisClients = append(redisClients, c.newClient(10))
	redisClients = append(redisClients, c.newClient(11))
	redisClients = append(redisClients, c.newClient(12))
	redisClients = append(redisClients, c.newClient(13))
	redisClients = append(redisClients, c.newClient(14))
	redisClients = append(redisClients, c.newClient(15))
	return redisClients
}

func (c *config) newClient(db int) *redis.Client {
	addr := fmt.Sprintf("%s:%d", c.addr, c.port)
	client := redis.NewClient(&redis.Options{Addr: addr, Password: c.password, DB: db})
	_, err := client.Ping(context.Background()).Result() // 检测心跳
	if err != nil {
		if c.logger != nil {
			c.logger.Error("connect redis failed" + err.Error())
		}
	}
	if c.tracer != nil {
		client.AddHook(newTracingHook(c.logger, c.tracer, c.attrs))
	}
	return client
}

func NewRedisClient(opts ...Option) *redis.Client {
	c := &config{
		ctx:      context.TODO(),
		addr:     "127.0.0.1",
		port:     6379,
		password: "",
		db:       0,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.newClient(c.db)
}
