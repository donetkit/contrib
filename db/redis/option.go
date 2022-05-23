package redis

import (
	"context"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/tracer"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type config struct {
	ctx      context.Context
	logger   glog.ILogger
	tracer   *tracer.Server
	attrs    []attribute.KeyValue
	addr     string
	port     int
	password string
	db       int
}

// Option specifies instrumentation configuration options.
//type Option interface {
//	apply(*config)
//}

type Option func(p *config)

// WithTracer specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracer(tracerServer *tracer.Server) Option {
	return func(cfg *config) {
		if tracerServer != nil {
			cfg.tracer = tracerServer
		}
	}
}

// WithAttributes specifies additional attributes to be added to the span.
func WithAttributes(attrs ...string) Option {
	return func(cfg *config) {
		for _, attr := range attrs {
			cfg.attrs = append(cfg.attrs, semconv.NetPeerNameKey.String(attr))
		}
	}
}

// WithLogger prevents logger.
func WithLogger(logger glog.ILogger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

// WithAddr prevents addr.
func WithAddr(addr string) Option {
	return func(cfg *config) {
		cfg.addr = addr
	}
}

// WithPort prevents port.
func WithPort(port int) Option {
	return func(cfg *config) {
		cfg.port = port
	}
}

// WithPassword prevents password.
func WithPassword(password string) Option {
	return func(cfg *config) {
		cfg.password = password
	}
}

// WithDB prevents db.
func WithDB(db int) Option {
	return func(cfg *config) {
		cfg.db = db
	}
}
