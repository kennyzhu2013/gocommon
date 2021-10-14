/*
* Copyright(C),2019-2029, email: 277251257@qq.com
* Author:  kennyzhu
* Version: 1.0.0
* Date:    2021/3/19 16:33
* Description:
*
 */
package main

import (
	"common/util/file"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	sourcefile := flag.Arg(0)

	lockfile, err := file.OpenAndLock(sourcefile)
	if err != nil || lockfile == nil {
		fmt.Println("OpenAndLock error:", err.Error())
		return
	}
	_, _ = lockfile.WriteString("helloworld. ")

	notify := make(chan os.Signal, 1)
	signal.Notify(notify, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		select {
		case <-notify:
			_ = lockfile.Close()
			println("file transfer exit!")
			return
		}
	}
}
