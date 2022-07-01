package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/wire"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/spf13/viper"
)

var (
	httpPool           *http.Client
	HttpClientProvider = wire.NewSet(NewClientOptions, NewClient)
)

const (
	// DefaultTimeout
	DefaultConnectTimeout = time.Second * 10
	DefaultReadTimeout    = time.Second * 10
	DefaultWriteTimeout   = time.Second * 10
	// DefaultRetryTimes 如果请求失败，最多重试3次
	DefaultRetryTimes = 3
	// DefaultRetryDelay 在重试前，延迟等待100毫秒
	DefaultRetryDelay = time.Millisecond * 100
)

// ClientOptions
type ClientOptions struct {
	connectTimeout time.Duration
	readTimeout    time.Duration
	writeTimeout   time.Duration
	retryTimes     int
	consulOptions  *consulApi.Config
}

// NewClientOptions
func NewClientOptions(v *viper.Viper) (*ClientOptions, error) {
	var (
		err error
		o   = new(ClientOptions)
	)
	if err = v.UnmarshalKey("http.client", o); err != nil {
		return nil, err
	}
	return o, nil
}

func httpClient(o *ClientOptions) (*http.Client, error) {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   (time.Duration(o.connectTimeout)) * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConnsPerHost:   100, //默认是2
		ResponseHeaderTimeout: time.Duration(o.readTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(o.writeTimeout) * time.Second,
	}
	if o.consulOptions != nil {
		return consulApi.NewHttpClient(transport, o.consulOptions.TLSConfig)
	}

	to := o.connectTimeout + o.readTimeout + o.writeTimeout
	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(to) * time.Second, //默认是0，无超时
	}, nil
}

// ClientOptional
type ClientOptional func(o *ClientOptions)

// WithConnectTimeout
func WithConnectTimeout(t time.Duration) ClientOptional {
	return func(opt *ClientOptions) {
		opt.connectTimeout = t
	}
}

// WithReadTimeout
func WithReadTimeout(t time.Duration) ClientOptional {
	return func(opt *ClientOptions) {
		opt.readTimeout = t
	}
}

// WithWriteTimeout
func WithWriteTimeout(t time.Duration) ClientOptional {
	return func(opt *ClientOptions) {
		opt.writeTimeout = t
	}
}

// WithRetryTimes 设置失败重试
func WithRetryTimes(retryTimes int) ClientOptional {
	return func(opt *ClientOptions) {
		opt.retryTimes = retryTimes
	}
}

// WithConsulConfig
func WithConsulConfig(consul *consulApi.Config) ClientOptional {
	return func(opt *ClientOptions) {
		opt.consulOptions = consul
	}
}

// Client
type Client struct {
	client  *http.Client
	options *ClientOptions
	tracer  opentracing.Tracer
}

// NewClient
func NewClient(o *ClientOptions, tracer opentracing.Tracer) (cli *Client, err error) {
	if httpPool == nil {
		httpPool, err = httpClient(o)
		if err != nil {
			return
		}
	}
	cli = &Client{
		client:  httpPool,
		options: o,
		tracer:  tracer,
	}
	return
}

// Get
func (c *Client) Get(ctx context.Context, uri string,
	form url.Values, header map[string]string, options ...ClientOptional) (body []byte, err error) {
	o := parseOptions(options...)
	return c.withoutBody(ctx, http.MethodGet, uri, form, header, o)
}

// Delete
func (c *Client) Delete(ctx context.Context, uri string,
	form url.Values, header map[string]string, options ...ClientOptional) (body []byte, err error) {
	o := parseOptions(options...)
	return c.withoutBody(ctx, http.MethodDelete, uri, form, header, o)
}

