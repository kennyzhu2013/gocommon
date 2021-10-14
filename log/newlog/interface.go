/*
@Time : 2018/8/16 12:46
@Author : kenny zhu
@File : interface.go
@Software: GoLand
@Others: Default operations.
*/
package log

import "fmt"

// logger
var defaultLog Log

func GetLogger() Log {
	return defaultLog
}

func InitLogger(opts ...Option) {
	defaultLog = newOS(opts...)
}

// outer .interface...
func Debug(args ...interface{}) {
	if defaultLog == nil {
		fmt.Print(args)
		return
	}
	defaultLog.Debug(args)
}
func Info(args ...interface{}) {
	if defaultLog == nil {
		fmt.Print(args)
		return
	}
	defaultLog.Info(args...)
}
func Error(args ...interface{}) {
	if defaultLog == nil {
		fmt.Print(args)
		return
	}
	defaultLog.Error(args...)
}

func Warn(args ...interface{}) {
	if defaultLog == nil {
		fmt.Print(args)
		return
	}
	defaultLog.Warn(args...)
}

func Fatal(args ...interface{}) {
	if defaultLog == nil {
		fmt.Print(args)
		return
	}
	defaultLog.Fatal(args...)
}

func Sys(args ...interface{}) {
	if defaultLog == nil {
		fmt.Print(args)
		return
	}
	defaultLog.Sys(args...)
}

// Formatted logger
func Debugf(format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.Infof(format, args...)
}

func Printf(format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.Errorf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.Fatalf(format, args...)
}

// Formatted logger
func DebugK(key string, format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.DebugK(key, format, args...)
}

func InfoK(key string, format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.InfoK(key, format, args...)
}

func WarnK(key string, format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.WarnK(key, format, args...)
}

func ErrorK(key string, format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.ErrorK(key, format, args...)
}
func FatalK(key string, format string, args ...interface{}) {
	if defaultLog == nil {
		fmt.Printf(format, args)
		return
	}
	defaultLog.FatalK(key, format, args...)
}
