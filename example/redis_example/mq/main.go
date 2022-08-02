package main

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/db/queue/queue_stream"
	rredis "github.com/donetkit/contrib/db/redis"
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

	var RedisClient = rredis.New(rredis.WithLogger(logs), rredis.WithAddr("127.0.0.1"), rredis.WithDB(13), rredis.WithPassword(""), rredis.WithTracer(traceServer))

	fullRedis := queue_stream.NewMQRedisStream(RedisClient, logs)

	queue1 := fullRedis.GetStream(topic)
	queue1.BlockTime = 5
	queue1.SetGroup("Group1")
	queue1.ConsumeBlock(ctx, Consumer1)

	queue2 := fullRedis.GetStream(topic)
	queue2.SetGroup("Group2")
	queue2.ConsumeBlock(ctx, Consumer2)

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

func Public(fullRedis *queue_stream.MQRedisStream, topic string) {
	var index = 0
	queue1 := fullRedis.GetStream(topic)
	queue1.MaxLength = 1000
	for {
		queue1.Add(MyModel{Id: fmt.Sprintf("%d", index), Name: fmt.Sprintf("掌聲%d", index)})
		time.Sleep(time.Millisecond * 10)
		index++
		if index > 20 {
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
