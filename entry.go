package log

import (
	"sync"
	"time"
)

var pool = sync.Pool{New: func() any { return new([]byte) }}

func putBuf(b *[]byte) {
	pool.Put(b)
}

func getBuf() *[]byte {
	b, _ := pool.Get().(*[]byte)
	*b = (*b)[:0]
	return b
}

// LogEntry represent a log call, and is passed into each Output registered into
// manager so no field must be edited.
//
// LogEntry.Prefix LogEntry.Fields and LogEntry.Compiled are guaranteed to be non-nil.
//
// Compiled should not be accessed directly, instead callers should use
// *LogEntry.AddCompiled() and *LogEntry.GetCompiled()
type LogEntry struct {
	Time     time.Time
	Prefixes []string
	Level    LogLevel
	Msg      string
	Fields   M

	Compiled []Compiled
	sync.Mutex
}

// GetBuf returns a buffer from a sync.Pool intend to lower allocations.
// It is recommended to use GetBuf when parsing data
//
// NOTE: see *LogEntry.AddCompiled to save used buffer into sync.Pool
func (entry *LogEntry) GetBuf() *[]byte {
	return getBuf()
}

// AddCompiled saves compiled buffer in order for potential next Output
// with same flags and output type to use instead of recompiling logEntry.
// If a buffer is already saved with same flags and output type, it will be Put into sync.Pool
//
// AddCompiled should always be called after a buffer has been compiled even
// if no other Outputs are present so to lower allocations.
func (entry *LogEntry) AddCompiled(flags int, outputType OutputType, Buf *[]byte) {
	if flags&F_NotSave != 0 {
		putBuf(Buf)
		return
	}
	entry.Lock()
	defer entry.Unlock()
	for _, v := range entry.Compiled {
		if v.Flag == flags && v.OutputType == outputType {
			putBuf(Buf)
			return
		}
	}
	entry.Compiled = append(entry.Compiled, Compiled{
		Flag:       flags,
		OutputType: outputType,
		Buf:        Buf,
	})
}

// GetCompiled checks if a buffer has already been saved into entry with
// provided flags and output type.
//
// buf is non-nil, if and only if, ok is true
func (entry *LogEntry) GetCompiled(flags int, outputType OutputType) (buf *[]byte, ok bool) {
	if flags&F_NotSave != 0 {
		return
	}
	entry.Lock()
	defer entry.Unlock()
	for _, v := range entry.Compiled {
		if v.Flag == flags && v.OutputType == outputType {
			ok = true
			buf = v.Buf
			return
		}
	}
	return
}

type Compiled struct {
	Flag       int
	OutputType OutputType
	Buf        *[]byte
}
