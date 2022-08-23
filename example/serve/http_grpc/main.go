package main

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib/server/http_grpc_serve"
	"google.golang.org/grpc"
	"net/http"
)

func registerHTTP(ctx context.Context, s *http_grpc_serve.Server) {
	//if err := pb.RegisterGreeterHandler(ctx, s.ServerMux, s.GRPClientConn); err != nil {
	//	panic(err)
	//}
}

func registerGRPC(ctx context.Context, s *http_grpc_serve.Server) {
	grpcServer := grpc.NewServer()
	//pb.RegisterGreeterServer(grpcServer, new(Server))
	if err := grpcServer.Serve(s.GRPCListener); err != nil {
		panic(err)
	}
}

type TestHandler struct{}

func (TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Test Handler!")
}

func main() {
	s := http_grpc_serve.New(
		http_grpc_serve.WithGRPCRegisterFunc(registerGRPC),
		http_grpc_serve.WithHTTPRegisterFunc(registerHTTP))

	s.AddHandler("/test", new(TestHandler))
	s.Run()
}
