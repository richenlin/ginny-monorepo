package service

import (
	"context"

	pb "MODULE_NAME/api/proto"

	"github.com/goriller/ginny/errs"
	"github.com/goriller/ginny/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

// METHOD_NAME implements grpc proto METHOD_NAME Method interface.
func (s *Service) METHOD_NAME(ctx context.Context, req *pb.METHOD_REQNAME) (*pb.METHOD_RESNAME, error) {
	log := logger.WithContext(ctx).With(zap.String("action", "Hello"))
	log.Debug("req", zap.Any("req", req))

	if req == nil {
		return nil, errs.New(codes.Code(pb.ErrorCode_CustomNotFound), "the error example for CustomNotFound")
	}
	if req.Name == "" {
		return nil, errs.New(codes.InvalidArgument, "the error example for 4xx")
	}

	// Demo: 自定义日志字段
	log.With(zap.String("custom2", "test2")).Info("xxx")

	// 返回结果
	return &pb.METHOD_RESNAME{}, nil
}
