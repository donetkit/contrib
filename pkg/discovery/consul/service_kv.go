package consul

import (
	consulApi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

func (s *Client) Get(key string) ([]byte, error) {
	kv, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, errors.New("not found value")
	}
	return kv.Value, nil
}

func (s *Client) Set(key string, value string) error {
	p := &consulApi.KVPair{Key: key, Value: []byte(value)}
	if _, err := s.client.KV().Put(p, nil); err != nil {
		return err
	}
	return nil
}

func (s *Client) Delete(key string) error {
	if _, err := s.client.KV().Delete(key, nil); err != nil {
		return err
	}
	return nil
}

func (s *Client) List(key string) (map[string][]byte, error) {
	p, _, err := s.client.KV().List(key, nil)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("not found value")
	}
	values := make(map[string][]byte, len(p))
	for _, v := range p {
		values[v.Key] = v.Value
	}
	return values, nil

}
