package http_grpc_serve

import (
	"github.com/donetkit/contrib-log/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Option for queue system
type Option func(*config)

// WithServiceName set serviceName function
func WithServiceName(serviceName string) Option {
	return func(cfg *config) {
		cfg.ServiceName = serviceName
	}
}

// WithHost set host function
func WithHost(host string) Option {
	return func(cfg *config) {
		cfg.Host = host
	}
}

// WithPort set port function
func WithPort(port int) Option {
	return func(cfg *config) {
		cfg.Port = port
	}
}

func WithHTTPRegisterFunc(registerHTTP registerFunc) Option {
	return func(cfg *config) {
		cfg.registerHTTP = registerHTTP
	}
}

func WithGRPCRegisterFunc(registerGRPC registerFunc) Option {
	return func(cfg *config) {
		cfg.registerGRPC = registerGRPC
	}
}

//
//
////WithHttpServer set httpServer function
//func WithHttpServer(httpServer http.Server) Option {
//	return func(cfg *config) {
//		cfg.httpServer = httpServer
//	}
//}
//
//// WithReadTimeout set readTimeout function
//func WithReadTimeout(readTimeout time.Duration) Option {
//	return func(cfg *config) {
//		cfg.readTimeout = readTimeout
//	}
//}
//
//// WithWriterTimeout set writerTimeout function
//func WithWriterTimeout(writerTimeout time.Duration) Option {
//	return func(cfg *config) {
//		cfg.writerTimeout = writerTimeout
//	}
//}
//
//// WithMaxHeaderBytes set maxHeaderBytes function
//func WithMaxHeaderBytes(maxHeaderBytes int) Option {
//	return func(cfg *config) {
//		cfg.maxHeaderBytes = maxHeaderBytes
//	}
//}

// WithLogger set logger function
func WithLogger(logger glog.ILogger) Option {
	return func(cfg *config) {
		cfg.Logger = logger.WithField("GrpcServe", "GrpcServe")
	}
}

// WithVersion set version function
func WithVersion(version string) Option {
	return func(cfg *config) {
		cfg.Version = version
	}
}

// WithProtocol set protocol function
func WithProtocol(protocol string) Option {
	return func(cfg *config) {
		cfg.protocol = protocol
	}
}

// WithGrpcServerOptions set grpc.ServerOption function
func WithGrpcServerOptions(grpcOpts ...grpc.ServerOption) Option {
	return func(cfg *config) {
		for _, grpcOpt := range grpcOpts {
			cfg.grpcOpts = append(cfg.grpcOpts, grpcOpt)
		}
	}
}

// WithGrpcServerOption set grpc.ServerOption function
func WithGrpcServerOption(grpcOpt grpc.ServerOption) Option {
	return func(cfg *config) {
		if grpcOpt != nil {
			cfg.grpcOpts = append(cfg.grpcOpts, grpcOpt)
		}
	}
}

// WithTransportCredentials set credentials.TransportCredentials function
func WithTransportCredentials(credentials credentials.TransportCredentials) Option {
	return func(cfg *config) {
		if credentials != nil {
			cfg.credentials = credentials
		}
	}
}
