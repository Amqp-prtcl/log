package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type FileOutput struct {
	add        int
	f          *os.File
	flags      int
	logLevel   LogLevel
	outputType OutputType
	LogFunc    func(*[]byte, *LogEntry, int, io.Writer) error
}

// NewFileOutput opens or creates a file either appending or truncating it and returns it as an Output.
// It automatically closes file with 'LogClose()'.
//
// date arg specifies wether to add date and time before file name
//
// If file is closed in any way other than by output, the output is automatically removed from each log manager
// it is attached on next log call by such.
//
// Output type default to Text.
//
// NOTE: if for any reason FileOutput is not added to any logger, it is caller's responsibility to call LogClose once.
func NewFileOutput(path string, date bool, flags int, logLevel LogLevel, outputType OutputType, append bool) (Output, error) {
	var osFlags int
	if append {
		osFlags = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	} else {
		osFlags = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	}
	if date {
		t := time.Now()
		dir, name := filepath.Split(path)
		path = filepath.Join(dir, fmt.Sprintf("%v/%v/%v %v:%v:%v ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())+name)
	}
	f, err := os.OpenFile(path, osFlags, 0664)
	if err != nil {
		return nil, err
	}

	output := &FileOutput{
		f:        f,
		flags:    flags,
		logLevel: logLevel,
	}

	switch outputType {
	case T_Text:
		output.LogFunc = TextLogFunc
		output.outputType = T_Text
	case T_JSON:
		output.LogFunc = JSONLogFunc
		output.outputType = T_JSON
	default:
		output.LogFunc = TextLogFunc
		output.outputType = T_Text
	}

	return output, nil
}

func (o *FileOutput) Log(entry *LogEntry) error {
	if o.logLevel.Permits(entry.Level) {
		var e error
		buf, ok := entry.GetCompiled(o.flags, o.outputType)
		if !ok {
			buf = entry.GetBuf()
			e = o.LogFunc(buf, entry, o.flags, o.f)
		} else {
			_, e = o.f.Write(*buf)
		}
		if errors.Is(e, ErrOutputClosed) {
			e = ErrOutputClosed
		}
		return e
	}
	return nil
}

func (o *FileOutput) OnAdd() {
	o.add++
}

func (o *FileOutput) GetFlags() int {
	return o.flags
}

func (o *FileOutput) SetFlags(flags int) {
	o.flags = flags
}

func (o *FileOutput) SetLogLevel(logLevel LogLevel) {
	o.logLevel = logLevel
}

func (o *FileOutput) GetLogLevel() LogLevel {
	return o.logLevel
}

func (o *FileOutput) GetOutputType() OutputType {
	return o.outputType
}

func (o *FileOutput) LogClose() error {
	if o.add > 1 {
		o.add--
		return nil
	}
	return o.f.Close()
}
