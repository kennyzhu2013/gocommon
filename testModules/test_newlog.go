package main

import (
	log "common/log/newlog"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"
)

func init()  {
	log.InitLogger(
		log.WithLevel(log.InfoLevel),
		log.WithOutput(
			log.NewOutput2(
				log.NewOutputOptions(log.OutputDir(""), log.OutputName("test.log")),
				log.NewAsyncOptions(log.EnableAsync(true), log.WriteTimeout(10 * time.Millisecond)),
			),
		),
	)
}

var stopped = make(chan os.Signal)
var cpuProfile = flag.String("cpu", "cpu.prof", "write cpu profile to file")
var wNumPerG = flag.Int("num", 100000, "the number of logs written per goroutine")
var wGoroutines = flag.Int("wg", 100, "the number of concurrent writing goroutines")

func main() {

	flag.Parse()
	if len(*cpuProfile) != 0 {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fmt.Printf("create the cpu pprof file failed: %v\n", err.Error())
			return
		}
		err = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	signal.Notify(stopped, os.Kill, os.Interrupt)

	fmt.Println("test new log start...")

	testMultiGoroutineLog(*wGoroutines, *wNumPerG)

	<- stopped
	fmt.Println("test new log end...")
}

func testSingleGoroutineLog(id, times int)  {
	for i := 0; i < times; i++ {
		log.Infof("G %v test new log: %v", id, i)
	}
}

func testMultiGoroutineLog(num, times int) {
	for i := 0; i < num; i++ {
		go testSingleGoroutineLog(i, times)
	}
}
