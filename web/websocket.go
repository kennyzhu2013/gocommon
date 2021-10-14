package web

import (
	"common/log/log"
	"common/registry"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	//"rcache"
	"common/selector"
	"strings"
	//"time"

	"sync"
)

//封装websocket并发读写操作.
//server refer:https://github.com/owenliang/go-websocket/blob/master/server.go.
// client refer: https://medium.com/@sachinshinde7676/getting-started-with-websocket-client-in-go-73baaf8b5caf.

type WSocket struct {
	Conn              *websocket.Conn
	WebsocketDialer   *websocket.Dialer
	Url               string
	ConnectionOptions ConnectionOptions
	opts              Options
	// rc 				  rcache.Cache
	RequestHeader   http.Header
	OnConnected     func(socket WSocket)
	OnTextMessage   func(message string, socket WSocket)
	OnBinaryMessage func(data []byte, socket WSocket)
	OnConnectError  func(err error, socket WSocket)
	OnDisconnected  func(err error, socket WSocket)
	OnPingReceived  func(data string, socket WSocket)
	OnPongReceived  func(data string, socket WSocket)
	IsConnected     bool
	sendMu          *sync.Mutex // Prevent "concurrent write to websocket connection"
	receiveMu       *sync.Mutex
}

type ConnectionOptions struct {
	UseCompression bool
	UseSSL         bool
	Proxy          func(*http.Request) (*url.URL, error)
	Subprotocols   []string
}

// TODO: Yet to be done, not support reconnect now!.
type ReconnectionOptions struct {
}

// url is like: "ws://echo.websocket.org:8080/get".
// set proxy with et-cd registry
func New(url string) *WSocket {
	return &WSocket{
		Url:           url,
		RequestHeader: http.Header{},
		ConnectionOptions: ConnectionOptions{
			UseCompression: false,
			UseSSL:         true,
		},
		WebsocketDialer: &websocket.Dialer{Proxy: websocket.DefaultDialer.Proxy, HandshakeTimeout: websocket.DefaultDialer.HandshakeTimeout},
		sendMu:          &sync.Mutex{},
		receiveMu:       &sync.Mutex{},
	}
}

func NewWithRegistry(url string, opts ...Option) *WSocket {
	// for present.
	options := Options{
		Registry:    registry.DefaultRegistry,
		Selector:    selector.RoundBinSelect,
		Destination: "X-Media-Server",
		RouteTag:    "serverId",
	}
	for _, o := range opts {
		o(&options)
	}

	result := &WSocket{
		Url:           url,
		RequestHeader: http.Header{},
		ConnectionOptions: ConnectionOptions{
			UseCompression: false,
			UseSSL:         true,
		},
		WebsocketDialer: &websocket.Dialer{Proxy: websocket.DefaultDialer.Proxy, HandshakeTimeout: websocket.DefaultDialer.HandshakeTimeout},
		opts:            options,
		sendMu:          &sync.Mutex{},
		receiveMu:       &sync.Mutex{},
	}

	//result.WebsocketDialer = &websocket.Dialer{Proxy: result.proxy, HandshakeTimeout: websocket.DefaultDialer.HandshakeTimeout}
	//result.ConnectionOptions.Proxy = result.proxy
	// result.rc = result.newRCache()
	return result
}

//func (socket *WSocket) newRCache() rcache.Cache {
//	ropts := []rcache.Option{}
//	if socket.opts.Context != nil {
//		if t, ok := socket.opts.Context.Value("selector_ttl").(time.Duration); ok {
//			ropts = append(ropts, rcache.WithTTL(t))
//		}
//	}
//	return rcache.New(socket.opts.Registry, ropts...)
//}

