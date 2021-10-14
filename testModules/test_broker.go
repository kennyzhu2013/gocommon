/*
* Copyright(C),2019-2029, email: 277251257@qq.com
* Author:  kennyzhu
* Version: 1.0.0
* Date:    2021/3/19 16:34
* Description:
*
 */
package main

import (
	"common/broker"
	"common/rabbitmq"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var (
	storeTopic = "recording.storesession"
	audioTopic = "recording.audio"
	rbroker    = rabbitmq.NewBroker()
)

func pub() {
	tick := time.NewTicker(time.Second)
	i := 0
	for _ = range tick.C {
		var params struct {
			LeftFilename  string
			RightFilename string
		}

		params.LeftFilename = "/tmp/record/z9hG4bK10629d31d11b4syd6sz16s9w9042buu6z@139.120.40.22_early_left.amr"
		params.RightFilename = "/tmp/record/z9hG4bK10629d31d11b4syd6sz16s9w9042buu6z@139.120.40.22_early_right.amr"
		prefersJson, _ := json.Marshal(params)
		msg := &broker.Message{
			Header: map[string]string{
				"id": fmt.Sprintf("%d", i),
			},
			Body: []byte(prefersJson),
		}
		if err := rbroker.Publish(storeTopic, msg); err != nil {
			log.Printf("[pub] failed: %v", err)
			fmt.Printf("[pub] failed: %s", err)
		} else {
			fmt.Println("[pub] pubbed message:", string(msg.Body))
		}
		i++
	}
}

func pub2() {
	tick := time.NewTicker(time.Second)
	i := 0
	for _ = range tick.C {
		var params struct {
			CallId string
		}

		params.CallId = "9852596813833"
		prefersJson, _ := json.Marshal(params)
		msg := &broker.Message{
			Header: map[string]string{
				"id": fmt.Sprintf("%d", i),
			},
			Body: []byte(prefersJson),
		}
		if err := rbroker.Publish(audioTopic, msg); err != nil {
			log.Printf("[pub] failed: %v", err)
			fmt.Printf("[pub] failed: %s", err)
		} else {
			fmt.Println("[pub] pubbed message:", string(msg.Body))
		}
		i++
	}
}
func sub() {
	_, err := rbroker.Subscribe(audioTopic, func(p broker.Publication) error {
		fmt.Println("[sub] received message:", string(p.Message().Body), "header", p.Message().Header)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	// cmd.Init()
	// broker.DefaultBroker = rbroker
	audioTopic += ".10-153-90-11"
	// amqp://admin:cmcc888@10.153.90.11:5672
	if err := rbroker.Init(broker.Addrs("amqp://ziyan:cmcc666@10.153.138.49:5672/zjh")); err != nil {
		log.Fatalf("Broker Init error: %v", err)
		fmt.Printf("Broker Init error:%s", err)
	}
	if err := rbroker.Connect(); err != nil {
		log.Fatalf("Broker Connect error: %v", err)
		fmt.Printf("Broker Connect error:%s", err)
	}

	go pub2()
	go sub()

	<-time.After(time.Second * 600)
}
