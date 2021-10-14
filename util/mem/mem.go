package mem

import (
	"reflect"
	"unsafe"
)

// MemCopy copies size bytes from src to dst.
func MemCopy(src, dst uintptr, size uintptr) {
	if size == 0 {
		return
	}

	srcSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Len:  int(size),
		Cap:  int(size),
		Data: src,
	}))
	dstSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Len:  int(size),
		Cap:  int(size),
		Data: dst,
	}))

	copy(dstSlice, srcSlice)
}
