package consul

import (
	"fmt"
	"github.com/donetkit/contrib/discovery"
	"github.com/donetkit/contrib/utils/host"
	"github.com/donetkit/contrib/utils/uuid"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"os"
	"time"
)

type Client struct {
	client  *consulApi.Client
	options *discovery.Config
}

func New(opts ...discovery.Option) (*Client, error) {
	cfg := &discovery.Config{
		Id:                  uuid.NewUUID(),
		ServiceName:         "Service",
		ServiceRegisterAddr: "127.0.0.1",
		ServiceRegisterPort: 8500,
		ServiceCheckAddr:    host.GetOutBoundIp(),
		ServiceCheckPort:    80,
		Tags:                []string{"v0.0.1"},
		IntervalTime:        15,
		DeregisterTime:      15,
		TimeOut:             3,
		OnTime:              time.Now().UnixNano() / 1000 / 1000,
		RetryCount:          3,
	}
	cfg.Router = func(url string, fn discovery.UpdateServerTime) {}
	cfg.UpdateTime = func() {
		if cfg.CheckOnLine {
			cfg.OnTime = time.Now().UnixNano() / 1000 / 1000
		}
	}
	for _, opt := range opts {
		opt(cfg)
	}
	consulCli, err := consulApi.NewClient(&consulApi.Config{Address: fmt.Sprintf("%s:%d", cfg.ServiceRegisterAddr, cfg.ServiceRegisterPort)})
	if err != nil {
		return nil, errors.Wrap(err, "create consul client error")
	}
	consulClient := &Client{
		options: cfg,
		client:  consulCli,
	}
	consulClient.checkServerOnLine()
	return consulClient, nil
}

// SetTags set tags []string
func (s *Client) SetTags(tags ...string) {
	s.options.Tags = tags
}

func (s *Client) checkServerOnLine() {
	if s.options.CheckOnLine {
		go func() {
			time.Sleep(time.Second * 5)
			ticker := time.NewTicker(time.Duration(s.options.IntervalTime) * time.Second)
			for {
				select {
				case <-ticker.C:
					var timeNow = time.Now().Add(time.Duration(s.options.IntervalTime*s.options.RetryCount)*time.Second*-1).UnixNano() / 1000 / 1000
					if timeNow > s.options.OnTime {
						os.Exit(3)
					}
				}
			}
		}()
	}
}
