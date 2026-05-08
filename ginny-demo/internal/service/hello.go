package service

import (
	"context"
	"fmt"
	"os"

	pb "github.com/goriller/ginny-demo/api/proto"
	"github.com/goriller/ginny-demo/internal/repo/entity"
	"github.com/goriller/ginny/errs"
	"github.com/goriller/ginny/logger"
	"github.com/goriller/gorm-plus/gplus"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

// Hello implements grpc proto Hello Method interface.
func (s *Service) Hello(ctx context.Context, req *pb.HelloReq) (*pb.HelloRes, error) {
	log := logger.WithContext(ctx)
	log.Debug("req", zap.Any("req", req))

	// topic := config.Get().Broker.Topic

	// err := s.task.Publish(ctx, topic, &broker.Message{
	// 	Header: map[string]string{},
	// 	Body:   []byte("test"),
	// })
	// if err != nil {
	// 	return nil, errs.New(codes.Canceled, "the error example for 4xx")
	// }

	switch req.Name {
	case "error":
		return nil, errs.New(codes.Code(pb.ErrorCode_CustomNotFound), "the error example for CustomNotFound")
	case "error1":
		return nil, errs.New(codes.InvalidArgument, "the error example for 4xx")
	case "panic":
		panic("the error example for panic")
	case "host":
		host, _ := os.Hostname()
		return &pb.HelloRes{
			Msg: fmt.Sprintf("hello %s form %s", req.Name, host),
		}, nil
	}
	// Demo: 自定义日志字段
	log.With(zap.String("custom2", "test2")).Info("xxx")

	_, err := s.userRepository.Insert(ctx, &entity.UserEntity{
		Name: "test1",
	})
	if err != nil {
		return nil, errs.New(codes.Internal, err.Error())
	}

	user, err := s.userRepository.Find(ctx, entity.UserEntity{
		Id: 1,
	}, nil)
	if err != nil {
		return nil, errs.New(codes.Internal, err.Error())
	}
	log.Info("user", zap.Any("user", user))

	query, model := gplus.NewQuery[entity.UserEntity]()
	query.Eq(&model.Name, "")
	ret, err := s.userRepository.SelectPage(ctx, query, 0, 2)
	if err != nil {
		return nil, errs.New(codes.Internal, err.Error())
	}
	log.Info("user", zap.Any("ret", ret))

	_, err = s.userRepository.Update(ctx, entity.UserEntity{
		Id: 1,
	}, entity.UserEntity{
		Name: "0",
	})
	if err != nil {
		return nil, errs.New(codes.Internal, err.Error())
	}

	_, err = s.userRepository.Delete(ctx, entity.UserEntity{
		Id: 1,
	})
	if err != nil {
		return nil, errs.New(codes.Internal, err.Error())
	}

	// 返回结果
	return &pb.HelloRes{
		Msg: fmt.Sprintf("hello %s ", req.Name),
		// Msg: fmt.Sprintf("hello %s ", req.Name),
	}, nil
}
