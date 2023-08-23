// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

const (
	service     = "http-tracer-server1"
	environment = "development" // "production" "development"
)

func initTracer() *sdktrace.TracerProvider {
	exp, _ := jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost("127.0.0.1"), jaeger.WithAgentPort(fmt.Sprintf("%d", 6831))))
	//exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(service))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func main() {
	//tp := initTracer()
	//defer func() {
	//	if err := tp.Shutdown(context.Background()); err != nil {
	//		log.Printf("Error shutting down tracer provider: %v", err)
	//	}
	//}()
	//
	//uk := attribute.Key("username")
	//
	//helloHandler := func(w http.ResponseWriter, req *http.Request) {
	//	ctx := req.Context()
	//	span := trace.SpanFromContext(ctx)
	//	bag := baggage.FromContext(ctx)
	//	span.AddEvent("handling this...", trace.WithAttributes(uk.String(bag.Member("username").Value())))
	//
	//	_, _ = io.WriteString(w, "Hello, world!\n")
	//}
	//
	//otelHandler := otel_http.NewHandler(http.HandlerFunc(helloHandler), "Hello")
	//
	//http.Handle("/hello", otelHandler)
	//err := http.ListenAndServe(":7777", nil)
	//if err != nil {
	//	panic(err)
	//}
}
