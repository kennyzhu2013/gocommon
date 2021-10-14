package log

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	los "os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type output struct {
	err 		error
	dirs 		string
	file   		*los.File
	mu   		sync.Mutex

	// record the current log file size
	current		int32

	outputs 	OutputOptions
	asyncOption AsyncOptions

	pool		*pool
	rollCh    	chan bool
	startRoll 	sync.Once
}

// logInfo is a convenience struct to return the filename and its embedded timestamp.
type logInfo struct {
	sequence int
	los.FileInfo
}

// byFormatTime sorts by newest time formatted in the name.
type byFormatSequence []logInfo

func (b byFormatSequence) Less(i, j int) bool {
	return b[i].sequence > b[j].sequence
}

func (b byFormatSequence) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byFormatSequence) Len() int {
	return len(b)
}

// 用 seps 进行分割, 根据协议栈信息查找...
func GetFunctionName(i interface{}, seps ...rune) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})

	if size := len(fields); size > 0 {
		return fields[size - 1]
	}
	return ""
}

func NewOutputOptions(opts ...OutputOption) OutputOptions {
	var options OutputOptions
	for _, opt := range opts {
		opt(&options)
	}
	if len(options.Name) == 0 {
		options.Name = DefaultOutputName
	}
	return options
}

func NewAsyncOptions(opts ...AsyncOption) AsyncOptions {
	var options AsyncOptions
	for _, opt := range opts {
		opt(&options)
	}
	if options.Enabled {
		if options.PoolSize == 0 {
			options.PoolSize = DefaultPoolSize
		}
		if options.BufferSize == 0 {
			options.BufferSize = DefaultBufferSize
		}
		if options.FlushNum == 0 {
			options.FlushNum = DefaultFlushNum
		}
		if options.WriteTimeout == 0 {
			options.WriteTimeout = DefaultWriteTimeout
		}
		if options.FlushInterval == 0 {
			options.FlushInterval = DefaultFlushInterval
		}
	}
	return options
}

func NewOutput(opts ...OutputOption) Output {
	options := NewOutputOptions(opts...)
	file, err := openLogFile(options)
	if err != nil {
		panic(err.Error())
	}
	current := refreshFileSize(file)
	return &output{
		outputs: options,
		err:  	 err,
		file:    file,
		dirs: filepath.Dir(file.Name()),
		current: current,
	}
}

func NewOutput2(outputs OutputOptions, asyncOption AsyncOptions) Output {
	file, err := openLogFile(outputs)
	if err != nil {
		panic(err.Error())
	}
	current := refreshFileSize(file)

	if asyncOption.Enabled {
		o := &output{
			outputs: outputs,
			asyncOption: asyncOption,
			pool: 	newPool(asyncOption.PoolSize, asyncOption.BufferSize),
			err:	err,
			file:	file,
			dirs:	filepath.Dir(file.Name()),
			current: current,
		}
		// start read goroutines for processing the log events
		go o.process()
		return o
	}
	return &output{
		outputs: outputs,
		err:	err,
		file:	file,
		dirs:	filepath.Dir(file.Name()),
		current: current,
	}
}

func (o *output) Send(e *Event, force bool) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.file == nil {
		return o.err
	}

	var err error
	if o.current >= FileSize {
		// judge is full
		// rotate directly
		if err = o.rollRunOnce(); err != nil {
			_ = fmt.Errorf("rollRunOnce failed: %s", err)
		}
		o.current = 0
	}

	if o.pool == nil || force {
		// write directly
		// like:[2019-07-23 16:32:55,104] [BillPusher] [INFO] - com.cmic.Pusher.BillPusher.pushBill(BillPusher.java:120) - [45730985705483] - pushed bill success!
		_, err = o.file.Write(generateLogInfo(e))
	} else {
		var timeout bool
		err, timeout = o.pool.enc(e, o.asyncOption.WriteTimeout)
		if timeout {
			_, err = o.file.Write(generateLogInfo(e))
		}
	}
	if err == nil {
		o.current += 1
	}
	return err
}

func (o *output) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.close()
}

func (o *output) String() string {
	return o.outputs.Name
}

// openNew opens a new log file for writing, moving any old log file out of the
// way.  This methods assumes the file has already been closed.
func (o *output) openNew() error {
	file, err := los.OpenFile(o.dirs + "/" +o.outputs.Name, los.O_CREATE|los.O_APPEND|los.O_WRONLY, 0666)
	o.file = file
	return err
}

// if file is full, move..
func (o *output) rotate() error {
	// sync run.
	o.startRoll.Do(func() {
		o.rollCh = make(chan bool, 1)
		go o.rollRun()
	})
	select {
	case o.rollCh <- true:
	default:
	}
	return nil
}

