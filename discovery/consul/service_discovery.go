package consul

import (
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

func (s *Client) Register() error {
	check := &consulApi.AgentServiceCheck{
		Timeout:                        fmt.Sprintf("%ds", s.options.TimeOut),        // 超时时间
		Interval:                       fmt.Sprintf("%ds", s.options.IntervalTime),   // 健康检查间隔
		DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", s.options.DeregisterTime), //check失败后多少秒删除本服务，注销时间，相当于过期时间
	}
	switch s.options.CheckType {
	case "HTTP":
		check.HTTP = s.options.CheckPath
	case "TCP":
		s.options.CheckPath = fmt.Sprintf("%s:%d", s.options.CheckAddr, s.options.CheckPort)
		check.TCP = s.options.CheckPath
	case "GRPC":
		check.HTTP = s.options.CheckPath
	}

	svcReg := &consulApi.AgentServiceRegistration{
		ID:                s.options.Id,
		Name:              s.options.Name,
		Tags:              s.options.Tags,
		Port:              s.options.CheckPort,
		Address:           s.options.CheckAddr,
		EnableTagOverride: true,
		Check:             check,
		Checks:            nil,
	}
	err := s.client.Agent().ServiceRegister(svcReg)
	if err != nil {
		return errors.Wrap(err, "register service error")
	}
	return nil
}

func (s *Client) Deregister() error {
	err := s.client.Agent().ServiceDeregister(s.options.Id)
	if err != nil {
		return errors.Wrapf(err, "deregister service error[key=%s]", s.options.Id)
	}
	return nil
}
