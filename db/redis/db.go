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
	ctx := context.Background()
	redisClients := make([]*redis.Client, 0)
	redisClients = append(redisClients, c.newClient(ctx, 0))
	redisClients = append(redisClients, c.newClient(ctx, 1))
	redisClients = append(redisClients, c.newClient(ctx, 2))
	redisClients = append(redisClients, c.newClient(ctx, 3))
	redisClients = append(redisClients, c.newClient(ctx, 4))
	redisClients = append(redisClients, c.newClient(ctx, 5))
	redisClients = append(redisClients, c.newClient(ctx, 6))
	redisClients = append(redisClients, c.newClient(ctx, 7))
	redisClients = append(redisClients, c.newClient(ctx, 8))
	redisClients = append(redisClients, c.newClient(ctx, 9))
	redisClients = append(redisClients, c.newClient(ctx, 10))
	redisClients = append(redisClients, c.newClient(ctx, 11))
	redisClients = append(redisClients, c.newClient(ctx, 12))
	redisClients = append(redisClients, c.newClient(ctx, 13))
	redisClients = append(redisClients, c.newClient(ctx, 14))
	redisClients = append(redisClients, c.newClient(ctx, 15))
	return redisClients
}

func (c *config) newClient(ctx context.Context, db int) *redis.Client {
	addr := fmt.Sprintf("%s:%d", c.addr, c.port)
	client := redis.NewClient(&redis.Options{Addr: addr, Password: c.password, DB: db})

	_, err := client.Ping(ctx).Result() // 检测心跳
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
		ctx:      context.Background(),
		addr:     "127.0.0.1",
		port:     6379,
		password: "",
		db:       0,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.newClient(c.ctx, c.db)
}
