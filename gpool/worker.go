package gpool

import (
	"errors"
	"runtime"
	"time"
)

var (
	// errQueueIsFull will be returned when the worker queue is full.
	errQueueIsFull = errors.New("the queue is full")

	// errQueueIsReleased will be returned when trying to insert item to a released worker queue.
	errQueueIsReleased = errors.New("the queue length is zero")
)

type workerArray interface {
	len() int
	isEmpty() bool
	insert(worker *goWorker) error
	detach() *goWorker
	retrieveExpiry(duration time.Duration) []*goWorker
	reset()
}

type arrayType int

const (
	stackType arrayType = 1 << iota
	loopQueueType
)

func newWorkerArray(aType arrayType, size int) workerArray {
	switch aType {
	case stackType:
		return newWorkerStack(size)
	case loopQueueType:
		return newWorkerLoopQueue(size)
	default:
		return newWorkerStack(size)
	}
}

// goWorker is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and performs function calls.
type goWorker struct {
	// pool who owns this worker.
	pool *Pool

	// 放回队列的回收时间.
	recycleTime time.Time

	// task is a job should be done.
	// default PoolTypeWorker
	// task chan func() // any func.

	// PoolTypeWorkerFunc
	args chan interface{} // args may be func.
}

// run starts a goroutine to repeat the process that performs the function calls.
func (w *goWorker) run() {
	w.pool.incRunning()
	go func() {
		defer func() {
			w.pool.decRunning()
			w.pool.workerCache.Put(w) // crashed.
			if p := recover(); p != nil {
				if ph := w.pool.options.PanicHandler; ph != nil {
					ph(p)
				} else {
					w.pool.options.Logger.Printf("worker exits from a panic: %v\n", p)
					var buf [4096]byte
					n := runtime.Stack(buf[:], false)
					w.pool.options.Logger.Printf("worker exits from panic: %s\n", string(buf[:n]))
				}
			}
			// Call Signal() here in case there are goroutines waiting for available workers.
			w.pool.cond.Signal()
		}()

		for arg := range w.args {
			if arg == nil {
				return
			}
			if w.pool.PoolType == PoolTypeWorkerFunc {
				w.pool.poolFunc(arg)
			} else {
				f := arg.(func())
				if f == nil {
					return
				}
				f()
			}
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}
