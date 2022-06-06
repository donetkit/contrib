package gorm

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/tracer"
	"go.opentelemetry.io/otel/attribute"
	"time"
)

const (
	callBackBeforeName = "otel:before"
	callBackAfterName  = "otel:after"
	opCreate           = "INSERT"
	opQuery            = "SELECT"
	opDelete           = "DELETE"
	opUpdate           = "UPDATE"
)

type Config struct {
	Logger           glog.ILogger
	TracerServer     *tracer.Server
	Attrs            []attribute.KeyValue
	ExcludeQueryVars bool
	ExcludeMetrics   bool
	QueryFormatter   func(query string) string
}

type LogSql struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	Logger                    glog.ILogger
}
