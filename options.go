package downloadmanager

import context "context"

type Options struct {
	ctx         context.Context
	checkMd5Sum bool
	md5Sum      string
	cache       bool
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

func CheckMD5Sum(b bool) Option {
	return func(o *Options) {
		o.checkMd5Sum = b
	}
}

func MD5Sum(s string) Option {
	return func(o *Options) {
		if s == "" {
			return
		}
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
