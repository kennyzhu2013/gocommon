package log

import (
	"fmt"
	"time"

	"context"
)

type os struct {
	opts Options
}

type logFunc func(l Level, f Fields, m string) error

func newOS(opts ...Option) Log {
	o := new(os)
	_ = o.Init(opts...)
	return o
}

func (o *os) Init(opts ...Option) error {
	o.opts.Level = DefaultLevel
	o.opts.Context = context.TODO()
	for _, opt := range opts {
		opt(&o.opts)
	}

	if o.opts.Output == nil {
		panic("The log output options are not set")
	}
	return nil
}

func (o *os) SetOption(opt Option) error {
	opt(&o.opts)
	return nil
}

func (o *os) Options() Options {
	return o.opts
}

func (o *os) String() string {
	return "os"
}

func (o *os) Debug(args ...interface{}) {
	_ = o.log(DebugLevel, "", fmt.Sprint(args...))
}

func (o *os) Info(args ...interface{}) {
	_ = o.log(InfoLevel, "", fmt.Sprint(args...))
}

func (o *os) Warn(args ...interface{}) {
	_ = o.log(WarnLevel, "", fmt.Sprint(args...))
}
func (o *os) Error(args ...interface{}) {
	_ = o.log(ErrorLevel, "", fmt.Sprint(args...))
}

func (o *os) Fatal(args ...interface{}) {
	_ = o.log(FatalLevel, "", fmt.Sprint(args...))
}

func (o *os) Sys(args ...interface{}) {
	_ = o.logForce(InfoLevel, "", fmt.Sprint(args...))
}

func (o *os) Debugf(format string, args ...interface{}) {
	_ = o.log(DebugLevel, "", fmt.Sprintf(format, args...))
}

func (o *os) Infof(format string, args ...interface{}) {
	_ = o.log(InfoLevel, "", fmt.Sprintf(format, args...))
}

func (o *os) Printf(format string, args ...interface{}) {
	_ = o.log(InfoLevel, "", fmt.Sprintf(format, args...))
}

func (o *os) Warnf(format string, args ...interface{}) {
	_ = o.log(WarnLevel, "", fmt.Sprintf(format, args...))
}

func (o *os) Errorf(format string, args ...interface{}) {
	_ = o.log(ErrorLevel, "", fmt.Sprintf(format, args...))
}

func (o *os) Fatalf(format string, args ...interface{}) {
	_ = o.log(FatalLevel, "", fmt.Sprintf(format, args...))
}

func (o *os) DebugK(key string, format string, args ...interface{}) {
	_ = o.log(DebugLevel, key, fmt.Sprintf(format, args...))
}

func (o *os) InfoK(key string, format string, args ...interface{}) {
	_ = o.log(InfoLevel, key, fmt.Sprintf(format, args...))
}

func (o *os) WarnK(key string, format string, args ...interface{}) {
	_ = o.log(WarnLevel, key, fmt.Sprintf(format, args...))
}

func (o *os) ErrorK(key string, format string, args ...interface{}) {
	_ = o.log(ErrorLevel, key, fmt.Sprintf(format, args...))
}

func (o *os) FatalK(key string, format string, args ...interface{}) {
	_ = o.log(FatalLevel, key, fmt.Sprintf(format, args...))
}

func (o *os) Log(level Level, args ...interface{}) {
	_ = o.log(level, "", fmt.Sprint(args...))
}

func (o *os) Logf(level Level, format string, args ...interface{}) {
	_ = o.log(level, "", fmt.Sprintf(format, args...))
}

func (o *os) WithFields(f Fields) Logger {
	options := o.opts

	for k, v := range o.opts.Fields {
		options.Fields[k] = v
	}

	for k, v := range f {
		options.Fields[k] = v
	}

	return &os{
		opts: options,
	}
}

func (o *os) log(level Level, key string, msg string) error {
	// discard if we're not at the right level
	if level < o.opts.Level {
		return nil
	}

	e := &Event{
		Timestamp: time.Now().Format("2006-1-2 15:04:05,000"),
		Level:     level,
		Key:       key,
		Fields:    o.opts.Fields,
		Message:   msg,
	}
	if err := o.opts.Output.Send(e, false); err != nil {
		return err
	}
	return nil
}

func (o *os) logForce(level Level, key string, msg string) error {
	// discard if we're not at the right level
	if level < o.opts.Level {
		return nil
	}

	e := &Event{
		Timestamp: time.Now().Format("2006-1-2 15:04:05,000"),
		Level:     level,
		Key:       key,
		Fields:    o.opts.Fields,
		Message:   msg,
	}
	if err := o.opts.Output.Send(e, true); err != nil {
		return err
	}
	return nil
}
