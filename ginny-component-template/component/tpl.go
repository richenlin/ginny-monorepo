package component

const Tpl = `package {{ .Name }}

import (
	"github.com/goriller/ginny/v2"
)

// Option for {{ .Name }} component.
type Option func(*options)

type options struct{}

// New creates a new {{ .Name }} component.
func New(opts ...Option) ginny.LifecycleHook {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return &component{opts: o}
}

type component struct {
	opts *options
}

func (c *component) Name() string                    { return "{{ .Name }}" }
func (c *component) OnStart(ctx context.Context) error { return nil }
func (c *component) OnStop(ctx context.Context) error  { return nil }
func (c *component) Priority() int                    { return 50 }
`