func (c *Client) Post(ctx context.Context, uri string,
	bodyData interface{}, header map[string]string, options ...ClientOptional) (body []byte, err error) {
	o := parseOptions(options...)
	return c.withBody(ctx, http.MethodPost, uri, bodyData, header, o)
}
func (c *Client) Put(ctx context.Context, uri string,
	bodyData interface{}, header map[string]string, options ...ClientOptional) (body []byte, err error) {
	o := parseOptions(options...)
	return c.withBody(ctx, http.MethodPut, uri, bodyData, header, o)
}
func (c *Client) Patch(ctx context.Context, uri string,
	bodyData interface{}, header map[string]string, options ...ClientOptional) (body []byte, err error) {
	o := parseOptions(options...)
	return c.withBody(ctx, http.MethodPatch, uri, bodyData, header, o)
}

func (c *Client) withoutBody(ctx context.Context, method, uri string,
	form url.Values, header map[string]string, options *ClientOptions) (body []byte, err error) {
	if uri == "" {
		return nil, errors.New("uri required")
	}

	if len(form) > 0 {
		if uri, err = buildQuery(uri, form); err != nil {
			return nil, err
		}
	}

	clientSpan := parseTrace(ctx, method, "httpClient-"+method, c.tracer)
	defer clientSpan.Finish()

	if header["Content-Type"] == "" {
		header["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	}

	return c.request(ctx, method, uri, nil, header)
}

func (c *Client) withBody(ctx context.Context, method, uri string, bodyData interface{}, header map[string]string, options *ClientOptions) (body []byte, err error) {
	if uri == "" {
		return nil, errors.New("uri required")
	}

	clientSpan := parseTrace(ctx, method, "httpClient-"+method, c.tracer)
	defer clientSpan.Finish()

	if header["Content-Type"] == "" {
		header["Content-Type"] = "application/json; charset=utf-8"
	}

	return c.request(ctx, method, uri, bodyData, header)
}

func (c *Client) request(ctx context.Context, method, uri string, bodyData interface{},
	header map[string]string) ([]byte, error) {
	var body io.Reader
	if bodyData != nil {
		bodyRaw, err := json.Marshal(bodyData)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bodyRaw)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// 重试body定义
	var (
		retryBody io.ReadCloser
		retryFlag = false
		retryErr  error
		resp      *http.Response
	)

	for i := 0; i < c.options.retryTimes; i++ {
		// 赋值重试body 执行请求会读取buffer 导致body为空
		if req.Method == http.MethodPost {
			retryBody, _ = req.GetBody()
			if retryFlag {
				req.Body = retryBody
			}
		}
		resp, retryErr = c.client.Do(req)
		if retryErr != nil {
			retryFlag = true
		} else if resp.StatusCode == http.StatusOK {
			// retryFlag = false
			break
		} else {
			// 如果状态码不为200，报错并继续重试
			retryFlag = true
		}
	}

	if retryErr != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	bin, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bin, nil
}

func parseTrace(ctx context.Context, method, tag string, tracer opentracing.Tracer) opentracing.Span {
	var parentCtx opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentCtx = parent.Context()
	}
	clientSpan := tracer.StartSpan(
		method,
		opentracing.ChildOf(parentCtx),
		ext.SpanKindRPCClient,
		opentracing.Tag{Key: string(ext.Component), Value: tag},
	)
	return clientSpan
}

func parseOptions(options ...ClientOptional) *ClientOptions {
	o := &ClientOptions{
		connectTimeout: DefaultConnectTimeout,
		readTimeout:    DefaultReadTimeout,
		writeTimeout:   DefaultWriteTimeout,
		retryTimes:     DefaultRetryTimes,
	}
	for _, option := range options {
		option(o)
	}
	return o
}

func buildQuery(uri string, form url.Values) (string, error) {
	if len(form) == 0 {
		return "", errors.New("form required")
	}
	target, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	urlValues := target.Query()
	for k, v := range form {
		for _, i := range v {
			urlValues.Add(k, i)
		}
	}

	target.RawQuery = urlValues.Encode()
	return target.String(), nil
}
