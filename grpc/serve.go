package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/gorillazer/ginny-serve/options"
	util "github.com/gorillazer/ginny-util"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	otgrpc "github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ServerOption
type ServerOption struct {
	options.ServerOption
}

// NewOptions
func NewOptions(v *viper.Viper) (*ServerOption, error) {
	var (
		err error
		o   = new(options.ServerOption)
		s   = new(ServerOption)
	)
	if err = v.UnmarshalKey("grpc", o); err != nil {
		return nil, err
	}

	s.Host = o.Host
	s.Port = o.Port
	return s, nil
}

// Server
type Server struct {
	appName   string
	option    *ServerOption
	logger    *zap.Logger
	server    *grpc.Server
	consulCli *consulApi.Client
}

// InitServers
type InitServers func(s *grpc.Server)

// NewServer
func NewServer(o *ServerOption, logger *zap.Logger, tracer opentracing.Tracer, init InitServers) (*Server, error) {
	// initialize grpc server
	var gs *grpc.Server
	logger = logger.With(zap.String("type", "grpc"))

	{
		grpc_prometheus.EnableHandlingTimeHistogram()
		gs = grpc.NewServer(
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				grpc_recovery.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(logger),
				grpc_ctxtags.StreamServerInterceptor(),
				otgrpc.OpenTracingStreamServerInterceptor(tracer),
				grpc_prometheus.StreamServerInterceptor,
				grpc_validator.StreamServerInterceptor(),
			)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(),
				grpc_zap.UnaryServerInterceptor(logger),
				grpc_ctxtags.UnaryServerInterceptor(),
				otgrpc.OpenTracingServerInterceptor(tracer),
				grpc_prometheus.UnaryServerInterceptor,
				grpc_validator.UnaryServerInterceptor(),
			)),
		)
		init(gs)
	}

	s := &Server{
		option: o,
		logger: logger,
		server: gs,
	}

	return s, nil
}

// AppName
func (s *Server) AppName(name string) {
	s.appName = name
}

// ConsulClient
func (s *Server) ConsulClient(cli *consulApi.Client) {
	s.consulCli = cli
}

// Start
func (s *Server) Start(opts ...options.ServerOptional) error {
	o := &options.ServerOption{
		Host: s.option.Host,
		Port: s.option.Port,
		Mode: s.option.Mode,
	}
	for _, opt := range opts {
		opt(o)
	}

	if o.Port == 0 {
		o.Port = util.GetAvailablePort()
		s.option.Port = o.Port
	}
	//
	if o.Host == "" {
		o.Host = util.GetLocalIP4()
		s.option.Host = o.Host
		// return errors.New("get local ipv4 error")
	}

	addr := fmt.Sprintf("%s:%d", o.Host, o.Port)

	s.logger.Info("grpc server starting ...", zap.String("addr", addr))
	go func() {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		if err := s.server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", zap.Error(err))
		}
	}()

	if o.Consul != nil {
		s.consulCli = o.Consul
	}
	if err := s.register(); err != nil {
		return errors.Wrap(err, "register grpc server error")
	}

	return nil
}

// Stop
func (s *Server) Stop() error {
	s.logger.Info("grpc server stopping ...")
	if err := s.deRegister(); err != nil {
		return errors.Wrap(err, "deregister grpc server error")
	}

	s.server.GracefulStop()
	return nil
}

// register
func (s *Server) register() error {
	if s.consulCli == nil {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", s.option.Host, s.option.Port)

	for key, _ := range s.server.GetServiceInfo() {
		check := &consulApi.AgentServiceCheck{
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "60m",
			TCP:                            addr,
		}

		id := fmt.Sprintf("%s[%s:%d]", key, s.option.Host, s.option.Port)

		svcReg := &consulApi.AgentServiceRegistration{
			ID:                id,
			Name:              key,
			Tags:              []string{"grpc"},
			Port:              s.option.Port,
			Address:           s.option.Host,
			EnableTagOverride: true,
			Check:             check,
			Checks:            nil,
		}

		err := s.consulCli.Agent().ServiceRegister(svcReg)
		if err != nil {
			return errors.Wrap(err, "register service error")
		}
		s.logger.Info("register grpc service success", zap.String("id", id))
	}

	return nil
}

// deRegister
func (s *Server) deRegister() error {
	if s.consulCli == nil {
		return nil
	}
	for key, _ := range s.server.GetServiceInfo() {
		id := fmt.Sprintf("%s[%s:%d]", key, s.option.Host, s.option.Port)

		err := s.consulCli.Agent().ServiceDeregister(id)
		if err != nil {
			return errors.Wrapf(err, "deregister service error[id=%s]", id)
		}
		s.logger.Info("deregister service success ", zap.String("id", id))
	}

	return nil
}