func (o *output) rollRun() {
	for range o.rollCh {
		_ = o.rollRunOnce()
	}
}

func (o *output) rollRunOnce() error {
	files, err := o.oldLogFiles()
	if err != nil || len(files) < 1 {
		return err
	}

	if err := o.close(); err != nil {
		return err
	}

	dir := o.dir()
	for _, f := range files {
		if f.sequence + 1 >= DefaultFileMaxNum {
			errRemove := los.Remove(filepath.Join(dir, f.Name()))
			if err == nil && errRemove != nil {
				err = errRemove
			}
			continue
		}

		errMove := los.Rename(filepath.Join(dir, f.Name()), filepath.Join(dir, o.nextFileName(f.sequence)))
		if err == nil && errMove != nil {
			err = errMove
		}
	}

	// after rotate, then create new.
	if err := o.openNew(); err != nil {
		return err
	}
	return err
}

// oldLogFiles returns the list of backup log files stored in the same
// directory as the current log file, sorted by ModTime
func (o *output) oldLogFiles() ([]logInfo, error) {
	files, err := ioutil.ReadDir(o.dir())
	if err != nil {
		return nil, errors.New("can't read log file directory: " + err.Error())
	}

	var logFiles []logInfo
	prefix := o.outputs.Name
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), prefix) {
			filename := file.Name()
			var sequence = 0
			if len(filename) > len(prefix) {
				ts := filename[len(prefix) + 1:]
				if ts != "" {
					sequence, _ = strconv.Atoi(ts)
				}
			}
			logFiles = append(logFiles, logInfo{sequence, file})
		}
		// error parsing means that the suffix at the end was not generated
		// by lumberjack, and therefore it's not a backup file.
	}
	sort.Sort(byFormatSequence(logFiles))
	return logFiles, nil
}

func (o *output) persist() fullHandle {
	return func (data []byte) error {
		var err error
		if len(data) != 0 {
			_, err = o.file.Write(data)
		}
		return err
	}
}

// consume the log events
func (o *output) process() {
	callBack := o.persist()
	ticker := time.NewTicker(o.asyncOption.FlushInterval)
	for {
		select {
		case <- o.pool.closed:
			ticker.Stop()
			ticker = nil
			return
		case <- ticker.C:
			// reaching flush interval
			// write to out chan directly
			o.pool.flush(callBack)
		case event := <- o.pool.queue:
			if event != nil {
				o.pool.write(generateLogInfo(event), callBack)
			}
		}
	}
}

func (o *output) flush() error {
	if o.file == nil {
		return o.err
	}
	return o.file.Sync()
}

func (o *output) close() error {
	if o.file == nil {
		return o.err
	}
	_ = o.flush()
	err := o.file.Close()
	o.file = nil
	return err
}

func (o *output) nextFileName(i int) string {
	return o.outputs.Name + "." + strconv.Itoa(i + 1)
}

func (o *output) dir() string {
	if len(o.dirs) == 0 {
		o.dirs = filepath.Dir(o.file.Name())
	}
	return o.dirs
}

func pathExists(path string) (bool, error) {
	_, err := los.Stat(path)
	if err == nil {
		return true, nil
	}
	if los.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func generateLogInfo(event *Event) []byte {
	var builder strings.Builder
	builder.Grow(200)
	builder.WriteString("[")
	builder.WriteString(event.Timestamp)
	builder.WriteString("] ")
	builder.WriteString("[")
	builder.WriteString(event.Fields[0][0])
	builder.WriteString("] ")
	builder.WriteString("[")
	builder.WriteString(Levels[event.Level])
	builder.WriteString("] - [log/output.go] - [")
	builder.WriteString(event.Key)
	builder.WriteString("] - ")
	builder.WriteString(event.Message)
	builder.WriteString("\n")
	return []byte(builder.String())
}

func openLogFile(options OutputOptions) (*los.File, error) {
	var logDir string
	if len(options.Dir) != 0 {
		// create if not exist
		if bExist,_ := pathExists(options.Dir); !bExist {
			const shellToUse = "bash"
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd := exec.Command(shellToUse, "-c", fmt.Sprintf("mkdir -p  %s", options.Dir))
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			_ = cmd.Run()
		}
		logDir = options.Dir + "/" + options.Name
	} else {
		logDir = options.Name
	}
	return los.OpenFile(logDir, los.O_CREATE|los.O_APPEND|los.O_WRONLY, 0666)
}

func refreshFileSize(file *los.File) int32 {
	statInfo, _ := file.Stat()
	if statInfo != nil {
		return int32(statInfo.Size() / 1024)
	}
	return 0
}

func SetDefaultOption(opt Option)  {
	_ = defaultLog.SetOption(opt)
}