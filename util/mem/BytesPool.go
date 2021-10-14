package mem

import (
	"sync"
)

type BytesPool struct {
	pool sync.Pool
	byteSize int
}

func NewBytesPool(iSize int) (b *BytesPool) {
	b = new(BytesPool)
	b.byteSize = iSize
	b.pool.New = func() interface{} {
		return b.NewBytes()
	}

	return
}

func (b *BytesPool) NewBytes() []byte {
	return make([]byte, b.byteSize)
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (b *BytesPool) AcquireBytes() []byte {
	return b.pool.Get().([]byte)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (b *BytesPool) ReleaseBytes(bs []byte) {
	b.pool.Put(bs)
}
