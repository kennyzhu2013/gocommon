package log

import (
	"fmt"
	"time"
)

type pool struct {
	queue		chan *Event
	// the current indicates the position of the current buffer for pending written
	current		int
	buffer  	[]byte
	encTimeout	*time.Timer
	closed		chan bool
}

type fullHandle func(data []byte) error

func newPool(poolSize, bufferSize int) *pool {
	p := new(pool)
	p.queue = make(chan *Event, poolSize)
	p.buffer = make([]byte, bufferSize)
	p.encTimeout = time.NewTimer(DefaultWriteTimeout)
	p.closed = make(chan bool)
	return p
}

func (p *pool) enc(event *Event, timeout time.Duration) (error, bool) {
	p.encTimeout.Reset(timeout)
	select {
	case <- p.closed:
		return fmt.Errorf("the log event pool is closed, the event is %v", *event), false
	case p.queue <- event:
		return nil, false
	case <- p.encTimeout.C:
		return nil, true
	}
}

func (p *pool) write(data []byte, handle fullHandle) {
	if p.current + len(data) > len(p.buffer) {
		// if the buffer size is reaching to the threshold, callBack
		_ = handle(p.buffer[:p.current])
		p.current = 0
	}
	p.buffer = append(p.buffer[:p.current], data...)
	p.current += len(data)
}

func (p *pool) flush(handle fullHandle) {
	if p.current > 0 {
		_ = handle(p.buffer[:p.current])
		p.current = 0
		p.buffer = p.buffer[:p.current]
	}
}

func (p *pool) close() {
	close(p.closed)
	if p.encTimeout != nil {
		p.encTimeout.Stop()
		p.encTimeout = nil
	}
}