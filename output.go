package log

import (
	"encoding/json"
	"io"
)

// Output represents a log writer.
//
// Once registered in a logger, each log call to the logger wil trigger
// a the Log method of the Outputs
type Output interface {
	//it is called each time the output is added to a log manager
	//
	// it is used to know how many close calls there needs to be
	// before actually closing io.Writer
	//
	// NOTE: CLose can be called before any OnAdd call (if manually closed before adding it to any logger)
	OnAdd()

	// error must be related to io
	Log(*LogEntry) error

	GetFlags() int
	SetFlags(int)

	SetLogLevel(LogLevel)
	GetLogLevel() LogLevel

	GetOutputType() OutputType

	// LogClose is called upon close call to Logger, if output
	// should not be closed Close() should do nothing and return nil.
	//
	// Close means that no more log calls will be made from closed log manager
	// but can still potentially be called if output has been added to multiple
	// Loggers
	LogClose() error

	/*// Close is the actual close implementation, if called, the output must be closed
	// regardless of OnAdd() amount
	Close() error*/
}

type outputWrapper struct {
	add int

	w     io.Writer
	close bool

	flags int

	logLevel   LogLevel
	outputType OutputType

	LogFunc   func(*[]byte, *LogEntry, int, io.Writer) error
	CloseFunc func(io.Writer, bool) error
}

// NewOutputWrapper return a wrapper for the logFunc and closeFunc callback with provided parameters
//
// closeFunc is only called once the Output is removed from every log manager it has been added
func NewOutputWrapper(w io.Writer, close bool, flags int, output OutputType, logLevel LogLevel, logFunc func(*[]byte, *LogEntry, int, io.Writer) error, closeFunc func(io.Writer, bool) error) Output {
	return &outputWrapper{
		w:         w,
		close:     close,
		flags:     flags,
		logLevel:  logLevel,
		LogFunc:   logFunc,
		CloseFunc: closeFunc,
	}
}

func (o *outputWrapper) OnAdd() {
	o.add++
}

func (o *outputWrapper) GetFlags() int {
	return o.flags
}

func (o *outputWrapper) SetFlags(f int) {
	o.flags = f
}

func (o *outputWrapper) SetLogLevel(level LogLevel) {
	o.logLevel = level
}

func (o *outputWrapper) GetLogLevel() LogLevel {
	return o.logLevel
}

func (o *outputWrapper) GetOutputType() OutputType {
	return o.outputType
}

func (o *outputWrapper) LogClose() error {
	if o.add > 1 {
		o.add--
		return nil
	}
	if o.CloseFunc != nil {
		return o.CloseFunc(o.w, o.close)
	}
	return nil
}

func (o *outputWrapper) Log(entry *LogEntry) error {
	if o.logLevel.Permits(entry.Level) && o.LogFunc != nil {
		var e error
		buf, ok := entry.GetCompiled(o.flags, o.outputType)
		if !ok {
			buf = entry.GetBuf()
			e = o.LogFunc(buf, entry, o.flags, o.w)
		} else {
			_, e = o.w.Write(*buf)
		}
		return e
	}
	return nil
}

// default close func closes w if w implements io.Closer and if close is true
// otherwise only return nil
func DefaultCloseFunc(w io.Writer, close bool) error {
	if closer, ok := w.(io.Closer); close && ok {
		return closer.Close()
	}
	return nil
}

