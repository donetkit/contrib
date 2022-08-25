package consul

import (
	"fmt"
	discovery2 "github.com/donetkit/contrib/pkg/discovery"
	"github.com/donetkit/contrib/utils/host"
	"github.com/donetkit/contrib/utils/uuid"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"net"
	"os"
	"time"
)

/*
GRPC

["trace.enable=true",
"prometheus.enable=true",
"traefik.enable=true",
"traefik.http.routers.[ServiceName].middlewares=request-retry@file",
"traefik.http.routers.[ServiceName].rule=PathPrefix(`/[ApiVersion]`)",
"traefik.http.routers.[ServiceName].entryPoints=web",
"traefik.http.services.[ServiceName].loadbalancer.server.scheme=h2c"
]

HTTP

["trace.enable=true",
"prometheus.enable=true",
"traefik.enable=true",
"traefik.http.routers.[ServiceName].middlewares=request-retry@file",
"traefik.http.routers.[ServiceName].rule=PathPrefix(`/[ApiVersion]/`)",
"traefik.http.routers.[ServiceName].entryPoints=web",
"traefik.http.routers.[ServiceName].priority=[ServiceTimeStamp]"
]

*/

type Client struct {
	client  *consulApi.Client
	options *discovery2.Config
}

func New(opts ...discovery2.Option) (*Client, error) {
	cfg := &discovery2.Config{
		Id:             uuid.NewUUID(),
		Name:           "Service",
		RegisterAddr:   "127.0.0.1",
		RegisterPort:   8500,
		CheckAddr:      host.GetOutBoundIp(),
		CheckPort:      80,
		Tags:           []string{"v0.0.1"},
		IntervalTime:   15,
		DeregisterTime: 15,
		TimeOut:        3,
		CheckResponse:  &discovery2.CheckResponse{RetryCount: 3},
		CheckType:      "TCP",
	}
	cfg.CheckResponse.SetHealthy("Healthy")
	cfg.HttpRouter = func(r *discovery2.CheckResponse) {}
	for _, opt := range opts {
		opt(cfg)
	}
	consulCli, err := consulApi.NewClient(&consulApi.Config{Address: fmt.Sprintf("%s:%d", cfg.RegisterAddr, cfg.RegisterPort)})
	if err != nil {
		return nil, errors.Wrap(err, "create consul client error")
	}
	consulClient := &Client{
		options: cfg,
		client:  consulCli,
	}
	consulClient.checkHealthyStatus()
	return consulClient, nil
}

// SetTags set tags []string
func (s *Client) SetTags(tags ...string) {
	s.options.Tags = tags
}

func (s *Client) checkHealthyStatus() {
	if s.options.CheckHealthyStatus {
		switch s.options.CheckType {
		case "HTTP":
			s.checkHealthyHttp()
		case "TCP":
			s.checkHealthyTCP()
		case "GRPC":
			s.checkHealthyGRPC()
		}
	}

}

func (s *Client) checkHealthyHttp() {
	go func() {
		time.Sleep(time.Second * 5)
		ticker := time.NewTicker(time.Duration(s.options.IntervalTime) * time.Second)
		for {
			select {
			case <-ticker.C:
				var timeNow = time.Now().Add(time.Duration(s.options.IntervalTime*s.options.CheckResponse.RetryCount)*time.Second*-1).UnixNano() / 1000 / 1000
				if timeNow > s.options.CheckResponse.GetOnTime() {
					ticker.Stop()
					os.Exit(3)
				}
			}
		}
	}()
}

func (s *Client) checkHealthyTCP() {
	go func() {
		time.Sleep(time.Second * 5)
		ticker := time.NewTicker(time.Duration(s.options.IntervalTime) * time.Second)
		for {
			select {
			case <-ticker.C:
				conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.options.CheckAddr, s.options.CheckPort), 3*time.Second)
				if err == nil {
					conn.Close()
					s.options.CheckResponse.Result()
				}
				var timeNow = time.Now().Add(time.Duration(s.options.IntervalTime*s.options.CheckResponse.RetryCount)*time.Second*-1).UnixNano() / 1000 / 1000
				if timeNow > s.options.CheckResponse.GetOnTime() {
					ticker.Stop()
					os.Exit(3)
				}
			}
		}
	}()
}

func (s *Client) checkHealthyGRPC() {
	go func() {
		time.Sleep(time.Second * 5)
		ticker := time.NewTicker(time.Duration(s.options.IntervalTime) * time.Second)
		for {
			select {
			case <-ticker.C:
				var timeNow = time.Now().Add(time.Duration(s.options.IntervalTime*s.options.CheckResponse.RetryCount)*time.Second*-1).UnixNano() / 1000 / 1000
				if timeNow > s.options.CheckResponse.GetOnTime() {
					ticker.Stop()
					os.Exit(3)
				}
			}
		}
	}()
}
