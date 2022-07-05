package tracer

import (
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"math"
	"net/url"
	"time"
)

const (
	defaultPollingInterval = 300
)

type config struct {
	endpoint                     url.URL
	samplingRulesPollingInterval time.Duration
	logger                       glog.ILogger
}

// RemoteSamplerOption sets configuration on the sampler.
type RemoteSamplerOption interface {
	apply(*config) *config
}

type optionRemoteSamplerFunc func(*config) *config

func (f optionRemoteSamplerFunc) apply(cfg *config) *config {
	return f(cfg)
}

// WithEndpoint sets custom proxy endpoint.
// If this option is not provided the default endpoint used will be http://127.0.0.1:2000.
func WithEndpoint(endpoint url.URL) RemoteSamplerOption {
	return optionRemoteSamplerFunc(func(cfg *config) *config {
		cfg.endpoint = endpoint
		return cfg
	})
}

// WithSamplingRulesPollingInterval sets polling interval for sampling rules.
// If this option is not provided the default samplingRulesPollingInterval used will be 300 seconds.
func WithSamplingRulesPollingInterval(polingInterval time.Duration) RemoteSamplerOption {
	return optionRemoteSamplerFunc(func(cfg *config) *config {
		cfg.samplingRulesPollingInterval = polingInterval
		return cfg
	})
}

// WithLogger sets custom logging for remote sampling implementation.
// If this option is not provided the default logger used will be go-logr/stdr (https://github.com/go-logr/stdr).
func WithLogger(l glog.ILogger) RemoteSamplerOption {
	return optionRemoteSamplerFunc(func(cfg *config) *config {
		cfg.logger = l
		return cfg
	})
}

func newConfig(opts ...RemoteSamplerOption) (*config, error) {
	defaultProxyEndpoint, err := url.Parse("http://127.0.0.1:2000")
	if err != nil {
		return nil, err
	}

	cfg := &config{
		endpoint:                     *defaultProxyEndpoint,
		samplingRulesPollingInterval: defaultPollingInterval * time.Second,
	}

	for _, option := range opts {
		option.apply(cfg)
	}

	if math.Signbit(float64(cfg.samplingRulesPollingInterval)) {
		return nil, fmt.Errorf("config validation error: samplingRulesPollingInterval should be positive number")
	}

	return cfg, nil
}