// Url = ws://go.micro.xaudiobusiness/audiobusiness
// TODO: some problems here.
func (socket *WSocket) proxy(r *http.Request) (*url.URL, error) {
	destUrl, err := socket.getUrlFromRegistry()
	if destUrl == "" || err != nil {
		return nil, nil
	}

	u, error := url.Parse(destUrl)
	switch u.Scheme {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	default:
		return nil, nil
	}

	if u.User != nil {
		// User name and password are not allowed in websocket URIs.
		return nil, nil
	}
	r.Host = u.Host
	return u, error
}

func (socket *WSocket) getUrlFromRegistry() (string, error) {
	// check valid.
	if len(socket.Url) < 7 {
		log.Error("socket.Url is too short!")
		return "", nil
	}

	subUrl := socket.Url[5:]
	firtPos := strings.Index(subUrl, "/")
	if firtPos < 1 {
		log.Error("socket.Url is invalid!")
		return "", nil
	}

	// seps := strings.Split(subUrl, "/")
	serviceName := subUrl[:firtPos]
	suffix := subUrl[firtPos:]

	services, err := registry.GetService(serviceName)
	if err != nil || len(services) == 0 {
		log.Error("registry.GetService failed, serviceName:%v, err:%v!", serviceName, err)
		return "", nil
	}

	next := socket.opts.Selector(services)
	// rudimentary retry 3 times , may be the same one.
	for i := 0; i < 3; i++ {
		n, err := next()
		if err != nil {
			continue
		}
		if nil == n {
			log.Error("proxy failed not found any normal node")
			return "", errors.New("proxy failed not found any normal node")
		}
		destAddr := fmt.Sprintf("%s:%d", n.Address, n.Port)
		destUrl := "ws://" + destAddr + suffix
		log.Infof("proxy found node with ip:%v, port:%v, destUrl:%v", n.Address, n.Port, destUrl)
		return destUrl, nil
	}

	return "", nil
}

func (socket *WSocket) setConnectionOptions() {
	socket.WebsocketDialer.EnableCompression = socket.ConnectionOptions.UseCompression
	socket.WebsocketDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: socket.ConnectionOptions.UseSSL}
	socket.WebsocketDialer.Proxy = socket.ConnectionOptions.Proxy
	socket.WebsocketDialer.Subprotocols = socket.ConnectionOptions.Subprotocols
}

func (socket *WSocket) Connect() {
	var err error
	socket.setConnectionOptions()

	// write directly here.
	destUrl, err := socket.getUrlFromRegistry()
	if destUrl == "" || err != nil {
		destUrl = socket.Url
	}
	socket.Conn, _, err = socket.WebsocketDialer.Dial(destUrl, socket.RequestHeader)

	if err != nil {
		log.Errorf("Error while connecting to server ", err)
		socket.IsConnected = false
		if socket.OnConnectError != nil {
			socket.OnConnectError(err, *socket)
		}
		return
	}
	log.Info("Connected to server")

	if socket.OnConnected != nil {
		socket.IsConnected = true
		socket.OnConnected(*socket)
	}

	// for heart beat begin..........
	defaultPingHandler := socket.Conn.PingHandler()
	socket.Conn.SetPingHandler(func(appData string) error {
		log.Infof("Received PING from server:%v", appData)
		if socket.OnPingReceived != nil {
			socket.OnPingReceived(appData, *socket)
		}
		if defaultPingHandler != nil {
			return defaultPingHandler(appData)
		}
		return nil
	})

	defaultPongHandler := socket.Conn.PongHandler()
	socket.Conn.SetPongHandler(func(appData string) error {
		log.Infof("Received PONG from server:%v", appData)
		if socket.OnPongReceived != nil {
			socket.OnPongReceived(appData, *socket)
		}

		if defaultPongHandler != nil {
			return defaultPongHandler(appData)
		}
		return nil
	})

	defaultCloseHandler := socket.Conn.CloseHandler()
	socket.Conn.SetCloseHandler(func(code int, text string) error {
		result := defaultCloseHandler(code, text)
		log.Infof("Disconnected from server %v", result)
		if socket.OnDisconnected != nil {
			socket.IsConnected = false
			socket.OnDisconnected(errors.New(text), *socket)
		}
		return result
	})
	// for heart beat end..........

	// receive call back.
	go func() {
		for {
			socket.receiveMu.Lock()
			messageType, message, err := socket.Conn.ReadMessage()
			socket.receiveMu.Unlock()
			if err != nil {
				log.Infof("read:%v", err)
				return
			}
			log.Infof("recv: %v", string(message))

			switch messageType {
			case websocket.TextMessage:
				if socket.OnTextMessage != nil {
					socket.OnTextMessage(string(message), *socket)
				}
			case websocket.BinaryMessage:
				if socket.OnBinaryMessage != nil {
					socket.OnBinaryMessage(message, *socket)
				}
			}
		}
	}()
}

