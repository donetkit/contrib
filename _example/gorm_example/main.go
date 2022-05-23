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
	tp, err := tracer.NewTracerProvider(service, "192.168.5.110", environment, 6831)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	}
	var dns = map[string]string{}
	dns["default"] = "root:zxw123456@tcp(192.168.5.110:3306)/go_red_sentinel?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms"
	sql := gorm.NewDb(gorm.WithDNS(dns), gorm.WithLogger(log), gorm.WithTracer(traceServer))
	defer func() {
		tp.Shutdown(context.Background())
	}()
	ctx, span := traceServer.Tracer.Start(ctx, "testgorm")
	defer span.End()
	var num []string
	if err := sql.DB().WithContext(ctx).Raw("SELECT id FROM school").Scan(&num).Error; err != nil {
		log.Error(err)
	}
	if err := sql.DB().Table("school").WithContext(ctx).Where("id != ''").Select("id").Scan(&num).Error; err != nil {
		log.Error(err)
	}
	time.Sleep(time.Second * 10)

}
