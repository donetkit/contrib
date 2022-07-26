package tracer

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"

	otelBaggage "go.opentelemetry.io/otel/baggage"
	otelTrace "go.opentelemetry.io/otel/trace"
)

// New returns *tracer.Server
func New(opts ...Option) *Server {
	cfg := &Server{
		tracerName: "Service",
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	cfg.Tracer = cfg.TracerProvider.Tracer(
		cfg.tracerName,
		otelTrace.WithInstrumentationVersion(SemVersion()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	return cfg
}

func (s *Server) Stop(ctx context.Context) {
	tp, ok := s.TracerProvider.(*trace.TracerProvider)
	if ok {
		tp.Shutdown(ctx)
	}
}

func (s *Server) SpanFromContext(ctx context.Context) otelTrace.Span {
	return otelTrace.SpanFromContext(ctx)
}
func (s *Server) FromContext(ctx context.Context) otelBaggage.Baggage {
	return otelBaggage.FromContext(ctx)
}

// WithAttributes adds the attributes related to a span life-cycle event.
// These attributes are used to describe the work a Span represents when this
// option is provided to a Span's start or end events. Otherwise, these
// attributes provide additional information about the event being recorded
// (e.g. error, state change, processing progress, system event).
//
// If multiple of these options are passed the attributes of each successive
// option will extend the attributes instead of overwriting. There is no
// guarantee of uniqueness in the resulting attributes.
func (s *Server) WithAttributes(attributes ...attribute.KeyValue) otelTrace.SpanStartEventOption {
	return otelTrace.WithAttributes(attributes...)
}
