package downloadmanager

import context "golang.org/x/net/context"

type Options struct {
	ctx    context.Context
	md5Sum string
	cache  bool
}

type Option func(*Options)

var (
	DefaultCachePolicy = true
)

func Context(c context.Context) Option {
	return func(o *Options) {
		o.ctx = c
	}
}

func MD5Sum(s string) Option {
	return func(o *Options) {
		o.md5Sum = s
	}
}

func Cache(b bool) Option {
	return func(o *Options) {
		o.cache = b
	}
}

func WithOptions(e *Options) Option {
	return func(o *Options) {
		*o = *e
	}
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		ctx:    context.Background(),
		md5Sum: "",
		cache:  DefaultCachePolicy,
	}

	for _, o := range opts {
		o(options)
	}

	return options
}
