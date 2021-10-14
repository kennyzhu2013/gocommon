package main

import (
	"common/util/mem"
	"flag"
	"fmt"
	"github.com/pterm/pterm"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const LOOPCOUNT = 1000

var slabArray *mem.SyncArray
var slabChannel *mem.ChannelBuf
var byteArray [LOOPCOUNT]*mem.SlabBuffer
var byteChannel [LOOPCOUNT]*mem.SlabBuffer

func TestSyncArray() {
	fmt.Printf("TestSyncArray test: count:%d \n cpu:%d ", count, runtime.NumCPU())
	go func() {
		for j := 0; j < count; j++ {
			go func() {
				for i := 0; i < LOOPCOUNT; i++ {
					if byteArray[i] != nil {
						slabArray.PushBack(byteArray[i])
						byteArray[i] = nil
					}

					time.Sleep(time.Millisecond * 20)
				}
			}()
		}
	}()

	for j := 0; j < count; j++ {
		go func() {
			for i := 0; i < LOOPCOUNT; i++ {
				temp := slabArray.PopFront()
				if temp == nil {
					byteArray[i] = mem.NewSlabBuffInit(160 * 1024)
				} else {
					byteArray[i] = slabArray.PopFront().(*mem.SlabBuffer)
				}

				time.Sleep(time.Millisecond * 10)
			}
		}()
	}
}

func TestChannelBuf() {
	fmt.Printf("TestChannelBuf test: count:%d \n", count)
	go func() {
		for j := 0; j < count; j++ {
			for i := 0; i < LOOPCOUNT; i++ {
				if byteChannel[i] != nil {
					slabChannel.PushBack(byteArray[i])
					byteChannel[i] = nil
				}

				time.Sleep(time.Millisecond * 20)
			}
		}
	}()

	for j := 0; j < count; j++ {
		for i := 0; i < LOOPCOUNT; i++ {
			temp := slabChannel.PopFront()
			if temp == nil {
				byteChannel[i] = mem.NewSlabBuffInit(160 * 1024)
			} else {
				byteChannel[i] = slabChannel.PopFront().(*mem.SlabBuffer)
			}

			time.Sleep(time.Millisecond * 10)
		}
	}
}

var count int // false
var isChannel bool

func main() {
	// default command, eg: -w 5000 10.153.90.12 10.153.90.13
	flag.IntVar(&count, "c", 100, "运行秒数(秒)。")
	flag.BoolVar(&isChannel, "b", false, "是否channel测试(秒)。")
	// flag.StringVar(&hostPing, "p", "", "是否开启指定的ping主机。")
	// flag.StringVar(&testFlag, "r", "", "运行模式。")

	flag.Parse()
	// _ = flag.Args()
	if isChannel {
		slabChannel = mem.NewChannelBuf(200)
		go TestChannelBuf()
	} else {
		slabArray = mem.NewSyncList(200, time.Minute*10)
		go TestSyncArray()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	select {
	// wait on kill signal
	case <-ch:
		pterm.BgLightGreen.Print("Exit")
	}
}
