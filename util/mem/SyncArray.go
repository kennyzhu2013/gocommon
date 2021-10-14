package mem

import (
	"sync"
	"time"
)
// Element is an element of a linked list.
type Element struct {
	BEmpty bool
	When   time.Time

	// The value stored with this element.
	Value interface{}
}

// TODO: add channel ?
type SyncArray struct {
	arrays  []Element
	sync.Mutex
	// size int // todo: set count limits.
	selectKey int
	outDate time.Duration

	// todo :采用带缓存通道提高分配速度...
}

// Attention: used for thousands instances. if too big, not use this.
func NewSyncList(upLimits int, out time.Duration) *SyncArray {
	l := new(SyncArray)
	// l.size = s
	l.arrays = make([]Element, upLimits)
	for _, e := range l.arrays {
		e.Value = nil
		e.BEmpty = true
	}
	l.selectKey = 0
	l.outDate = out
	return l
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *SyncArray) PushBack(v interface{})  {
	l.Lock()
	defer l.Unlock()

	// var dest *Session
	if l.selectKey >= len(l.arrays) {
		l.selectKey = 0
	}

	bNeedLoop := true
	for index := l.selectKey; index < len(l.arrays); index++ {
		e := l.arrays[index]
		if e.BEmpty == true {
			e.Value = v
			e.BEmpty = false
			e.When = time.Now()
			l.selectKey++
			return
		} else if e.Value != nil && bNeedLoop { // e.bEmpty == false
			if time.Since(e.When) > l.outDate {
				e.BEmpty = true
				e.Value = nil
			} else {
				bNeedLoop = false
			}
		}
	}

	bNeedLoop = true
	for index := 0; index < len(l.arrays); index++ {
		e := l.arrays[index]
		if e.BEmpty == true {
			e.Value = v
			e.BEmpty = false
			e.When = time.Now()
			l.selectKey++
			return
		} else if e.Value != nil && bNeedLoop { // e.bEmpty == false
			if time.Since(e.When) > l.outDate {
				e.BEmpty = true
				e.Value = nil
			} else {
				bNeedLoop = false
			}
		}
	}
}

// may be nil
func (l *SyncArray) PopFront() interface{} {
	l.Lock()
	defer l.Unlock()
	for _, e := range l.arrays {
		if e.BEmpty == false {
			e.BEmpty = true
			dst := e.Value
			e.Value = nil
			return dst
		}
	}

	return nil
}

func (l *SyncArray) Clear()  {
	l.Lock()
	defer l.Unlock()
	for _, e := range l.arrays {
		e.Value = nil
		e.BEmpty = true
	}
	l.selectKey = 0
}

// Foreach all items, for test
func (l *SyncArray) Foreach(ops func(selectKey int, item *Element)) {
	l.Lock()
	defer l.Unlock()

	for _, v := range l.arrays {
		ops(l.selectKey, &v)
	}
}
