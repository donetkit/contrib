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
	"context"
	"github.com/donetkit/contrib/tracer"
	"github.com/donetkit/contrib/utils/otel_http"
	"go.opentelemetry.io/otel/propagation"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
)

const (
	service     = "http-tracer-server"
	environment = "development" // "production" "development"
)

func main() {
	var traceServer *tracer.Server
	tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, nil)
	if err == nil {
		jaeger := tracer.Jaeger{}
		traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))

		defer func() {
			tp.Shutdown(context.Background())
		}()
	}

	uk := attribute.Key("username")

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		ctx = traceServer.Propagators.Extract(ctx, propagation.HeaderCarrier(req.Header))
		ctx, span := traceServer.Tracer.Start(ctx, req.Method)
		if !span.IsRecording() {
			return
		}
		defer span.End()

		bag := traceServer.FromContext(ctx)
		span.AddEvent("handling this...", traceServer.WithAttributes(uk.String(bag.Member("username").Value())))

		_, _ = io.WriteString(w, "Hello, world!\n")
	}

	otelHandler := otel_http.NewHandler(http.HandlerFunc(helloHandler), "Hello")

	http.Handle("/hello", otelHandler)
	err = http.ListenAndServe(":7777", nil)
	if err != nil {
		panic(err)
	}
}
