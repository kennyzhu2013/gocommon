package log

import (
	"context"
	"time"
)

type Options struct {
	// the current log level
	Level Level

	// the output to write to
	Output

	// include a set of fields
	Fields Fields
	// size of one log file

	FileSize int

	// Alternative options
	Context context.Context
}

func WithLevel(l Level) Option {
	return func(o *Options) {
		o.Level = l
	}
}

func WithFields(f Fields) Option {
	return func(o *Options) {
		o.Fields = f
	}
}

// Output options
func WithOutput(ot Output) Option {
	return func(o *Options) {
		o.Output = ot
	}
}

func OutputName(name string) OutputOption {
	return func(o *OutputOptions) {
		o.Name = name
	}
}

func OutputDir(dir string) OutputOption {
	return func(o *OutputOptions) {
		o.Dir = dir
	}
}

// Async options
func EnableAsync(enabled bool) AsyncOption {
	return func(a *AsyncOptions) {
		a.Enabled = enabled
	}
}

func PoolSize(poolSize int) AsyncOption {
	return func(a *AsyncOptions) {
		a.PoolSize = poolSize
	}
}

func BufferSize(bufferSize int) AsyncOption {
	return func(a *AsyncOptions) {
		a.BufferSize = bufferSize
	}
}

func FlushNum(flushNum int) AsyncOption {
	return func(a *AsyncOptions) {
		a.FlushNum = flushNum
	}
}

func WriteTimeout(timeout time.Duration) AsyncOption {
	return func(a *AsyncOptions) {
		a.WriteTimeout = timeout
	}
}

func FlushInterval(flushInterval time.Duration) AsyncOption {
	return func(a *AsyncOptions) {
		a.FlushInterval = flushInterval
	}
}
