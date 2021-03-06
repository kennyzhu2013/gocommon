/*
@Time : 2019/6/10 12:00
@Author : kenny zhu
@File : test_ping
@Software: GoLand
@Others:
*/
package main

import (
	"common/log/log"
	"common/util/net"
	"flag"
	"os"
	"time"
)

func main() {
	// init log here.
	log.InitLogger(
		log.WithLevel(log.Level(1)),
		log.WithFields(log.Fields{
			0: {"logger test_ping"},
		}),
		log.WithOutput(
			log.NewOutput(log.OutputDir("./"), log.OutputName("all.log")),
		),
	)
	urls := []string{"117.161.4.159", "callback.pstn.avc.qcloud.com", "www.ym,56.com", "ljbk.tortoiseapi.ucc360.com", "140.143.181.12", "hdh.1ketong.com"}
	for {
		for i := 0; i < len(urls); i++ {
			var count int
			var timeout int64
			var size int
			var neverstop bool

			// default command, eg: -n 2 www.baidu.com.
			flag.Int64Var(&timeout, "w", 1000, "等待每次回复的超时时间(毫秒)。")
			flag.IntVar(&count, "n", 4, "要发送的回显请求数。")
			flag.IntVar(&size, "l", 32, "要发送缓冲区大小。")
			flag.BoolVar(&neverstop, "t", false, "Ping 指定的主机，直到停止。")

			flag.Parse()
			args := flag.Args()

			if len(args) > 1 {
				//fmt.Println("Usage: ", os.Args[0], "host")
				//flag.PrintDefaults()
				//flag.Usage()
				//os.Exit(1)
				urls = args
			}

			argsmap := map[string]interface{}{}

			argsmap["w"] = timeout
			argsmap["n"] = count
			argsmap["l"] = size
			argsmap["t"] = neverstop

			for _, host := range urls {
				result := net.Ping(host, argsmap)
				net.Stat(result)
			}
		}
		time.Sleep(time.Second)
	}

	// for i := 0; i < len(args); i++ {
	// 	<-ch
	// }
	// go net.Ping(host, ch, argsmap)

	os.Exit(0)
}
