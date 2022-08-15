package main

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	redisRedis "github.com/donetkit/contrib/db/redis"
	"github.com/donetkit/contrib/tracer"
	"github.com/donetkit/contrib/utils/cache"
	"time"
)

const (
	service     = "redis-test"
	environment = "development" // "production" "development"
)

func main() {
	ctx := context.Background()

	log := glog.New()
	var traceServer *tracer.Server

	// 1
	fs := tracer.NewFallbackSampler(0.1)
	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, fs)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}

	// 2
	//url1, _ := url.Parse("http://127.0.0.1")
	//rs, err := tracer.NewRemoteSampler(ctx, service, 0.05, tracer.WithLogger(log), tracer.WithSamplingRulesPollingInterval(time.Second*10), tracer.WithEndpoint(*url1))
	//if err == nil {
	//	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, rs)
	//	if err == nil {
	//		jaeger := tracer.Jaeger{}
	//		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	//	}
	//}

	rdb := redisRedis.New(redisRedis.WithLogger(log), redisRedis.WithTracer(traceServer), redisRedis.WithAddr("127.0.0.1"), redisRedis.WithPassword(""), redisRedis.WithDB(0))
	if err := redisCommands(ctx, traceServer, rdb); err != nil {
		log.Error(err.Error())
		return
	}

	for {
		time.Sleep(time.Second * 5)
		if err := redisCommands(ctx, traceServer, rdb); err != nil {
			log.Error(err.Error())
			return
		}
	}

	time.Sleep(time.Hour)

}
func redisCommands(ctx context.Context, traceServer *tracer.Server, rdb cache.ICache) error {
	ctx, span := traceServer.Tracer.Start(ctx, "cache")
	defer span.End()
	if err := rdb.WithDB(0).WithContext(ctx).Set("foo", "bar", 0); err != nil {
		return err
	}
	rdb.WithDB(0).WithContext(ctx).Get("foo")

	rdb.WithDB(0).WithContext(ctx).HashSet("myhash", map[string]interface{}{"key1": "1", "key2": "2"})
	rdb.WithDB(0).WithContext(ctx).HashSet("myhash", "key3", "3", "key4", "4")
	rdb.WithDB(0).WithContext(ctx).HashSet("myhash", []string{"key5", "5", "key6", "6"})

	fmt.Println(rdb.WithDB(0).WithContext(ctx).HashLen("myhash"))

	fmt.Println(rdb.WithDB(0).WithContext(ctx).HashAll("myhash"))

	fmt.Println(rdb.WithDB(0).WithContext(ctx).HashExist("myhash", "key33"))

	fmt.Println(rdb.WithDB(0).WithContext(ctx).HashGet("myhash", "key1"))

	fmt.Println(rdb.WithDB(0).WithContext(ctx).HashGets("myhash", "key1", "key6"))

	fmt.Println(rdb.WithDB(0).WithContext(ctx).HashKeys("myhash"))

	fmt.Println(rdb.WithDB(0).WithContext(ctx).HashAll("myhash"))

	rdb.WithDB(0).WithContext(ctx).HashDel("myhash", "key33")

	return nil
}
