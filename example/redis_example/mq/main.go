package main

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/db/queue"
	redisRedis "github.com/donetkit/contrib/db/redis"
	"github.com/donetkit/contrib/tracer"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	service     = "redis-mq-test"
	environment = "development" // "production" "development"
	topic       = "Full_Queue_Test_Dev"
)

var logs = glog.New()

func main() {
	ctx, _ := context.WithCancel(context.Background())
	var traceServer *tracer.Server

	fs := tracer.NewFallbackSampler(1)
	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, fs)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}

	rdb := redisRedis.NewRedisClient(redisRedis.WithLogger(logs), redisRedis.WithTracer(traceServer), redisRedis.WithAddr("127.0.0.1"), redisRedis.WithPassword(""), redisRedis.WithDB(13))

	fullRedis := queue.NewMQRedisStream(rdb, logs)

	queue1 := fullRedis.GetStream(topic)
	queue1.SetGroup("Group1")
	queue1.ConsumeBlock(ctx, Consumer1)

	queue2 := fullRedis.GetStream(topic)
	queue2.SetGroup("Group2")
	queue2.ConsumeBlock(ctx, Consumer1)

	go func() {
		Public(fullRedis, topic)
	}()

	//cancel()

	time.Sleep(time.Second * 3600)

}

type MyModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func Public(fullRedis *queue.MQRedisStream, topic string) {
	var index = 0
	queue1 := fullRedis.GetStream(topic)
	queue1.MaxLength = 1000
	for {
		queue1.Add(MyModel{Id: fmt.Sprintf("%d", index), Name: fmt.Sprintf("掌聲%d", index)})
		time.Sleep(time.Millisecond * 10)
		index++
		if index > 10000 {
			break
		}
	}
}

func Consumer1(msg []redis.XMessage) bool {
	logs.Debug("==========Consumer1==========", msg)
	return true
}

func Consumer2(msg []redis.XMessage) bool {
	logs.Debug("==========Consumer2==========", msg)
	return true
}