func NewWSocket(conn *websocket.Conn) *WSocket {
	socket := New("")
	socket.Conn = conn
	log.Infof("Listen to client:%v", conn.RemoteAddr())
	if socket.OnConnected != nil {
		socket.IsConnected = true
		socket.OnConnected(*socket)
	}

	// for heart beat begin..........
	defaultPingHandler := socket.Conn.PingHandler()
	socket.Conn.SetPingHandler(func(appData string) error {
		log.Infof("Received PING from client:%v", appData)
		if socket.OnPingReceived != nil {
			socket.OnPingReceived(appData, *socket)
		}
		if defaultPingHandler != nil {
			return defaultPingHandler(appData)
		}
		return nil
	})

	defaultPongHandler := socket.Conn.PongHandler()
	socket.Conn.SetPongHandler(func(appData string) error {
		log.Infof("Received PONG from client:%v", appData)
		if socket.OnPongReceived != nil {
			socket.OnPongReceived(appData, *socket)
		}

		if defaultPongHandler != nil {
			return defaultPongHandler(appData)
		}
		return nil
	})

	defaultCloseHandler := socket.Conn.CloseHandler()
	socket.Conn.SetCloseHandler(func(code int, text string) error {
		result := defaultCloseHandler(code, text)
		log.Infof("Disconnected from client %v", result)
		if socket.OnDisconnected != nil {
			socket.IsConnected = false
			socket.OnDisconnected(errors.New(text), *socket)
		}
		return result
	})
	// for heart beat end..........
	return socket
}

func (socket *WSocket) Start() {
	// receive call back.
	go func() {
		for {
			socket.receiveMu.Lock()
			messageType, message, err := socket.Conn.ReadMessage()
			socket.receiveMu.Unlock()
			if err != nil {
				log.Infof("read:%v", err)
				return
			}
			//log.Infof("recv: %v", message)

			switch messageType {
			case websocket.TextMessage:
				//log.Infof("receive text = %v",string(message))
				if socket.OnTextMessage != nil {
					//log.Infof("if socket.OnTextMessage != nil ,callback ")
					socket.OnTextMessage(string(message), *socket)
				}
			case websocket.BinaryMessage:
				if socket.OnBinaryMessage != nil {
					socket.OnBinaryMessage(message, *socket)
				}
			}
		}
	}()
}

func (socket *WSocket) SendText(message string) {
	err := socket.send(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Infof("write:%v", err)
		return
	}
}

func (socket *WSocket) SendBinary(data []byte) {
	err := socket.send(websocket.BinaryMessage, data)
	if err != nil {
		log.Infof("write:%v", err)
		return
	}
}

func (socket *WSocket) send(messageType int, data []byte) error {
	socket.sendMu.Lock()
	err := socket.Conn.WriteMessage(messageType, data)
	socket.sendMu.Unlock()
	return err
}

func (socket *WSocket) Close() {
	err := socket.send(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Infof("write close:%v", err)
	}
	socket.Conn.Close()
	if socket.OnDisconnected != nil {
		socket.IsConnected = false
		socket.OnDisconnected(err, *socket)
	}
}
