module github.com/gorillazer/ginny-serve

go 1.16

require (
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-contrib/zap v0.0.1
	github.com/gin-gonic/gin v1.7.4
	github.com/google/wire v0.5.0
	github.com/gorillazer/ginny-util v0.0.0-20210824061306-90e4d3ae4237
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/consul/api v1.9.1
	github.com/mbobakov/grpc-consul-resolver v1.4.4
	github.com/opentracing-contrib/go-gin v0.0.0-20201220185307-1dd2273433a4
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.8.1
	go.uber.org/zap v1.19.0
	google.golang.org/grpc v1.40.0
)
