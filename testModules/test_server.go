package main

import (
	"common/monitor"
	"common/registry"
	serviceWrapper "common/service-wrapper"
	"github.com/gin-gonic/gin"
	"github.com/pterm/pterm"
	"net/http"
	"strconv"
)

var (
	service = &registry.Service{
		Name: "go.micro.http-proxy",
		Metadata: map[string]string{
			"serverDescription": "http proxy for internet access",
		},
		Nodes: []*registry.Node{
			{
				Id:      "go.micro.http-proxy-",
				Address: "localhost",
				Port:    1210,
				Metadata: map[string]string{
					"serverTag":           "http-proxy",
					monitor.ServiceStatus: monitor.NormalState,
				},
			},
		},
		Version: "v1",
	}
)

var Handlers struct {
	Router *gin.Engine
}

func init() {
	gin.SetMode(gin.DebugMode)
	Handlers.Router = gin.New()
	Handlers.Router.Use(gin.Recovery())
}

type ExampleJson struct {
	Version string `json:"version"`
	Session int    `json:"session"`
	State   int    `json:"state"`
}

func ExampleHandler(ctx *gin.Context) {
	res := &ExampleJson{"1.0", 100, 1}
	ctx.JSON(http.StatusOK, res)
	pterm.FgWhite.Println("ExampleHandler returned.")
}

func main() {
	prefixName := "/micro-registry/xtest"
	_ = registry.DefaultRegistry.Init(registry.Addrs("10.153.90.4:2379"), registry.Prefix(prefixName))

	// service node init.
	service.Name = "com.cmic.test"
	// get internal ip
	service.Nodes[0].Address = "10.153.90.4"
	service.Nodes[0].Port = 9080
	service.Nodes[0].Id = service.Nodes[0].Address + ":" + strconv.Itoa(service.Nodes[0].Port)

	// for prometheus metrics..
	Handlers.Router.GET("/Example", ExampleHandler)
	Handlers.Router.POST("/Example", ExampleHandler)

	// init server
	// user service wrapper
	server := serviceWrapper.NewService(serviceWrapper.Address("10.153.90.4:9080"),
		serviceWrapper.Engine(Handlers.Router),
		serviceWrapper.ServiceInfo(service),
		serviceWrapper.RegisterInterval(monitor.HeartBeatCheck), serviceWrapper.Registry(registry.DefaultRegistry))

	if err := server.Run(); err != nil {
		pterm.FgRed.Printfln("Proxy server started failed, %v", err)
	}

	pterm.FgWhite.Println("server exist.")
}
