package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/gorillazer/ginny/util"
	consul "github.com/hashicorp/consul/api"
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Options
type Options struct {
	Host string
	Port int
	Mode string
}

// Server
type Server struct {
	o          *Options
	app        string
	host       string
	port       int
	logger     *zap.Logger
	router     *gin.Engine
	httpServer http.Server
	consulCli  *consul.Client
}

// NewOptions
func NewOptions(v *viper.Viper) (*Options, error) {
	var (
		err error
		o   = new(Options)
	)

	if err = v.UnmarshalKey("http", o); err != nil {
		return nil, err
	}

	return o, err
}

// InitServers
type InitServers func(r *gin.Engine)

// NewRouter
func NewRouter(o *Options, logger *zap.Logger, init InitServers, tracer opentracing.Tracer, middleware ...gin.HandlerFunc) *gin.Engine {
	// 配置gin
	gin.SetMode(o.Mode)
	r := gin.New()
	// panic之后自动恢复
	r.Use(gin.Recovery())
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(ginhttp.Middleware(tracer))
	r.Use(middleware...)

	pprof.Register(r)

	init(r)

	return r
}

// NewServer
func NewServer(o *Options, logger *zap.Logger, router *gin.Engine, consulCli ...*consul.Client) (*Server, error) {
	var s = &Server{
		logger: logger.With(zap.String("type", "http.Server")),
		router: router,
		// consulCli: consulCli,
		o: o,
	}
	// consul
	if len(consulCli) > 0 && consulCli[0] != nil {
		s.consulCli = consulCli[0]
	}

	return s, nil
}

// Application
func (s *Server) Application(name string) {
	s.app = name
}

// Start
func (s *Server) Start() error {
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

	s.httpServer = http.Server{Addr: addr, Handler: s.router}

	s.logger.Info("http server starting ...", zap.String("addr", addr))
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("start http server err", zap.Error(err))
			return
		}
	}()

	if s.consulCli != nil {
		if err := s.register(); err != nil {
			return errors.Wrap(err, "register http server error")
		}
	}

	return nil
}

// Stop
func (s *Server) Stop() error {
	s.logger.Info("stopping http server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // 平滑关闭,等待5秒钟处理
	defer cancel()
	if s.consulCli != nil {
		if err := s.deRegister(); err != nil {
			return errors.Wrap(err, "deregister http server error")
		}
	}

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "shutdown http server error")
	}

	return nil
}

// register
func (s *Server) register() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	check := &consul.AgentServiceCheck{
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "60m",
		TCP:                            addr,
	}

	id := fmt.Sprintf("%s[%s:%d]", s.app, s.host, s.port)

	svcReg := &consul.AgentServiceRegistration{
		ID:                id,
		Name:              string(s.app),
		Tags:              []string{"http"},
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
	s.logger.Info("register http server success", zap.String("id", id))

	return nil
}

// deRegister
func (s *Server) deRegister() error {
	id := fmt.Sprintf("%s[%s:%d]", s.app, s.host, s.port)

	err := s.consulCli.Agent().ServiceDeregister(id)
	if err != nil {
		return errors.Wrapf(err, "deregister service error[key=%s]", id)
	}
	s.logger.Info("deregister service success ", zap.String("service", id))

	return nil
}

var ProviderSet = wire.NewSet(NewOptions, NewRouter, NewServer)
