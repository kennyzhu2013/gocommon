// Package log is an interface for structured logging.
package log

import (
	"encoding/json"
	"time"
)

type Level int32

const (
	DebugLevel Level = 0
	InfoLevel  Level = 1
	WarnLevel  Level = 2
	ErrorLevel Level = 3
	FatalLevel Level = 4

	DefaultLevel      = InfoLevel
	DefaultOutputName = "server.log"
	DefaultFileSize   = 512 * 1000

	DefaultPoolSize      = 100
	DefaultFlushNum      = 1000
	DefaultBufferSize    = 1024 * 1000 * 2
	DefaultWriteTimeout  = 10 * time.Millisecond
	DefaultFlushInterval = 5000 * time.Millisecond
)

var DefaultFileMaxNum = 50

// for simple realize
type Fields [1][1]string

// A structure log interface which can output to multiple back-ends.
type Log interface {
	Logger
	Init(opts ...Option) error
	Options() Options
	SetOption(opt Option) error
	String() string
}

type Logger interface {
	// Logger interface
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Sys(args ...interface{})

	// Formatted logger
	Debugf(format string, args ...interface{})
	Printf(format string, args ...interface{}) // 兼容标准IO.
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	// for elk key
	// Formatted logger
	DebugK(key string, format string, args ...interface{})
	InfoK(key string, format string, args ...interface{})
	WarnK(key string, format string, args ...interface{})
	ErrorK(key string, format string, args ...interface{})
	FatalK(key string, format string, args ...interface{})

	// Specify your own levels
	Log(l Level, args ...interface{})
	Logf(l Level, format string, args ...interface{})
	// Returns with extra fields
	WithFields(f Fields) Logger
}

// Event represents a single log event
type Event struct {
	Timestamp string `json:"timestamp"`
	Level     Level  `json:"level"`
	Key       string `json:"key"`
	Fields    Fields `json:"fields"`
	Message   string `json:"message"`
}

// An output represents a file, indexer, syslog, etc
type Output interface {
	// Send an event
	Send(*Event, bool) error

	// Discard the output
	Close() error

	// Name of output
	String() string
}

type Option func(o *Options)

type OutputOption func(o *OutputOptions)

type OutputOptions struct {
	// file path, url, etc, Dir default is ""
	Name string
	Dir  string
}

type AsyncOption func(a *AsyncOptions)

type AsyncOptions struct {
	Enabled  bool
	PoolSize int

	// log infos buffer
	BufferSize int

	// the number of flushing log infos for persisting
	FlushNum int

	WriteTimeout  time.Duration
	FlushInterval time.Duration
}

var (
	// file sequences.
	FileSize int32 = DefaultFileSize

	Levels = map[Level]string{
		DebugLevel: "debug",
		InfoLevel:  "info",
		WarnLevel:  "warn",
		ErrorLevel: "error",
		FatalLevel: "fatal",
	}
)

func NewLog(opts ...Option) Log {
	return newOS(opts...)
}

func (e *Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"level"`
		Fields    Fields `json:"fields"`
		Message   string `json:"message"`
	}{
		Timestamp: e.Timestamp,
		Level:     Levels[e.Level],
		Fields:    e.Fields,
		Message:   e.Message,
	})
}
