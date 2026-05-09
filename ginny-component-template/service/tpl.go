package service

const Tpl = `package service

import (
	"context"

	"connectrpc.com/connect"
)

// {{ .Name }}Service implements the {{ .Name }} ConnectRPC service.
type {{ .Name }}Service struct{}

// New{{ .Name }}Service creates a new {{ .Name }}Service.
func New{{ .Name }}Service() *{{ .Name }}Service {
	return &{{ .Name }}Service{}
}

// Example handler — replace with your generated handler interface.
func (s *{{ .Name }}Service) Example(
	ctx context.Context,
	req *connect.Request[any],
) (*connect.Response[any], error) {
	return connect.NewResponse(nil), nil
}
`
