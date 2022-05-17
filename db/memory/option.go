package memory

import (
	"context"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/tracer"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"time"
)

type config struct {
	ctx               context.Context
	logger            glog.ILogger
	tracerServer      *tracer.Server
	attrs             []attribute.KeyValue
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
}

type Option func(p *config)

func WithDefaultExpiration(defaultExpiration time.Duration) Option {
	return func(cfg *config) {
		cfg.defaultExpiration = defaultExpiration
	}
}

func WithCleanupInterval(cleanupInterval time.Duration) Option {
	return func(cfg *config) {
		cfg.cleanupInterval = cleanupInterval
	}
}

// WithTracer specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracer(tracerServer *tracer.Server) Option {
	return func(cfg *config) {
		if tracerServer != nil {
			cfg.tracerServer = tracerServer
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
