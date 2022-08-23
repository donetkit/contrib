package http_grpc_serve

import (
	"github.com/donetkit/contrib-log/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

// Option for queue system
type Option func(*Server)

// WithServiceName set serviceName function
func WithServiceName(serviceName string) Option {
	return func(s *Server) {
		s.ServiceName = serviceName
	}
}

// WithHost set host function
func WithHost(host string) Option {
	return func(s *Server) {
		s.Host = host
	}
}

// WithPort set port function
func WithPort(port int) Option {
	return func(s *Server) {
		s.Port = port
	}
}

func WithHTTPRegisterFunc(registerHTTP registerFunc) Option {
	return func(s *Server) {
		s.registerHTTP = registerHTTP
	}
}

func WithGRPCRegisterFunc(registerGRPC registerFunc) Option {
	return func(s *Server) {
		s.registerGRPC = registerGRPC
	}
}

//WithHttpServer set httpServer function
//func WithHttpServer(httpServer http.Server) Option {
//	return func(s *Server) {
//		s.httpServer = httpServer
//	}
//}

// WithReadTimeout set readTimeout function
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}
}

// WithWriterTimeout set writerTimeout function
func WithWriterTimeout(writerTimeout time.Duration) Option {
	return func(s *Server) {
		s.writerTimeout = writerTimeout
	}
}

// WithIdleTimeout set idleTimeout function
func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(s *Server) {
		s.idleTimeout = idleTimeout
	}
}

//// WithMaxHeaderBytes set maxHeaderBytes function
//func WithMaxHeaderBytes(maxHeaderBytes int) Option {
//	return func(s *Server) {
//		s.maxHeaderBytes = maxHeaderBytes
//	}
//}

// WithLogger set logger function
func WithLogger(logger glog.ILogger) Option {
	return func(s *Server) {
		s.Logger = logger.WithField("GrpcServe", "GrpcServe")
	}
}

// WithVersion set version function
func WithVersion(version string) Option {
	return func(s *Server) {
		s.Version = version
	}
}

// WithProtocol set protocol function
func WithProtocol(protocol string) Option {
	return func(s *Server) {
		s.protocol = protocol
	}
}

// WithGrpcServerOptions set grpc.ServerOption function
func WithGrpcServerOptions(grpcOpts ...grpc.ServerOption) Option {
	return func(s *Server) {
		for _, grpcOpt := range grpcOpts {
			s.grpcOpts = append(s.grpcOpts, grpcOpt)
		}
	}
}

// WithGrpcServerOption set grpc.ServerOption function
func WithGrpcServerOption(grpcOpt grpc.ServerOption) Option {
	return func(s *Server) {
		if grpcOpt != nil {
			s.grpcOpts = append(s.grpcOpts, grpcOpt)
		}
	}
}

// WithTransportCredentials set credentials.TransportCredentials function
func WithTransportCredentials(credentials credentials.TransportCredentials) Option {
	return func(s *Server) {
		if credentials != nil {
			s.credentials = credentials
		}
	}
}
