package main

import (
	"context"
	"encoding/json"
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
	//
	//rdb1 := redisRedis.New(redisRedis.WithLogger(log), redisRedis.WithTracer(traceServer), redisRedis.WithAddr("192.168.5.110"), redisRedis.WithPassword("zxwP)O(I*u7y6t5"), redisRedis.WithDB(0))
	//msg11, err1 := rdb1.WithDB(13).XRead("FullQueue", 10)
	//if err1 != nil {
	//
	//}
	//log.Debug(msg11)
	rdb := redisRedis.NewRedisClient(redisRedis.WithLogger(log), redisRedis.WithTracer(traceServer), redisRedis.WithAddr("192.168.5.110"), redisRedis.WithPassword("zxwP)O(I*u7y6t5"), redisRedis.WithDB(13))

	var topic = "FullQueue"

	ctx, _ := context.WithCancel(context.Background())
	fullRedis := queue.NewFullRedis(rdb, log)

	queue1 := fullRedis.GetStream(topic)
	queue1.SetGroup("Group1")
	queue1.ConsumeAsync(ctx, func(msg []redis.XMessage) {
		log.Debug("==========ConsumeAsync===Group1=======", msg)
		for _, message := range msg {
			queue1.Delete(message.ID)
			log.Debug("删除Id: ", message.ID)
		}
	})

	queue2 := fullRedis.GetStream(topic)
	queue2.SetGroup("Group2")
	queue2.ConsumeAsync(ctx, func(msg []redis.XMessage) {
		log.Debug("==========ConsumeAsync====Group2======", msg)

		for _, message := range msg {
			queue2.Delete(message.ID)
			log.Debug("删除Id: ", message.ID)
		}

	})

	go func() {
		Public(fullRedis, topic)
	}()
	time.Sleep(time.Second * 1)
	//
	//var key = "StreamQueue_Test"
	//mqQueue := queue.NewRedisStream(rdb[14], key)
	//
	//mqQueue.SetGroup("MQGroup")
	////
	////mqQueue.Add(ToJson(MyModel{Id: "4", Name: "掌聲4"}))
	////mqQueue.Add(ToJson(MyModel{Id: "5", Name: "掌聲5"}))
	////mqQueue.Add(ToJson(MyModel{Id: "6", Name: "掌聲6"}))
	////mqQueue.Add(ToJson(MyModel{Id: "7", Name: "掌聲7"}))
	//for i := 0; i < 3; i++ {
	//	var msg = mqQueue.TakeOne()
	//	log.Debug(msg)
	//	if msg != nil {
	//		mqQueue.Acknowledge(msg[0].ID)
	//	}
	//	time.Sleep(time.Second * 1)
	//
	//}
	time.Sleep(time.Second * 13)
	//cancel()
	time.Sleep(time.Second * 33333)

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

func Public(fullRedis *queue.FullRedis, topic string) {

	var index = 200
	queue1 := fullRedis.GetStream(topic)
	for {
		queue1.Add(ToJson(MyModel{Id: fmt.Sprintf("%d", index), Name: fmt.Sprintf("掌聲%d", index)}))
		time.Sleep(time.Second * 1)
		index++
		if index > 210 {
			break
		}
	}
	//queue1.Add(ToJson(MyModel{Id: "40", Name: "掌聲40"}))
	//time.Sleep(time.Second * 3)
	//queue1.Add(ToJson(MyModel{Id: "50", Name: "掌聲50"}))
	//time.Sleep(time.Second * 3)
	//queue1.Add(ToJson(MyModel{Id: "60", Name: "掌聲60"}))
	//time.Sleep(time.Second * 3)
}
