package service

import (
	"context"

	pb "github.com/goriller/ginny-demo/api/proto"

	"github.com/goriller/ginny/errs"
	"github.com/goriller/ginny/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

// RpcCli implements grpc proto RpcCli Method interface.
func (s *Service) RpcCli(ctx context.Context, req *pb.RpcCliReq) (*pb.RpcCliRes, error) {
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
	return &pb.RpcCliRes{}, nil
}
