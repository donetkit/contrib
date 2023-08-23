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

const (
	service     = "http-tracer-client"
	environment = "development" // "production" "development"
)

func main() {
	//var traceServer *tracer.Server
	//tp, err := tracer.NewTracerProvider(service, "127.0.0.1", environment, 6831, nil)
	//if err == nil {
	//	jaeger := tracer.Jaeger{}
	//	traceServer = tracer.New(tracer.WithName(service), tracer.WithProvider(tp), tracer.WithPropagators(jaeger))
	//
	//	defer func() {
	//		tp.Shutdown(context.Background())
	//	}()
	//}
	//
	//url := flag.String("server", "http://localhost:7777/hello", "server url")
	//flag.Parse()
	//
	//client := http.Client{Transport: otel_http.NewTransport(http.DefaultTransport)}
	//
	//bag, _ := baggage.Parse("username=donuts")
	//ctx := baggage.ContextWithBaggage(context.Background(), bag)
	//
	//var body []byte
	//
	//tr := traceServer.Tracer
	//err = func(ctx context.Context) error {
	//	ctx, span := tr.Start(ctx, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	//	defer span.End()
	//
	//	req, _ := http.NewRequestWithContext(ctx, "GET", *url, nil)
	//	traceServer.Propagators.Inject(ctx, propagation.HeaderCarrier(req.Header))
	//	fmt.Printf("Sending request...\n")
	//	res, err := client.Do(req)
	//	if err != nil {
	//		panic(err)
	//	}
	//	body, err = ioutil.ReadAll(res.Body)
	//	_ = res.Body.Close()
	//
	//	return err
	//}(ctx)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Printf("Response Received: %s\n\n\n", body)
	//fmt.Printf("Waiting for few seconds to export spans ...\n\n")
	//time.Sleep(10 * time.Second)
	//fmt.Printf("Inspect traces on stdout\n")
}
