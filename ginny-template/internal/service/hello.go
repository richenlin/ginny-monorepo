package service

import (
	"context"
	"fmt"
	"os"

	pb "MODULE_NAME/api/proto"

	"github.com/goriller/ginny/errs"
	"github.com/goriller/ginny/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

// Hello implements grpc proto Hello Method interface.
func (s *Service) Hello(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log := logger.WithContext(ctx).With(zap.String("action", "Hello"))
	log.Debug("req", zap.Any("req", req))
	switch req.Name {
	case "error":
		return nil, errs.New(codes.Code(pb.ErrorCode_CustomNotFound), "the error example for CustomNotFound")
	case "error1":
		return nil, errs.New(codes.InvalidArgument, "the error example for 4xx")
	case "panic":
		panic("the error example for panic")
	case "host":
		host, _ := os.Hostname()
		return &pb.Response{
			Msg: fmt.Sprintf("hello %s form %s", req.Name, host),
		}, nil
	}
	// Demo: 自定义日志字段
	log.With(zap.String("custom2", "test2")).Info("xxx")

	// 返回结果
	return &pb.Response{
		Msg: fmt.Sprintf("hello %s ", req.Name),
		// Msg: fmt.Sprintf("hello %s ", req.Name),
	}, nil
}
