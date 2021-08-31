package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	prometheus "github.com/gorillazer/ginny-prometheus"
	"github.com/gorillazer/ginny-serve/options"
	util "github.com/gorillazer/ginny-util"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/opentracing-contrib/go-gin/ginhttp"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

	if err = v.UnmarshalKey("http", o); err != nil {
		return nil, err
	}

	s.Host = o.Host
	s.Port = o.Port
	s.Mode = o.Mode
	return s, nil
}

// Server
type Server struct {
	appName   string
	option    *ServerOption
	logger    *zap.Logger
	router    *gin.Engine
	server    http.Server
	consulCli *consulApi.Client
}

// InitHandlers
type InitHandlers func(r *gin.Engine)

// NewRouter
func NewRouter(o *ServerOption, logger *zap.Logger, tracer opentracing.Tracer, init InitHandlers) *gin.Engine {
	// 配置gin
	gin.SetMode(o.Mode)
	r := gin.New()
	// panic之后自动恢复
	r.Use(gin.Recovery())
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(ginhttp.Middleware(tracer))
	// 添加prometheus 监控
	r.Use(prometheus.New(r).Middleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	pprof.Register(r)

	init(r)

	return r
}

// NewServer
func NewServer(o *ServerOption, logger *zap.Logger, router *gin.Engine) (*Server, error) {
	var s = &Server{
		logger: logger.With(zap.String("type", "http")),
		router: router,
		option: o,
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
	}
	//
	if o.Host == "" {
		o.Host = util.GetLocalIP4()
		// return errors.New("get local ipv4 error")
	}

	addr := fmt.Sprintf("%s:%d", o.Host, o.Port)
	s.server = http.Server{Addr: addr, Handler: s.router}

	s.logger.Info("http server starting ...", zap.String("addr", addr))
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("start http server err", zap.Error(err))
			return
		}
	}()
	if o.Consul != nil {
		s.consulCli = o.Consul
	}

	if err := s.register(); err != nil {
		return errors.Wrap(err, "register http server error")
	}

	return nil
}

// Stop
func (s *Server) Stop() error {
	s.logger.Info("http server stopping ...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // 平滑关闭,等待5秒钟处理
	defer cancel()
	if err := s.deRegister(); err != nil {
		return errors.Wrap(err, "deregister http server error")
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "shutdown http server error")
	}

	return nil
}

// register
func (s *Server) register() error {
	if s.consulCli == nil {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", s.option.Host, s.option.Port)

	check := &consulApi.AgentServiceCheck{
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "60m",
		TCP:                            addr,
	}

	id := fmt.Sprintf("%s[%s:%d]", s.appName, s.option.Host, s.option.Port)

	svcReg := &consulApi.AgentServiceRegistration{
		ID:                id,
		Name:              string(s.appName),
		Tags:              []string{"http"},
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
	s.logger.Info("register http server success", zap.String("id", id))

	return nil
}

// deRegister
func (s *Server) deRegister() error {
	if s.consulCli == nil {
		return nil
	}
	id := fmt.Sprintf("%s[%s:%d]", s.appName, s.option.Host, s.option.Port)

	err := s.consulCli.Agent().ServiceDeregister(id)
	if err != nil {
		return errors.Wrapf(err, "deregister service error[key=%s]", id)
	}
	s.logger.Info("deregister service success ", zap.String("service", id))

	return nil
}
