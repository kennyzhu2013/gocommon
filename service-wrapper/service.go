/*
@Time : 2019/5/13 10:46
@Author : kenny zhu
@File : service.go
@Software: GoLand
@Others:
*/
package service_wrapper

import (
	"common/log/newlog"
	"common/registry"
	"common/web"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"net/http"
	"os"
	"sync"
	"time"

	maddr "common/util/addr"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// gin based..
type service struct {
	opts Options

	// mux *http.ServeMux
	server *http.Server
	notify chan error

	srv *registry.Service

	sync.Mutex
	running bool
}

func newService(opts ...Option) Service {
	options := newOptions(opts...)
	s := &service{
		opts: options,
		// mux:  http.NewServeMux(),
	}

	if s.opts.ServiceInfo != nil {
		s.srv = s.opts.ServiceInfo
	} else {
		s.srv = s.genSrv()
	}

	s.notify = make(chan error, 1)
	s.server = &http.Server{
		Handler:      s.opts.Engine,
		ReadTimeout:  s.opts.ReadTimeout,
		WriteTimeout: s.opts.WriteTimeout,
		Addr:         s.opts.Address,
	}
	return s
}

// gin service init..
func (s *service) genSrv() *registry.Service {
	// default host:port
	parts := strings.Split(s.opts.Address, ":")

	// support many ip-s
	host := strings.Join(parts[:len(parts)-1], ":")
	port, _ := strconv.Atoi(parts[len(parts)-1])

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(s.opts.Advertise) > 0 {
		parts = strings.Split(s.opts.Advertise, ":")

		// we have host:port
		if len(parts) > 1 {
			// set the host
			host = strings.Join(parts[:len(parts)-1], ":")

			// get the port
			if aport, _ := strconv.Atoi(parts[len(parts)-1]); aport > 0 {
				port = aport
			}
		} else {
			host = parts[0]
		}
	}

	addr, err := maddr.Extract(host)
	if err != nil {
		// best effort localhost, default
		addr = "127.0.0.1"
	}

	// format last address
	s.opts.Address = addr + ":" + strconv.Itoa(port)

	// create server.
	if s.opts.Engine == nil {
		gin.SetMode(gin.ReleaseMode)
		s.opts.Engine = gin.Default()
	}

	return &registry.Service{
		Name: s.opts.Name,
		Metadata: MetaData{
			"serverDescription": s.opts.Description, // server desc.
		},
		Nodes: []*registry.Node{
			{
				Id:       "go.micro.api.media-proxy-",
				Address:  addr,
				Port:     port,
				Metadata: s.opts.Metadata,
			},
		},
		Version: s.opts.Version,
	}
}

// heart beat run here, use monitor to update self tags.
func (s *service) run(exit chan bool) {
	if s.opts.RegisterInterval <= time.Duration(0) {
		// use HeartBeatCheck if no
		return
	}

	t := time.NewTicker(s.opts.RegisterInterval)
	for {
		select {
		case <-t.C:
			// just retry circle if failed
			if err := s.register(); err != nil {
				log.Errorf("s.register failed:%v", err)
			}
		case <-exit:
			t.Stop()
			return
		}
	}
}

func (s *service) register() error {
	if s.srv == nil {
		return nil
	}
	// default to service registry
	r := s.opts.Registry
	if s.opts.Registry == nil {
		return errors.New("Registry is empty. ")
	}
	return r.Register(s.srv, registry.RegisterTTL(s.opts.RegisterTTL))
}

func (s *service) deregister() error {
	if s.srv == nil {
		return nil
	}
	// default to service registry
	r := s.opts.Registry
	if s.opts.Registry == nil {
		return errors.New("Registry is empty. ")
	}
	return r.Deregister(s.srv)
}

// start server..
func (s *service) start() error {
	s.Lock()
	defer s.Unlock()

	if s.running {
		return nil
	}

	if s.opts.Secure {
		s.opts.Engine.Use(s.LoadTLS())
		go func() {
			s.notify <- s.server.ListenAndServeTLS(s.opts.TLSConfig.CertFile, s.opts.TLSConfig.KeyFile)
			close(s.notify)
		}()
	} else {
		go func() {
			s.notify <- s.server.ListenAndServe()
			close(s.notify)
		}()
	}

	// s.exit = make(chan chan error, 1)
	s.running = true

	log.Fatal("Service StartSuccess ! Listening on ", s.opts.Address)
	fmt.Println("Service StartSuccess ! Listening on", s.opts.Address)
	return nil
}

func (s *service) stop() {
	s.Lock()
	defer s.Unlock()
	if !s.running {
		return
	}

	s.running = false

	log.Info("Stopping")
	return
}

// load tls
func (s *service) LoadTLS() gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     s.opts.Address,
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}

// for http-client, must add own rt.
// TODO: use client.Client
func (s *service) Client(opts ...web.Option) *http.Client {
	// use random selector and  replace http.NewRoundTripper with web tripper
	rt := web.NewRoundShardTripper(opts...)

	return &http.Client{
		Transport: rt,
	}
}

// main run
func (s *service) Run() error {
	if err := s.start(); err != nil {
		return err
	}

	// start reg monitor loop
	ex := make(chan bool)
	if s.opts.Registry != nil {
		if err := s.register(); err != nil {
			return err
		}

		go s.run(ex)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	select {
	// wait on kill signal
	case sig := <-ch:
		log.Infof("Received signal %v\n", sig)
	// wait on context cancel
	case <-s.opts.Context.Done():
		s.Shutdown()
		log.Info("Received context shutdown")
	case err := <-s.notify:
		fmt.Printf("service StartFail Notify, listen address:%s, err:%s, quit now!", s.opts.Address, err.Error())
		log.Errorf("service StartFail Notify, listen address:%v,  err:%v, quit now!", s.opts.Address, err)
		// os.Exit(-1)
	}

	// exit reg loop
	if s.opts.Registry != nil {
		close(ex)
		if err := s.deregister(); err != nil {
			return err
		}
	}

	s.stop()
	return nil
}

// Safe: Shutdown with timeout-.
func (s *service) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.opts.ShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// Options returns the options for the given service
func (s *service) Options() Options {
	return s.opts
}
