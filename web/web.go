/*
@Time : 2019/4/19 16:23
@Author : kenny zhu
@File : web.go
@Software: GoLand
@Others:
*/
package web

import (
	"common/log/log"
	"common/registry"
	"common/selector"
	"net"
	"net/http"
	"time"
)

// https://stackoverflow.com/questions/26707941/go-roundtripper-and-transport.
func NewRoundShardTripper(opts ...Option) http.RoundTripper {
	options := Options{
		Registry:    registry.DefaultRegistry,
		Selector:    selector.Random,
		Destination: "X-Media-Server",
		RouteTag:    "serverId",
	}
	for _, o := range opts {
		o(&options)
	}

	var newRt http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   40,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	r := &roundShardTripper{
		rt:   newRt, // http.DefaultTransport,
		opts: options,
	}
	r.rc = r.newRCache()

	return r
}

// get the destination topics
func GetServiceNodeByIp(srvName string, ip string) (recordTopics string, audioTopics string) {
	s, err := registry.GetService(srvName)
	if err != nil {
		log.Error("GetServiceNodeByIp failed, no service named:", srvName)
		return "", ""
	}

	// var nodeResult *registry.Node
	for _, service := range s {
		for _, node := range service.Nodes {
			// if node.Address == proxyIp && node.Port == proxyPort {
			if node.Address == ip {
				return node.Metadata["storeTopic"], node.Metadata["audioTopic"]
			}
		}
	}

	log.Error("GetServiceNodeByIp failed, no node with ip:", ip)
	return "", ""
}
