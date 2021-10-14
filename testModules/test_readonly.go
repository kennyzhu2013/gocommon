/*
* Copyright(C),2019-2029, email: 277251257@qq.com
* Author:  kennyzhu
* Version: 1.0.0
* Date:    2021/3/19 16:31
* Description:
*
 */
package main

import (
	"syscall"
	"unsafe"
)

type ProtectVar []byte

func newProtectVar(size int) (ProtectVar, error) {
	b, err := syscall.Mmap(0, 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)

	if err != nil {
		return nil, err
	}

	return ProtectVar(b), nil
}

func (p ProtectVar) Free() error {
	return syscall.Munmap([]byte(p))
}

func (p ProtectVar) Readonly() error {
	return syscall.Mprotect([]byte(p), syscall.PROT_READ)
}

func (p ProtectVar) ReadWrite() error {
	return syscall.Mprotect([]byte(p), syscall.PROT_READ|syscall.PROT_WRITE)
}

func (p ProtectVar) Pointer() unsafe.Pointer {
	return unsafe.Pointer(&p[0])
}

func main() {
	pv, err := newProtectVar(8)
	if err != nil {
		println("error:", err)
	}

	defer pv.Free()

	p := (*int)(pv.Pointer())
	*p = 100
	println("1:", *p)

	pv.Readonly()
	println("2:", *p)

	pv.ReadWrite()
	*p += 100
	println("3:", *p)

	pv.Readonly()
	*p++
	println("4:", *p)
}
