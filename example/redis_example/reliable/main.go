package main

import (
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/db/queue/queue_reliable"
	rredis "github.com/donetkit/contrib/db/redis"
	"github.com/donetkit/contrib/tracer"
	"time"
)

const (
	service     = "redis-mq-test"
	environment = "development" // "production" "development"
	topic       = "Full_Queue_Test_Dev"
)

var logs = glog.New()

func main() {
	//ctx, _ := context.WithCancel(context.Background())
	var traceServer *tracer.Server

	fs := tracer.NewFallbackSampler(1)
	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, fs)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}
	var RedisClient = rredis.New(rredis.WithLogger(logs), rredis.WithAddr("127.0.0.1"), rredis.WithDB(13), rredis.WithPassword(""), rredis.WithTracer(traceServer))

	fullRedis := queue_reliable.NewReliableQueue(RedisClient, logs)
	go func() {
		queue1 := fullRedis.GetReliableQueue(topic)
		queue1.DB = 13
		// 清空
		//queue1.ClearAllAck()
		for {
			time.Sleep(time.Millisecond * 100)
			var mqMsg = queue1.TakeOne(10)
			if len(mqMsg) > 0 {
				fmt.Println(mqMsg)
				queue1.Acknowledge(mqMsg)
			}
		}
	}()

	go func() {
		Public(fullRedis, topic)
	}()
	time.Sleep(time.Second * 3600)

}

func Public(fullRedis *queue_reliable.ReliableQueue, topic string) {
	var index = 0
	queue1 := fullRedis.GetReliableQueue(topic)
	queue1.DB = 13
	for {
		queue1.Add(fmt.Sprintf("%d", index), 60)
		time.Sleep(time.Millisecond * 50)
		index++
		if index > 20 {
			break
		}
	}
}
