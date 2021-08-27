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
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	consul "github.com/hashicorp/consul/api"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Options
type Options struct {
	Host string
	Port int
}

// NewServerOptions
func NewServerOptions(v *viper.Viper) (*Options, error) {
	var (
		err error
		o   = new(Options)
	)
	if err = v.UnmarshalKey("grpc", o); err != nil {
		return nil, err
	}

	return o, nil
}

// Server
type Server struct {
	o         *Options
	appName   string
	host      string
	port      int
	logger    *zap.Logger
	server    *grpc.Server
	consulCli *consul.Client
}

// InitServers
type InitServers func(s *grpc.Server)

// NewServer
func NewServer(o *Options, logger *zap.Logger, tracer opentracing.Tracer, init InitServers) (*Server, error) {
	// initialize grpc server
	var gs *grpc.Server
	logger = logger.With(zap.String("type", "grpc"))

	{
		grpc_prometheus.EnableHandlingTimeHistogram()
		gs = grpc.NewServer(
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_prometheus.StreamServerInterceptor,
				grpc_zap.StreamServerInterceptor(logger),
				grpc_recovery.StreamServerInterceptor(),
				otgrpc.OpenTracingStreamServerInterceptor(tracer),
			)),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_prometheus.UnaryServerInterceptor,
				grpc_zap.UnaryServerInterceptor(logger),
				grpc_recovery.UnaryServerInterceptor(),
				otgrpc.OpenTracingServerInterceptor(tracer),
			)),
		)
		init(gs)
	}

	s := &Server{
		o:      o,
		logger: logger.With(zap.String("type", "grpc.Server")),
		server: gs,
	}

	return s, nil
}

// AppName
func (s *Server) AppName(name string) {
	s.appName = name
}

// ConsulClient
func (s *Server) ConsulClient(cli *consul.Client) {
	s.consulCli = cli
}

// Start
func (s *Server) Start(opt *options.ServerOption) error {
	s.port = s.o.Port
	if s.port == 0 {
		s.port = util.GetAvailablePort()
	}

	// s.host = util.GetLocalIP4()
	s.host = s.o.Host

	if s.host == "" {
		return errors.New("get local ipv4 error")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	s.logger.Info("grpc server starting ...", zap.String("addr", addr))
	go func() {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		if err := s.server.Serve(lis); err != nil {
			s.logger.Fatal("failed to serve: %v", zap.Error(err))
		}
	}()

	if opt.Consul != nil {
		s.consulCli = opt.Consul
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
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	for key, _ := range s.server.GetServiceInfo() {
		check := &consul.AgentServiceCheck{
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "60m",
			TCP:                            addr,
		}

		id := fmt.Sprintf("%s[%s:%d]", key, s.host, s.port)

		svcReg := &consul.AgentServiceRegistration{
			ID:                id,
			Name:              s.appName + "_" + key,
			Tags:              []string{"grpc"},
			Port:              s.port,
			Address:           s.host,
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
		id := fmt.Sprintf("%s[%s:%d]", key, s.host, s.port)

		err := s.consulCli.Agent().ServiceDeregister(id)
		if err != nil {
			return errors.Wrapf(err, "deregister service error[id=%s]", id)
		}
		s.logger.Info("deregister service success ", zap.String("id", id))
	}

	return nil
}
