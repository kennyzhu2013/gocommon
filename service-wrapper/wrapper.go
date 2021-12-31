/*
@Time : 2019/5/13 10:36
@Author : kenny zhu
@File : wrapper.go
@Software: GoLand
@Others:
*/
package service_wrapper

import (
	"common/web"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Service is a web service with service discovery built in
// TODO: use client.Client
type Service interface {
	Client(opts ...web.Option) *http.Client
	Options() Options
	Run() error
	Shutdown() error
}

type Option func(o *Options)
type Runner func(addr ...string) (err error)

var (
	// For serving
	_defaultName    = "gin-web"
	_defaultVersion = "latest"
	_defaultId      = uuid.New().String()
	_defaultAddress = ":0"

	// for registration
	_defaultRegisterTTL      = time.Minute
	_defaultRegisterInterval = time.Second * 15

	// http options.
	_defaultReadTimeout  = 5 * time.Second
	_defaultWriteTimeout = 5 * time.Second

	// shut time.
	_defaultShutdownTimeout = 3 * time.Second
)

// NewService returns a new web.ServiceWrapper, for future use service for extend.
func NewService(opts ...Option) Service {
	return newService(opts...)
}
