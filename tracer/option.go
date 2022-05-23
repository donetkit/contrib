package tracer

import (
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Server struct {
	tracerName     string
	Tracer         oteltrace.Tracer
	TracerProvider oteltrace.TracerProvider
	Propagators    propagation.TextMapPropagator
}

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*Server)
}

type optionFunc func(*Server)

func (o optionFunc) apply(c *Server) {
	o(c)
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *Server) {
		if propagators != nil {
			cfg.Propagators = propagators
		}
	})
}

// WithProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithProvider(provider oteltrace.TracerProvider) Option {
	return optionFunc(func(cfg *Server) {
		if provider != nil {
			cfg.TracerProvider = provider
		}
	})
}

// WithName name default Service
func WithName(name string) Option {
	return optionFunc(func(cfg *Server) {
		cfg.tracerName = name
	})
}
