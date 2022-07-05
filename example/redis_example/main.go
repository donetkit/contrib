package main

import (
	"context"
	"github.com/donetkit/contrib-log/glog"
	redisRedis "github.com/donetkit/contrib/db/redis"
	"github.com/donetkit/contrib/tracer"
	"github.com/donetkit/contrib/utils/cache"
	"net/url"
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
	//fs := tracer.NewFallbackSampler(0.1)

	url1, _ := url.Parse("http://127.0.0.1")

	rs, err := tracer.NewRemoteSampler(ctx, service, 0.05, tracer.WithLogger(log), tracer.WithSamplingRulesPollingInterval(time.Second*10), tracer.WithEndpoint(*url1))
	if err == nil {

	}

	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, rs)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}

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

	return nil
}
