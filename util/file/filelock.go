/*
@Time : 2019/7/4 9:44
@Author : kenny zhu
@File : filelock
@Software: GoLand
@Others:
*/
package file

import (
	"errors"
	"os"
	"syscall"
)

type LockFile struct {
	path string    // full path, eg: /home/XXX/go/src or /home/XXX/go/src/a.txt
	*os.File
}

// fullPath is a file or dir
func NewFileLock(fullPath string) *LockFile {
	return &LockFile{
		path: fullPath,
	}
}

// fullPath is a file or dir
func OpenAndLock(fullPath string) (*LockFile, error) {
	file := &LockFile{
		path: fullPath,
	}

	return  file, file.Open()
}

// lock ops
func (l *LockFile) Open() error {
	f, err := os.OpenFile( l.path, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0755 ) // 获取文件描述符
	if err != nil {
		return err
	}
	l.File = f
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) // 加上排他锁，当遇到文件加锁的情况直接返回 Error
	if err != nil {
		return errors.New("cannot flock directory " + l.path)
	}
	return nil
}

func (l *LockFile) GetFile() *os.File {
	return l.File
}

// 解锁操作
func (l *LockFile) Close() error {
	err := syscall.Flock(int(l.File.Fd()), syscall.LOCK_UN)
	_ = l.File.Close()
	return  err
}