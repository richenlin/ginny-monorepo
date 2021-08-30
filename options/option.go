package options

import (
	consul "github.com/hashicorp/consul/api"
)

// ServerOption
type ServerOption struct {
	Host   string
	Port   int
	Mode   string
	Consul *consul.Client
}

// ServerOptional
type ServerOptional func(s *ServerOption)

// WithHost
func WithHost(h string) ServerOptional {
	return func(s *ServerOption) {
		s.Host = h
	}
}

// WithPort
func WithPort(p int) ServerOptional {
	return func(s *ServerOption) {
		s.Port = p
	}
}

// WithMode
func WithMode(m string) ServerOptional {
	return func(s *ServerOption) {
		s.Mode = m
	}
}

// WithConsul
func WithConsul(c *consul.Client) ServerOptional {
	return func(s *ServerOption) {
		s.Consul = c
	}
}
