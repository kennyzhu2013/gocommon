/*
@Time : 2019/5/10 10:16
@Author : kenny zhu
@File : options.go
@Software: GoLand
@Others:
*/
package metrics

import "github.com/prometheus/client_golang/prometheus"

type Options struct {
	Namespace   	string
	Id          	string
	Version     	string
	BHostMethod 	bool
	UrlField    	string // http filed to get method label = "zy-url"
	MetaData    	map[string]string
	Merge			[]string
	Function    	map[string]func() float64

	// self define collector, integrate with gin framework.
	Others []prometheus.Collector
}

func Id(n string) Option {
	return func(o *Options) {
		o.Id = n
	}
}

func Namespace(n string) Option {
	return func(o *Options) {
		o.Namespace = n
	}
}

func Version(n string) Option {
	return func(o *Options) {
		o.Version = n
	}
}

func HostMethod(n bool) Option {
	return func(o *Options) {
		o.BHostMethod = n
	}
}

func MetaData(n map[string]string) Option {
	return func(o *Options) {
		o.MetaData = n
	}
}

func Merge(n []string) Option {
	return func(o *Options) {
		o.Merge = n
	}
}

func Function(n map[string]func() float64) Option {
	return func(o *Options) {
		o.Function = n
	}
}

func UrlField(n string) Option {
	return func(o *Options) {
		o.UrlField = n
	}
}

func AddCollector(c prometheus.Collector) Option {
	return func(o *Options) {
		o.Others = append(o.Others, c)
	}
}
