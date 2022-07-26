package main

import (
	"encoding/json"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/db/queue"
	redisRedis "github.com/donetkit/contrib/db/redis"
	"github.com/donetkit/contrib/tracer"
	"time"
)

const (
	service     = "redis-mq-test"
	environment = "development" // "production" "development"
)

func main() {
	//ctx := context.Background()

	log := glog.New()
	var traceServer *tracer.Server
	// 1
	fs := tracer.NewFallbackSampler(1)
	tp, err := tracer.NewTracerProvider(service, "192.168.5.110", environment, 6831, fs)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}

	rdb := redisRedis.NewRedisClient(redisRedis.WithLogger(log), redisRedis.WithTracer(traceServer), redisRedis.WithAddr("192.168.5.110"), redisRedis.WithPassword("zxwP)O(I*u7y6t5"), redisRedis.WithDB(0))
	var key = "StreamQueue_Test"
	mqQueue := queue.NewRedisStream(rdb[14], key)

	mqQueue.SetGroup("MQGroup")
	//
	//mqQueue.Add(ToJson(MyModel{Id: "4", Name: "掌聲4"}))
	//mqQueue.Add(ToJson(MyModel{Id: "5", Name: "掌聲5"}))
	//mqQueue.Add(ToJson(MyModel{Id: "6", Name: "掌聲6"}))

	for i := 0; i < 3; i++ {
		var msg = mqQueue.TakeOne()
		log.Debug(msg)
		if msg != nil {
			mqQueue.Acknowledge(msg[0].ID)
		}
		time.Sleep(time.Second * 1)

	}

	time.Sleep(time.Second * 5)

}

type MyModel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func ToJson(value interface{}) string {
	jsonStr, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(jsonStr)
}
