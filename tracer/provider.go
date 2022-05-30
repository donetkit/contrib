package tracer

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func NewTracerProvider(service, host, development string, port int, sampler trace.Sampler) (*trace.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(host), jaeger.WithAgentPort(fmt.Sprintf("%d", port))))
	//exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(sampler),
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", development),
			//attribute.Int64("ID", id),
		)),
	)
	return tracerProvider, nil
}
