package gpool

import "time"

// 循环队列.
type loopQueue struct {
	items  []*goWorker
	expiry []*goWorker
	head   int
	tail   int
	size   int // items预定的size.
	isFull bool
}

func newWorkerLoopQueue(size int) *loopQueue {
	return &loopQueue{
		items: make([]*goWorker, size),
		size:  size,
	}
}

func (wq *loopQueue) len() int {
	if wq.size == 0 {
		return 0
	}

	if wq.head == wq.tail {
		if wq.isFull {
			return wq.size
		}
		return 0
	}

	if wq.tail > wq.head {
		return wq.tail - wq.head
	}

	// tail重置了.
	return wq.size - wq.head + wq.tail
}

func (wq *loopQueue) isEmpty() bool {
	return wq.head == wq.tail && !wq.isFull
}

func (wq *loopQueue) insert(worker *goWorker) error {
	if wq.size == 0 {
		return errQueueIsReleased
	}

	if wq.isFull {
		return errQueueIsFull
	}
	wq.items[wq.tail] = worker
	wq.tail++

	// 达到最大长度.
	if wq.tail == wq.size {
		wq.tail = 0
	}
	if wq.tail == wq.head {
		wq.isFull = true
	}

	return nil
}

// 从头部获取, ..
func (wq *loopQueue) detach() *goWorker {
	if wq.isEmpty() {
		return nil
	}

	w := wq.items[wq.head]
	wq.items[wq.head] = nil
	wq.head++
	if wq.head == wq.size {
		wq.head = 0
	}
	wq.isFull = false

	return w
}

// 循环扫描队列所有元素检查是否超时.
func (wq *loopQueue) retrieveExpiry(duration time.Duration) []*goWorker {
	if wq.isEmpty() {
		return nil
	}

	wq.expiry = wq.expiry[:0]
	expiryTime := time.Now().Add(-duration)

	for !wq.isEmpty() {
		if expiryTime.Before(wq.items[wq.head].recycleTime) {
			break
		}
		wq.expiry = append(wq.expiry, wq.items[wq.head])
		wq.items[wq.head] = nil
		wq.head++
		if wq.head == wq.size {
			wq.head = 0
		}
		wq.isFull = false
	}

	return wq.expiry
}

func (wq *loopQueue) reset() {
	if wq.isEmpty() {
		return
	}

	for w := wq.detach(); w != nil; {
		w.args <- nil
	}
	//Releasing:
	//	if w := wq.detach(); w != nil {
	//		w.task <- nil
	//		goto Releasing
	//	}
	wq.items = wq.items[:0]
	wq.size = 0
	wq.head = 0
	wq.tail = 0
}