// JSONLogFunc formats entry with specified flags and writes it to w as a JSON object
// adding buf to entry with TEXT output type
func JSONLogFunc(buf *[]byte, entry *LogEntry, flags int, w io.Writer) error {
	var m = M{}
	if flags&F_Time != 0 {
		m.Add("time", entry.Time)
	}

	if entry.Prefixes != nil && len(entry.Prefixes) != 0 && flags&(F_Prefix|F_LastPrefix) != 0 {
		if flags&F_LastPrefix != 0 {
			m.Add("prefix", entry.Prefixes[len(entry.Prefixes)-1])
		} else {
			m.Add("prefix", append([]string{}, entry.Prefixes...))
		}
	}

	if flags&F_Level != 0 {
		m.Add("level", entry.Level.String())
	}
	if entry.Fields != nil && len(entry.Fields) != 0 && flags&(F_Fields|F_Fields_A|F_Fields_B) != 0 {
		if flags&F_Fields_A != 0 {
			m.Add("fields", entry.Fields)
		} else if flags&F_Fields != 0 {
			m.AddE(entry.Fields...)
		} else if flags&F_Fields_B != 0 {
			m.Add("fields", entry.Fields.AsArray())
		}
	}

	m.Add("msg", entry.Msg)

	data, _ := json.Marshal(m)
	entry.AddCompiled(flags, T_JSON, &data)
	entry.AddCompiled(F_NotSave, T_JSON, buf)
	if flags&F_NewLine != 0 {
		data = append(data, '\n')
	}
	_, err := w.Write(data)
	return err
}

// TextLogFunc uses buf to format entry with specified flags and writes it to w as a line
// adding buf to entry with TEXT output type
func TextLogFunc(buf *[]byte, entry *LogEntry, flags int, w io.Writer) error {
	if flags&(F_Time|F_Micro) != 0 {
		year, month, day := entry.Time.Date()
		appendInt(buf, day, 2)
		*buf = append(*buf, '/')
		appendInt(buf, int(month), 2)
		*buf = append(*buf, '/')
		appendInt(buf, year, 4)
		*buf = append(*buf, ' ')

		appendInt(buf, entry.Time.Hour(), 2)
		*buf = append(*buf, ':')
		appendInt(buf, entry.Time.Minute(), 2)
		*buf = append(*buf, ':')
		appendInt(buf, entry.Time.Second(), 2)
		if flags&F_Micro != 0 {
			*buf = append(*buf, '.')
			appendInt(buf, entry.Time.Nanosecond()/1000, 6)
		}
		*buf = append(*buf, ' ')
	}

	if entry.Prefixes != nil && len(entry.Prefixes) != 0 && flags&(F_Prefix|F_LastPrefix) != 0 {
		if flags&F_LastPrefix != 0 {
			*buf = append(*buf, '[')
			*buf = append(*buf, entry.Prefixes[len(entry.Prefixes)-1]...)
			*buf = append(*buf, "] "...)
		} else {
			for _, v := range entry.Prefixes {
				*buf = append(*buf, '[')
				*buf = append(*buf, v...)
				*buf = append(*buf, "] "...)
			}
		}
	}

	if flags&F_Level != 0 {
		*buf = append(*buf, '[')
		*buf = append(*buf, entry.Level.String()...)
		*buf = append(*buf, "] "...)
	}

	*buf = append(*buf, entry.Msg...)
	if flags&F_NewLine != 0 && entry.Msg[len(entry.Msg)-1] != '\n' {
		*buf = append(*buf, '\n')
	}

	entry.AddCompiled(flags, T_Text, buf)
	_, err := w.Write(*buf)
	return err
}

func appendInt(buf *[]byte, n int, w int) {
	var b [20]byte
	ind := len(b) - 1
	for n >= 10 || w > 1 {
		w--
		r := n / 10
		b[ind] = byte('0' + n - r*10)
		ind--
		n = r
	}
	b[ind] = byte('0' + n)
	*buf = append(*buf, b[ind:]...)
}

func NewJSONOutput(w io.Writer, flags int, close bool) Output {
	return &outputWrapper{
		w:          w,
		close:      close,
		flags:      flags,
		logLevel:   L_Info,
		outputType: T_JSON,
		LogFunc:    JSONLogFunc,
		CloseFunc:  DefaultCloseFunc,
	}
}

func NewTextOutput(w io.Writer, flags int, close bool) Output {
	return &outputWrapper{
		w:          w,
		close:      close,
		flags:      flags,
		logLevel:   L_Info,
		outputType: T_Text,
		LogFunc:    TextLogFunc,
		CloseFunc:  DefaultCloseFunc,
	}
}
