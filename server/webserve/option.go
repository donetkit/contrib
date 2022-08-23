package webserve

import (
	"github.com/donetkit/contrib-log/glog"
	"net/http"
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

// WithHandler set handler function
func WithHandler(handler http.Handler) Option {
	return func(s *Server) {
		s.handler = handler
	}
}

//WithHttpServer set httpServer function
func WithHttpServer(httpServer http.Server) Option {
	return func(s *Server) {
		s.httpServer = httpServer
	}
}

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

// WithMaxHeaderBytes set maxHeaderBytes function
func WithMaxHeaderBytes(maxHeaderBytes int) Option {
	return func(s *Server) {
		s.maxHeaderBytes = maxHeaderBytes
	}
}

// WithLogger set logger function
func WithLogger(logger glog.ILogger) Option {
	return func(s *Server) {
		s.Logger = logger.WithField("WebServe", "WebServe")
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
