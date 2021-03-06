/*
@Time : 2019/6/21 10:22
@Author : kenny zhu
@File : rcache
@Software: GoLand
@Others:
TODO: 增加ETCD失去链接或没有可用的servers情况下所有失活的list筛选处理...
*/
package rcache

import (
	"common/log/log"
	"common/monitor"
	"common/registry"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Cache is the registry cache interface
type Cache interface {
	// embed the registry interface
	registry.Registry
	// stop the cache watcher
	Stop()
}

type Option func(o *Options)

type cache struct {
	registry.Registry
	opts Options

	// registry cache
	sync.RWMutex
	cache   map[string][]*registry.Service
	ttls    map[string]time.Time
	watched map[string]bool

	exit chan bool
}

var (
	DefaultTTL = time.Minute
)

// 10 的 attempts 幂次方时间秒回落..
func backoff(attempts int) time.Duration {
	if attempts == 0 {
		return time.Duration(0)
	}
	return time.Duration(math.Pow(10, float64(attempts))) * time.Millisecond
}

// isValid checks if the service is valid
func (c *cache) isValid(services []*registry.Service, ttl time.Time) bool {
	// no services exist
	if len(services) == 0 {
		return false
	}

	// ttl is invalid
	if ttl.IsZero() {
		return false
	}

	// time since ttl is longer than timeout,
	if time.Since(ttl) > c.opts.TTL {
		return false
	}

	// ok
	return true
}

// exit cache
func (c *cache) quit() bool {
	select {
	case <-c.exit:
		return true
	default:
		return false
	}
}

// cp copies a service. Because we're caching handing back pointers would
// create a race condition, so we do this instead its fast enough
// attention: pointers.
func (c *cache) cp(current []*registry.Service) []*registry.Service {
	var services []*registry.Service

	for _, service := range current {
		// copy service
		s := new(registry.Service)
		*s = *service

		// copy nodes
		var nodes []*registry.Node
		for _, node := range service.Nodes {
			n := new(registry.Node)
			*n = *node
			nodes = append(nodes, n)
		}
		s.Nodes = nodes

		// copy endpoints
		var eps []*registry.Endpoint
		for _, ep := range service.Endpoints {
			e := new(registry.Endpoint)
			*e = *ep
			eps = append(eps, e)
		}
		s.Endpoints = eps

		// append service
		services = append(services, s)
	}

	return services
}

func (c *cache) del(service string) {
	delete(c.cache, service)
	delete(c.ttls, service)
}

func (c *cache) set(service string, services []*registry.Service) {
	c.cache[service] = services
	c.ttls[service] = time.Now().Add(c.opts.TTL)
}

func (c *cache) get(service string) ([]*registry.Service, error) {
	// read lock
	c.RLock()

	// check the cache first
	services := c.cache[service]
	// get cache ttl
	ttl := c.ttls[service]

	// got services && within ttl so return cache
	if c.isValid(services, ttl) {
		// make a copy
		cp := c.cp(services)
		// unlock the read
		c.RUnlock()
		// return servics
		return cp, nil
	}

	// get does the actual request for a service and cache it
	// if services over ttl..
	invalidServices := c.cp(services)
	get := func(service string) ([]*registry.Service, error) {
		// ask the registry
		services, err := c.Registry.GetService(service)
		if err != nil {
			// 防止ET-CD挂掉读取本地超时ttl...
			if len(invalidServices) > 0 {
				return invalidServices, nil
			}
			return nil, err
		}

		// cache results
		c.Lock()
		c.set(service, c.cp(services))
		c.Unlock()

		return services, nil
	}

	// watch service if not watched
	if _, ok := c.watched[service]; !ok {
		go c.run(service)
	}

	// unlock the read lock
	c.RUnlock()

	// get and return services
	return get(service)
}

// update by result
func (c *cache) update(res *registry.Result) {
	if res == nil || res.Service == nil {
		return
	}

	c.Lock()
	defer c.Unlock()

	services, ok := c.cache[res.Service.Name]
	if !ok {
		// we're not going to cache anything
		// unless there was already a lookup
		return
	}
	if len(res.Service.Nodes) == 0 {
		switch res.Action {
		case "delete":
			c.del(res.Service.Name)
		}
		return
	}

	var targetNode *registry.Node
	switch res.Action {
	case "create", "update":
		newService := true
		for _, service := range services {
			if service.Version != res.Service.Version {
				continue
			}
			newService = false
			for j, node := range service.Nodes {
				for _, newNode := range res.Service.Nodes {
					if node.Id != newNode.Id {
						continue
					}
					targetNode = newNode
					break
				}
				if targetNode != nil {
					service.Nodes[j] = targetNode
					targetNode = nil
				}
			}
		}
		if newService {
			c.set(res.Service.Name, append(services, res.Service))
		}
	case "delete":
		for _, service := range services {
			if service.Version != res.Service.Version {
				continue
			}
			for j, node := range service.Nodes {
				for _, newNode := range res.Service.Nodes {
					if node.Id != newNode.Id {
						continue
					}
					targetNode = newNode
					break
				}
				if targetNode != nil {
					service.Nodes[j] = targetNode
					service.Nodes[j].Metadata[monitor.ServiceStatus] = monitor.DeleteState
					targetNode = nil
				}
			}
		}
	}

	// existing services found
	/*var index int
	var service *registry.Service
	for i, s := range services {
		if s.Version == res.Service.Version {
			service = s
			index = i
		}
	}

	// log.Info("res.Action:" + res.Action)
	switch res.Action {
	case "create", "update":
		if service == nil {
			c.set(res.Service.Name, append(services, res.Service))
			return
		}

		// append old nodes to new service
		for _, cur := range service.Nodes {
			var seen bool
			for _, node := range res.Service.Nodes {
				if cur.Id == node.Id {
					seen = true
					break
				}
			}
			if !seen {
				res.Service.Nodes = append(res.Service.Nodes, cur)
			}
		}

		services[index] = res.Service
		c.set(res.Service.Name, services)
	case "delete":
		if service == nil {
			return
		}

		var nodes []*registry.Node

		// filter cur nodes to remove the dead one
		for _, cur := range service.Nodes {
			var seen bool
			for _, del := range res.Service.Nodes {
				if del.Id == cur.Id {
					seen = true
					break
				}
			}
			if seen {
				cur.Metadata[monitor.ServiceStatus] = monitor.DeleteState
			}
			nodes = append(nodes, cur)
		}

		// still got nodes, save and return
		if len(nodes) > 0 {
			service.Nodes = nodes
			services[index] = service
			c.set(service.Name, services)
			return
		}

		// zero nodes left
		// only have one thing to delete
		// nuke the thing
		if len(services) == 1 {
			c.del(service.Name)
			return
		}

		// still have more than 1 service
		// check the version and keep what we know
		var srvs []*registry.Service
		for _, s := range services {
			if s.Version != service.Version {
				srvs = append(srvs, s)
			}
		}

		// save
		c.set(service.Name, srvs)
	}*/
}

// run starts the cache watcher loop
// it creates a new watcher if there's a problem
func (c *cache) run(service string) {
	// set watcher
	c.Lock()
	c.watched[service] = true
	c.Unlock()

	// delete watcher on exit
	defer func() {
		c.Lock()
		delete(c.watched, service)
		c.Unlock()
	}()

	var a, b int

	for {
		// exit early if already dead
		if c.quit() {
			return
		}

		// jitter before starting
		j := rand.Int63n(100)
		time.Sleep(time.Duration(j) * time.Millisecond)

		// create new watcher
		w, err := c.Registry.Watch(
			registry.WatchService(service),
		)

		if err != nil {
			if c.quit() {
				return
			}

			d := backoff(a)

			if a > 3 {
				log.Error("r-cache: ", err, " backing off ", d)
				a = 0
			}

			time.Sleep(d)
			a++

			continue
		}

		// reset a
		a = 0

		// watch for events
		if err := c.watch(w); err != nil {
			if c.quit() {
				return
			}

			d := backoff(b)

			if b > 3 {
				log.Error("r-cache: ", err, " backing off ", d)
				b = 0
			}

			time.Sleep(d)
			b++

			continue
		}

		// reset b
		b = 0
	}
}

// watch loops the next event and calls update
// it returns if there's an error
func (c *cache) watch(w registry.Watcher) error {
	defer w.Stop()

	// manage this loop
	go func() {
		// wait for exit
		<-c.exit
		w.Stop()
	}()

	for {
		// block until result return.
		res, err := w.Next()
		if err != nil {
			return err
		}
		c.update(res)
	}
}

func (c *cache) GetService(service string) ([]*registry.Service, error) {
	// get the service
	services, err := c.get(service)
	if err != nil {
		return nil, err
	}

	// if there's nothing return err
	if len(services) == 0 {
		return nil, registry.ErrNotFound
	}

	// return services
	return services, nil
}

// deregister if error happens
func (c *cache) Deregister(s *registry.Service) error {
	err := c.Registry.Deregister(s)
	if err != nil {
		return err
	}

	// todo : delete service node, wait for next watch or just update here?

	return nil
}

func (c *cache) Stop() {
	select {
	case <-c.exit:
		return
	default:
		close(c.exit)
	}
}

func (c *cache) String() string {
	return "rcache"
}

// New returns a new cache
func New(r registry.Registry, opts ...Option) Cache {
	rand.Seed(time.Now().UnixNano())
	options := Options{
		TTL: DefaultTTL,
	}

	for _, o := range opts {
		o(&options)
	}

	return &cache{
		Registry: r,
		opts:     options,
		watched:  make(map[string]bool),
		cache:    make(map[string][]*registry.Service),
		ttls:     make(map[string]time.Time),
		exit:     make(chan bool),
	}
}
