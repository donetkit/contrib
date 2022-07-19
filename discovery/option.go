package discovery

import (
	"fmt"
	"time"
)

type HttpRouter func(r *CheckResponse)

type CheckResponse struct {
	Url        string
	healthy    string
	onTime     int64
	RetryCount int
}

func (r *CheckResponse) Result() string {
	r.onTime = time.Now().UnixNano() / 1000 / 1000
	return r.healthy
}

func (r *CheckResponse) GetOnTime() int64 {
	return r.onTime
}

func (r *CheckResponse) SetHealthy(healthy string) {
	r.healthy = healthy
}

// Option for queue system
type Option func(*Config)

// WithId set id function
func WithId(id string) Option {
	return func(cfg *Config) {
		cfg.Id = id
	}
}

// WithName set name function
func WithName(name string) Option {
	return func(cfg *Config) {
		cfg.Name = name
	}
}

// WithRegisterAddr set addr function
func WithRegisterAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.RegisterAddr = addr
	}
}

// WithRegisterPort set port function
func WithRegisterPort(port int) Option {
	return func(cfg *Config) {
		cfg.RegisterPort = port
	}
}

// WithCheckAddr set addr function
func WithCheckAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.CheckAddr = addr
	}
}

// WithCheckPort set port function
func WithCheckPort(port int) Option {
	return func(cfg *Config) {
		cfg.CheckPort = port
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

// WithEnableHealthyStatus  checkHealthyStatus function
func WithEnableHealthyStatus() Option {
	return func(cfg *Config) {
		cfg.CheckHealthyStatus = true
	}
}

// WithCheckType  检查类型 HTTP TCP GRPC
func WithCheckType(checkType string) Option {
	return func(cfg *Config) {
		cfg.CheckType = checkType
	}
}

// WithCheckTCP set checkHttp function r.GET(url, func(c *gin.Context) { c.String(200, "Healthy") })
func WithCheckTCP() Option {
	return func(cfg *Config) {
		cfg.CheckType = "TCP"
		cfg.CheckPath = fmt.Sprintf("%s:%d", cfg.CheckAddr, cfg.CheckPort)
	}
}

// WithCheckHTTP set checkHttp function r.GET(url, func(c *gin.Context) { c.String(200, "Healthy") })
func WithCheckHTTP(router HttpRouter, checkHttp ...string) Option {
	return func(cfg *Config) {
		cfg.HttpRouter = router
		var url = fmt.Sprintf("/health/%s.health", cfg.Id)
		if len(checkHttp) > 0 {
			url = checkHttp[0]
		}
		cfg.CheckType = "HTTP"
		cfg.CheckPath = url
		cfg.CheckPath = fmt.Sprintf("http://%s:%d%s", cfg.CheckAddr, cfg.CheckPort, url)
		cfg.CheckResponse.Url = url
		cfg.HttpRouter(cfg.CheckResponse)
	}
}

// WithCheckGrpc set checkHttp function r.GET(url, func(c *gin.Context) { c.String(200, "Healthy") })
func WithCheckGrpc(router HttpRouter, checkPath ...string) Option {
	return func(cfg *Config) {
		cfg.HttpRouter = router
		var url = fmt.Sprintf("/health/%s.health", cfg.Id)
		if len(checkPath) > 0 {
			url = checkPath[0]
		}
		cfg.CheckType = "GRPC"
		cfg.CheckPath = url
		cfg.CheckPath = fmt.Sprintf("%v:%v/%v", cfg.CheckAddr, cfg.CheckPort, url)
		cfg.CheckResponse.Url = url
		cfg.HttpRouter(cfg.CheckResponse)
	}
}
