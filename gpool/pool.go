package gpool

import (
	csync "common/util/sync"
	"sync"
	"sync/atomic"
	"time"
)

// Pool accepts the tasks from client, it limits the total of goroutines to a given number by recycling goroutines.
type Pool struct {
	// Pool的容量，为负值则表示无限容量.
	capacity int32

	// running is the number of the currently running goroutines.
	running int32

	// 工作队列保护锁.
	lock sync.Locker

	// workers is a slice that store the available workers.
	// 采用stack或者loopQueue,存的是goroutine正在运行的worker.
	workers workerArray

	// pool的状态.
	state int32

	// cond for waiting to get a idle worker.
	cond *sync.Cond

	// workerCache speeds up the obtainment of the an usable worker in function:retrieveWorker.
	// workerCache中的 goworker run还没开启.
	workerCache sync.Pool // Pool支持并发，因此是用地方不用加锁.

	// blockingNum is the number of the goroutines already been blocked on pool.Submit, protected by pool.lock
	blockingNum int

	options *Options

	PoolType
	// poolFunc is the function for processing tasks, 类型为PoolTypeWorkerFunc情况下有效.
	poolFunc func(interface{})
}

// purgePeriodically clears expired workers periodically which runs in an individual goroutine, as a scavenger.
func (p *Pool) purgePeriodically() {
	heartbeat := time.NewTicker(p.options.ExpiryDuration)
	defer heartbeat.Stop()

	// 定时tick清理.
	for range heartbeat.C {
		if p.IsClosed() {
			break
		}

		p.lock.Lock()
		// expiredWorkers为scavengers.
		expiredWorkers := p.workers.retrieveExpiry(p.options.ExpiryDuration)
		p.lock.Unlock()

		// Notify obsolete workers to stop.
		// This notification must be outside the p.lock, since w.task
		// may be blocking and may consume a lot of time if many workers are located on non-local CPUs.
		for i := range expiredWorkers {
			expiredWorkers[i].args <- nil // goroutine will quit.
			expiredWorkers[i] = nil
		}

		// There might be a situation that all workers have been cleaned up(no any worker is running)
		// while some invokers still get stuck in "p.cond.Wait()", then it ought to wake all those invokers.
		// 之前正在wait的有空余的没来得及通知，此时延时通知那些延时了的invokers(保持完整性).
		if p.Running() == 0 {
			p.cond.Broadcast()
		}
	}
}

// NewPool generates an instance of ants pool.
func NewPool(size int, options ...Option) (*Pool, error) {
	opts := loadOptions(options...)

	if size <= 0 {
		size = -1
	}

	if expiry := opts.ExpiryDuration; expiry < 0 {
		return nil, ErrInvalidPoolExpiry
	} else if expiry == 0 {
		opts.ExpiryDuration = DefaultCleanIntervalTime
	}

	if opts.Logger == nil {
		opts.Logger = defaultLogger
	}

	p := &Pool{
		capacity: int32(size),
		lock:     csync.NewSpinLock(),
		options:  opts,
		PoolType: PoolTypeWorker,
	}
	p.workerCache.New = func() interface{} {
		return &goWorker{
			pool: p,
			args: make(chan interface{}, workerChanCap),
		}
	}
	if p.options.PreAlloc {
		if size == -1 {
			return nil, ErrInvalidPreAllocSize
		}
		p.workers = newWorkerArray(loopQueueType, size) // 预定大小.
	} else {
		p.workers = newWorkerArray(stackType, 0)
	}

	p.cond = sync.NewCond(p.lock)

	// Start a goroutine to clean up expired workers periodically.
	go p.purgePeriodically()

	return p, nil
}

// NewPoolWithFunc generates an instance of ants pool with a specific function.
func NewPoolWithFunc(size int, pf func(interface{}), options ...Option) (*Pool, error) {
	if size <= 0 {
		size = -1
	}

	if pf == nil {
		return nil, ErrLackPoolFunc
	}

	opts := loadOptions(options...)

	if expiry := opts.ExpiryDuration; expiry < 0 {
		return nil, ErrInvalidPoolExpiry
	} else if expiry == 0 {
		opts.ExpiryDuration = DefaultCleanIntervalTime
	}

	if opts.Logger == nil {
		opts.Logger = defaultLogger
	}

	p := &Pool{
		capacity: int32(size),
		poolFunc: pf,
		lock:     csync.NewSpinLock(),
		options:  opts,
		PoolType: PoolTypeWorkerFunc,
	}
	p.workerCache.New = func() interface{} {
		return &goWorker{
			pool: p,
			args: make(chan interface{}, workerChanCap),
		}
	}
	if p.options.PreAlloc {
		if size == -1 {
			return nil, ErrInvalidPreAllocSize
		}
		p.workers = newWorkerArray(loopQueueType, size) // 预定大小.
	} else {
		p.workers = newWorkerArray(stackType, 0)
	}
	p.cond = sync.NewCond(p.lock)

	// Start a goroutine to clean up expired workers periodically.
	go p.purgePeriodically()

	return p, nil
}

