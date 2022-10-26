package main

import (
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/db/queue/queue_delay"
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
	var traceServer *tracer.Server

	fs := tracer.NewFallbackSampler(1)
	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, fs)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}
	var RedisClient = rredis.New(rredis.WithLogger(logs), rredis.WithAddr("127.0.0.1"), rredis.WithDB(13), rredis.WithPassword(""), rredis.WithTracer(traceServer))

	//var RedisClient = rredis.NewRedisClient(rredis.WithLogger(logs), rredis.WithAddr("127.0.0.1"), rredis.WithDB(14), rredis.WithPassword(""), rredis.WithTracer(traceServer))

	delayQueue := queue_delay.NewDelayQueue(RedisClient, logs)
	go func() {
		queue1 := delayQueue.GetDelayQueue(topic)
		for {
			var msg = queue1.TakeOne(10)
			if len(msg) > 0 {
				fmt.Println(msg)
			}
		}
	}()

	go func() {
		Public(delayQueue, topic)
	}()

	time.Sleep(time.Second * 3600)

}

func Public(delayQueue *queue_delay.DelayQueue, topic string) {
	var index = 0
	queue := delayQueue.GetDelayQueue(topic)
	for {
		queue.Add(fmt.Sprintf("%d", index), 60)
		time.Sleep(time.Millisecond * 1000)
		index++
		if index > 20 {
			break
		}
	}
}
