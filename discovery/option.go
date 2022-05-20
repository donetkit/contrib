package discovery

import (
	"fmt"
)

type Router func(url string, fn UpdateServerTime)

type UpdateServerTime func()

// Option for queue system
type Option func(*Config)

// WithId set id function
func WithId(id string) Option {
	return func(cfg *Config) {
		cfg.Id = id
	}
}

// WithServiceName set serviceName function
func WithServiceName(serviceName string) Option {
	return func(cfg *Config) {
		cfg.ServiceName = serviceName
	}
}

// WithServiceRegisterAddr set serviceRegisterAddr function
func WithServiceRegisterAddr(serviceRegisterAddr string) Option {
	return func(cfg *Config) {
		cfg.ServiceRegisterAddr = serviceRegisterAddr
	}
}

// WithServiceRegisterPort set serviceRegisterPort function
func WithServiceRegisterPort(serviceRegisterPort int) Option {
	return func(cfg *Config) {
		cfg.ServiceRegisterPort = serviceRegisterPort
	}
}

// WithServiceCheckAddr set serviceCheckAddr function
func WithServiceCheckAddr(serviceCheckAddr string) Option {
	return func(cfg *Config) {
		cfg.ServiceCheckAddr = serviceCheckAddr
	}
}

// WithServiceCheckPort set serviceCheckPort function
func WithServiceCheckPort(serviceCheckPort int) Option {
	return func(cfg *Config) {
		cfg.ServiceCheckPort = serviceCheckPort
	}
}

// WithTags set tags function
func WithTags(tags ...string) Option {
	return func(cfg *Config) {
		cfg.Tags = tags
	}
}

// WithIntervalTime set intervalTime function
func WithIntervalTime(intervalTime int) Option {
	return func(cfg *Config) {
		if intervalTime <= 0 {
			intervalTime = 15
		}
		cfg.IntervalTime = intervalTime
	}
}

// WithDeregisterTime set deregisterTime function
func WithDeregisterTime(deregisterTime int) Option {
	return func(cfg *Config) {
		if deregisterTime <= 0 {
			deregisterTime = 15
		}
		cfg.DeregisterTime = deregisterTime
	}
}

// WithTimeOut set timeOut function
func WithTimeOut(timeOut int) Option {
	return func(cfg *Config) {
		if timeOut <= 0 {
			timeOut = 3
		}
		cfg.TimeOut = timeOut
	}
}

// WithCheckOnLine set checkOnLine function
func WithCheckOnLine(checkOnLine bool) Option {
	return func(cfg *Config) {
		cfg.CheckOnLine = checkOnLine
	}
}

// WithRetryCount set retryCount function
func WithRetryCount(retryCount int) Option {
	return func(cfg *Config) {
		if retryCount <= 0 {
			retryCount = 3
		}
		cfg.RetryCount = retryCount
	}
}

// WithCheckHTTP set checkHttp function r.GET(url, func(c *gin.Context) { c.String(200, "Healthy") })
func WithCheckHTTP(router Router, checkHttp ...string) Option {
	return func(cfg *Config) {
		cfg.Router = router
		var checkHttpUrl = fmt.Sprintf("/health/%s.health", cfg.Id)
		if len(checkHttp) > 0 {
			checkHttpUrl = checkHttp[0]
		}
		cfg.CheckHTTP = checkHttpUrl
		cfg.CheckHTTP = fmt.Sprintf("http://%s:%d%s", cfg.ServiceCheckAddr, cfg.ServiceCheckPort, checkHttpUrl)
		cfg.Router(checkHttpUrl, cfg.UpdateTime)
	}
}
