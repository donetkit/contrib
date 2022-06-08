package consoleserve

import (
	"github.com/donetkit/contrib-log/glog"
)

// Option for queue system
type Option func(*Server)

// WithServiceName set serviceName function
func WithServiceName(serviceName string) Option {
	return func(cfg *Server) {
		cfg.ServiceName = serviceName
	}
}

// WithLogger set logger function
func WithLogger(logger glog.ILogger) Option {
	return func(cfg *Server) {
		cfg.Logger = logger.WithField("Serve", "Serve")
	}
}

// WithVersion set version function
func WithVersion(version string) Option {
	return func(cfg *Server) {
		cfg.Version = version
	}
}

// WithProtocol set protocol function
func WithProtocol(protocol string) Option {
	return func(cfg *Server) {
		cfg.protocol = protocol
	}
}
