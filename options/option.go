package options

import (
	consul "github.com/gorillazer/ginny-consul"
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