// ---------------------------------------------------------------------------

// Submit submits a task to this pool.
func (p *Pool) Submit(task func()) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}
	var w *goWorker
	if w = p.retrieveWorker(); w == nil {
		return ErrPoolOverload
	}
	w.args <- task
	return nil
}

// Invoke submits a task to pool.
func (p *Pool) Invoke(args interface{}) error {
	if p.IsClosed() {
		return ErrPoolClosed
	}
	var w *goWorker
	if w = p.retrieveWorker(); w == nil {
		return ErrPoolOverload
	}
	w.args <- args
	return nil
}

// Running returns the number of the currently running goroutines.
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// Free returns the available goroutines to work, -1 indicates this pool is unlimited.
func (p *Pool) Free() int {
	c := p.Cap()
	if c < 0 {
		return -1
	}
	return c - p.Running()
}

// Cap returns the capacity of this pool.
func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

// Tune changes the capacity of this pool, note that it is noneffective to the infinite or pre-allocation pool.
func (p *Pool) Tune(size int) {
	if capacity := p.Cap(); capacity == -1 || size <= 0 || size == capacity || p.options.PreAlloc {
		return
	}
	atomic.StoreInt32(&p.capacity, int32(size))
}

// IsClosed indicates whether the pool is closed.
func (p *Pool) IsClosed() bool {
	return atomic.LoadInt32(&p.state) == CLOSED
}

// Release closes this pool and releases the worker queue.
func (p *Pool) Release() {
	atomic.StoreInt32(&p.state, CLOSED)
	p.lock.Lock()
	p.workers.reset()
	p.lock.Unlock()
	// There might be some callers waiting in retrieveWorker(), so we need to wake them up to prevent
	// those callers blocking infinitely.
	p.cond.Broadcast()
}

// Reboot reboots a closed pool.
func (p *Pool) Reboot() {
	if atomic.CompareAndSwapInt32(&p.state, CLOSED, OPENED) {
		go p.purgePeriodically()
	}
}

// ---------------------------------------------------------------------------

// incRunning increases the number of the currently running goroutines.
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// retrieveWorker returns a available worker to run the tasks.
func (p *Pool) retrieveWorker() (w *goWorker) {
	spawnWorker := func() {
		w = p.workerCache.Get().(*goWorker) // runnable worker.
		w.run()
	}

	p.lock.Lock()

	w = p.workers.detach()
	if w != nil { // first try to fetch the worker from the queue
		p.lock.Unlock()
	} else if capacity := p.Cap(); capacity == -1 || capacity > p.Running() {
		// if the worker queue is empty and we don't run out of the pool capacity, then just spawn a new worker goroutine.
		p.lock.Unlock()
		spawnWorker()
	} else {
		if p.options.Nonblocking { // 非阻塞情况下没有空余goWork直接返回.
			p.lock.Unlock()
			return
		}

		// 阻塞情况下无限等待准入条件.
		for {
			if p.options.MaxBlockingTasks != 0 && p.blockingNum >= p.options.MaxBlockingTasks { // 超过阻塞的最大协程.
				p.lock.Unlock()
				return
			}
			p.blockingNum++
			p.cond.Wait() // block and wait for an available worker
			p.blockingNum--
			var nw int
			if nw = p.Running(); nw == 0 { // awakened by the scavenger
				p.lock.Unlock()
				if !p.IsClosed() {
					spawnWorker() // awakened by broadcast.
				}
				return
			}
			if w = p.workers.detach(); w == nil { // 先试用现有的协程.
				if nw < capacity {
					p.lock.Unlock()
					spawnWorker() // 回收goWorker.
					return
				}
				continue //没抢到继续争抢.
			}
			p.lock.Unlock()
			return
		}
	}
	return
}

// revertWorker puts a worker back into free pool, recycling the goroutines.
func (p *Pool) revertWorker(worker *goWorker) bool {
	if capacity := p.Cap(); (capacity > 0 && p.Running() > capacity) || p.IsClosed() {
		return false
	}
	worker.recycleTime = time.Now()
	p.lock.Lock()

	// 防止资源泄露先判断是否关闭.
	if p.IsClosed() {
		p.lock.Unlock()
		return false
	}

	err := p.workers.insert(worker)
	if err != nil {
		p.lock.Unlock()
		return false
	}

	// Notify the invoker stuck in 'retrieveWorker()' of there is an available worker in the worker queue.
	p.cond.Signal()
	p.lock.Unlock()
	return true
}
