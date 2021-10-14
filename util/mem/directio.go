package mem

import (
	"errors"
	"io"
	"os"
	"unsafe"
)

const (
	// O_DIRECT alignment is 512B
	blockSize = 512

	// Default buffer is 8KB (2 pages).
	defaultBufSize = 8192
)

// align returns an offset for alignment.
func align(b []byte) int {
	return int(uintptr(unsafe.Pointer(&b[0])) & uintptr(blockSize-1))
}

// allocAlignedBuf allocates buffer that is aligned by blockSize.
func allocAlignedBuf(n int) ([]byte, error) {
	if n == 0 {
		return nil, errors.New("size is `0` can't allocate buffer")
	}

	// Allocate memory buffer
	buf := make([]byte, n+blockSize)

	// First memmory alignment
	a1 := align(buf)
	offset := 0
	if a1 != 0 {
		offset = blockSize - a1
	}

	buf = buf[offset : offset+n]

	// Was alredy aligned. So just exit
	if a1 == 0 {
		return buf, nil
	}

	// Second alignment – check and exit
	a2 := align(buf)
	if a2 != 0 {
		return nil, errors.New("can't allocate aligned buffer")
	}

	return buf, nil
}

// DirectIO bypasses page cache.
type DirectIO struct {
	f   *os.File
	buf *SlabBuffer
	// n   int
	err error
}

// NewSize returns a new DirectIO writer.
func NewSize(f *os.File, slab *SlabBuffer) (*DirectIO, error) {
	if err := checkDirectIO(f.Fd()); err != nil {
		return nil, err
	}

	return &DirectIO{
		buf: slab,
		f:   f,
	}, nil
}

// flush writes buffered data to the underlying os.File.
func (d *DirectIO) flush() (int, error) {
	if d.err != nil {
		return 0, d.err
	}

	length := d.buf.BytesUnreadLength()
	n, err := d.f.Write(d.buf.BytesUnread())
	if n < length && err == nil {
		err = io.ErrShortWrite
	}

	return n, err
}

// Flush writes buffered data to the underlying file.
func (d *DirectIO) Flush() (n int, err error) {
	fd := d.f.Fd()

	// Disable direct IO
	err = setDirectIO(fd, false)
	if err != nil {
		return 0, err
	}

	// Making write without alignment
	n, err = d.flush()
	if err != nil {
		return
	}

	// Enable direct IO back
	err = setDirectIO(fd, true)
	return
}

func (d *DirectIO) SetDirectIO(bOpen bool) error {
	return setDirectIO(d.f.Fd(), bOpen)
}

// Available returns how many bytes are unused in the buffer.
func (d *DirectIO) Available() int { return d.buf.Len() }

// Buffered returns the number of bytes that have been written into the current buffer.
func (d *DirectIO) Buffered() int { return d.buf.BytesUnreadLength() }

// not safe when the write is short.
func (d *DirectIO) WriteUnsafe(p []byte) (nn int, err error) {
	return d.buf.WriteUnsafeDirectly(p)
}
