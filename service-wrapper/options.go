/*
@Time : 2019/5/13 10:37
@Author : kenny zhu
@File : options.go
@Software: GoLand
@Others:
*/
package service_wrapper

import (
	"common/registry"
	"github.com/gin-gonic/gin"
	"time"

	"context"
)

type Options struct {
	Address   string
	Advertise string

	// service
	Name        string
	Version     string
	Id          string
	Metadata    MetaData
	Description string

	// or service struct directly..
	ServiceInfo *registry.Service

	Registry         registry.Registry
	RegisterTTL      time.Duration
	RegisterInterval time.Duration

	// define gin.Engine
	Engine *gin.Engine

	// https config
	Secure    bool
	TLSConfig TLSFile

	// Alternative Options
	Context context.Context

	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

type TLSFile struct {
	CertFile string
	KeyFile  string
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Name:             _defaultName,
		Version:          _defaultVersion,
		Id:               _defaultId,
		Address:          _defaultAddress,
		RegisterTTL:      _defaultRegisterTTL,
		RegisterInterval: _defaultRegisterInterval,
		Context:          context.TODO(),
		ShutdownTimeout:  _defaultShutdownTimeout,
		ReadTimeout:      _defaultReadTimeout,
		WriteTimeout:     _defaultWriteTimeout,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Server name
func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// Unique server id
func Id(id string) Option {
	return func(o *Options) {
		o.Id = id
	}
}

// Version of the service
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// MetaData associated with the service
func Metadata(md MetaData) Option {
	return func(o *Options) {
		o.Metadata = md
	}
}

// Address to bind to - host:port
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// The address to advertise for discovery - host:port
func Advertise(a string) Option {
	return func(o *Options) {
		o.Advertise = a
	}
}

func Description(a string) Option {
	return func(o *Options) {
		o.Description = a
	}
}

// Context specifies a context for the service.
// Can be used to signal shutdown of the service.
// Can be used for extra option values.
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

func Engine(g *gin.Engine) Option {
	return func(o *Options) {
		o.Engine = g
	}
}

func ServiceInfo(s *registry.Service) Option {
	return func(o *Options) {
		o.ServiceInfo = s
	}
}

// Secure Use secure communication. If TLSConfig is not specified we use InsecureSkipVerify and generate a self signed cert
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// TLSConfig to be used for the transport.
func TLSConfig(t TLSFile) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

// ReadTimeout -.
func ReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WriteTimeout -.
func WriteTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = timeout
	}
}

// ShutdownTimeout -.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ShutdownTimeout = timeout
	}
}
