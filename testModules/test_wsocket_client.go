package main

import (
	"common/web"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"time"
)

func main() {
	wsClient := web.New("ws://10.153.90.4:8080/echo")
	done := make(chan struct{})
	wsClient.OnConnected = func(socket web.WSocket) {
		log.Println("Connected to server")
	}

	wsClient.OnConnectError = func(err error, socket web.WSocket) {
		close(done)
		log.Println("Received connect error %v", err)
	}

	wsClient.OnTextMessage = func(message string, socket web.WSocket) {
		log.Println("Received message" + message)
	}

	wsClient.OnBinaryMessage = func(data []byte, socket web.WSocket) {
		log.Println("Received binary data " + string(data))
	}

	wsClient.OnPingReceived = func(data string, socket web.WSocket) {
		log.Println("Received ping " + data)
	}

	wsClient.OnPongReceived = func(data string, socket web.WSocket) {
		log.Println("Received pong " + data)
	}

	wsClient.OnDisconnected = func(err error, socket web.WSocket) {
		close(done)
		log.Println("Disconnected from server")
		return
	}

	wsClient.Connect()
	if wsClient.IsConnected == false {
		log.Println("Could not connect ws server:ws://10.153.90.4:8080/echo")
		os.Exit(1)
	}

	interrupt := make(chan os.Signal, 2)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			wsClient.SendBinary([]byte(t.String()))
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			wsClient.SendBinary(websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			wsClient.Close()
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
