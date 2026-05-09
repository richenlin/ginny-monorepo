// Package service implements the GreeterService ConnectRPC handler.
package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	helloworldv1 "github.com/goriller/ginny-demo/v2/gen/helloworld/v1"
	"github.com/goriller/ginny/v2/errs"
)

// GreeterService implements the GreeterServiceHandler interface.
type GreeterService struct{}

// NewGreeterService creates a new GreeterService.
func NewGreeterService() *GreeterService {
	return &GreeterService{}
}

// SayHello returns a greeting for the given name.
func (s *GreeterService) SayHello(ctx context.Context, req *connect.Request[helloworldv1.SayHelloRequest]) (*connect.Response[helloworldv1.SayHelloResponse], error) {
	name := req.Msg.GetName()
	if name == "" {
		return nil, errs.New(connect.CodeInvalidArgument, 40001, "name is required")
	}
	if name == "error" {
		return nil, errs.New(connect.CodeNotFound, 40401, "user not found")
	}
	if name == "panic" {
		panic("intentional panic for recovery interceptor demo")
	}
	return connect.NewResponse(&helloworldv1.SayHelloResponse{
		Message: fmt.Sprintf("Hello, %s! Welcome to Ginny v2.", name),
	}), nil
}

// ServerStreamGreetings streams greetings for the given name.
func (s *GreeterService) ServerStreamGreetings(ctx context.Context, req *connect.Request[helloworldv1.SayHelloRequest], stream *connect.ServerStream[helloworldv1.SayHelloResponse]) error {
	name := req.Msg.GetName()
	if name == "" {
		return errs.New(connect.CodeInvalidArgument, 40001, "name is required")
	}
	for i := 1; i <= 3; i++ {
		if err := stream.Send(&helloworldv1.SayHelloResponse{
			Message: fmt.Sprintf("Hello #%d, %s!", i, name),
		}); err != nil {
			return err
		}
	}
	return nil
}
