package consul

import (
	"fmt"
	"github.com/donetkit/contrib/discovery"
	"github.com/donetkit/contrib/utils/host"
	"github.com/donetkit/contrib/utils/uuid"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
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
		IntervalTime:        5,
		DeregisterTime:      15,
		TimeOut:             3,
	}
	cfg.CheckHTTPRouter = func(url string) {}
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
	return consulClient, nil
}

// SetTags set tags []string
func (s *Client) SetTags(tags ...string) {
	s.options.Tags = tags
}
