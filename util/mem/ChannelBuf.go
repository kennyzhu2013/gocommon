package mem


// // todo :采用带缓存通道提高分配速度...
type ChannelBuf struct {
	freeList  chan interface{}
}

// Attention: used for thousands instances. if too big, not use this.
func NewChannelBuf(upLimits int) *ChannelBuf {
	l := new(ChannelBuf)

	// l.size = s
	l.freeList = make(chan interface{}, upLimits)
	return l
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *ChannelBuf) PushBack(v interface{}) {
	select {
	case l.freeList <- v: // Packet on free list; nothing more to do.
	default: // Free list full, just carry on.
	}
}

//// may be nil, must create outside.
func (l *ChannelBuf) PopFront() (v interface{}) {
	// Grab a packet if available; allocate if not.
	select {
	case v = <-l.freeList: // Got one; nothing more to do.
	default:
	}
	return nil
}
//
func (l *ChannelBuf) Clear()  {
	// nothing todo or clear ?....

}
