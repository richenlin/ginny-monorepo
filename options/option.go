package options

import (
	consul "github.com/hashicorp/consul/api"
)

// ServerOption
type ServerOption struct {
	Consul *consul.Client
}

// ServerOptional
type ServerOptional func(s *ServerOption)

// WithConsul
func WithConsul(c *consul.Client) ServerOptional {
	return func(s *ServerOption) {
		s.Consul = c
	}
}
