package main

import (
	"context"
	"github.com/donetkit/contrib-log/glog"
	gorm "github.com/donetkit/contrib/db/gorm2"
	"github.com/donetkit/contrib/tracer"
	"time"
)

const (
	service     = "gorm-test"
	environment = "development" // "production" "development"
)

func main() {
	ctx := context.Background()
	log := glog.New()
	var traceServer *tracer.Server
	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}
	var dns = map[string]string{}
	dns["default"] = "root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms"
	sql := gorm.NewDb(gorm.WithDNS(dns), gorm.WithLogger(log), gorm.WithTracer(traceServer))
	defer func() {
		tp.Shutdown(context.Background())
	}()
	ctx, span := traceServer.Tracer.Start(ctx, "testgorm")
	defer span.End()
	var num []string
	if err := sql.DB().WithContext(ctx).Raw("SELECT id FROM test").Scan(&num).Error; err != nil {
		log.Error(err)
	}
	if err := sql.DB().Table("test").WithContext(ctx).Where("id != ''").Select("id").Scan(&num).Error; err != nil {
		log.Error(err)
	}
	time.Sleep(time.Second * 10)

}
