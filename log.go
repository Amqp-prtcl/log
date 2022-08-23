package log

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// TODO: doc
type Logger struct {
	m *manager

	block  bool
	prefix []string
	fields M
}

// NewLogger creates an Async Logger with a new underlying Output Manager
// so callers need to add Outputs in order to make it do something useful.
func NewLogger() Logger {
	return Logger{
		m:      newManager(),
		block:  false,
		prefix: []string{},
		fields: M{},
	}
}

// makes all log calls blocking (meaning each log call will wait for
// the call to be parsed and printed by all outputs before returning)
//
// first log call after Sync might be slow as it will wait for any older
// call to end before starting
func (l Logger) Sync() Logger {
	return Logger{
		m:      l.m,
		block:  true,
		prefix: duplicate(l.prefix),
		fields: duplicate(l.fields),
	}
}

// makes log calls non-blocking (meaning that parsing and
// printing is done in a separate goroutine so each log call returns immediately)
func (l Logger) Async() Logger {
	return Logger{
		m:      l.m,
		block:  false,
		prefix: duplicate(l.prefix),
		fields: duplicate(l.fields),
	}
}

// error will always be nil and int equal to the length of formatted message
func (l Logger) Log(level LogLevel, format string, a ...any) (int, error) {
	var entry = &LogEntry{
		Time:     time.Now(),
		Prefixes: l.prefix,
		Level:    level,
		Msg:      fmt.Sprintf(format, a...),
		Fields:   l.fields,
		Compiled: []Compiled{},
		Mutex:    sync.Mutex{},
	}

	err := l.m.log(entry, l.block)

	return len(entry.Msg), err
}

// error will always be nil and int equal to the length of formatted message
func (l Logger) Debug(format string, a ...any) (int, error) {
	return l.Log(L_Debug, format, a...)
}

// error will always be nil and int equal to the length of formatted message
func (l Logger) Info(format string, a ...any) (int, error) {
	return l.Log(L_Info, format, a...)
}

// error will always be nil and int equal to the length of formatted message
func (l Logger) Warn(format string, a ...any) (int, error) {
	return l.Log(L_Warn, format, a...)
}

// error will always be nil and int equal to the length of formatted message
func (l Logger) Error(format string, a ...any) (int, error) {
	return l.Log(L_Error, format, a...)
}

// Fatal waits for log call to end and then exits program by 'os.Exit(1)'
//
// All Fatal calls are Sync meaning that if there are other log calls that have not
// finished yet it will wait for all those before actually exiting
//
// NOTE: Fatal won't close Logger (any file or socket used as Output in manager
// won't be automatically closed by logger)
func (l Logger) Fatal(format string, a ...any) (int, error) {
	l.Sync().Log(L_Fatal, format, a...)
	os.Exit(1)
	return 0, nil
}

func (l Logger) SetPrefix(prefix string) Logger {
	return Logger{
		m:      l.m,
		block:  l.block,
		prefix: []string{prefix},
		fields: duplicate(l.fields),
	}
}

func (l Logger) ResetPrefix() Logger {
	return Logger{
		m:      l.m,
		block:  l.block,
		prefix: []string{},
		fields: duplicate(l.fields),
	}
}

func (l Logger) AddPrefix(prefixes ...string) Logger {
	return Logger{
		m:      l.m,
		block:  l.block,
		prefix: duplicate(l.prefix, prefixes),
		fields: duplicate(l.fields),
	}
}

func (l Logger) SetFields(fields M) Logger {
	return Logger{
		m:      l.m,
		block:  l.block,
		prefix: duplicate(l.prefix),
		fields: duplicate(fields),
	}
}

func (l Logger) ResetFields() Logger {
	return Logger{
		m:      l.m,
		block:  l.block,
		prefix: duplicate(l.prefix),
		fields: M{},
	}
}

func (l Logger) AddFields(fields M) Logger {
	return Logger{
		m:      l.m,
		block:  l.block,
		prefix: duplicate(l.prefix),
		fields: duplicate(l.fields, fields),
	}
}

// add an outputs
//
// NOTE: removing or adding an output will affect all
// loggers sharing same underlying Output manager
func (l Logger) AddOutput(o Output) {
	l.m.addOutput(o)
}

// remove an outputs
//
// NOTE: removing or adding an output will affect all
// loggers sharing same underlying Output manager
func (l Logger) RemoveOutput(o Output) {
	l.m.removeOutput(o)
}

// runs fn for each output registered to any logger sharing its
// underlying Output manager
func (l Logger) ForEachOutput(fn func(int, Output)) {
	l.m.forEach(fn)
}

// Overwrites log level for all outputs registered to any logger sharing its
// underlying Output manager
func (l Logger) SetLogLevel(level LogLevel) {
	l.ForEachOutput(func(i int, o Output) {
		o.SetLogLevel(level)
	})
}

// Overwrites flags for all outputs registered
func (l Logger) SetFlags(flags int) {
	l.ForEachOutput(func(i int, o Output) {
		o.SetFlags(flags)
	})
}

// Lock waits the end of current log call before preventing
// execution of any subsequent log calls until next Unlock call
//
// It is intended to defer potentially heavy parsing and io
// during computing intensive operations
//
// NOTE: Lock will prevent any log call that has not already
// started from returning thus blocking all
// Sync Loggers sharing same underlying Output manager will wait for
// Unlock before executing. As for related Async Loggers, they will also
// block if manager's buffer/queue is full
func (l Logger) Lock() {
	l.m.Lock()
}

// NOTE: Unlock call to Unlocked Logger results in a panic
func (l Logger) Unlock() {
	l.m.Unlock()
}

// closes logger's underlying Output manager.
//
// Close flushes, waiting for all log calls to be treated
// before closing all submitted outputs
//
// any log call to any logger sharing same Output manager will return a ErrManagerClose
// as well for other Close calls.
//
// If logger is Locked, it waits for it to be Unlocked.
//
// NOTE: Since Close closes logger's underlying Output manager,
// it will close all loggers sharing this manager (basically any
// logger derived from the logger returned by NewLogger())
func (l Logger) Close() error {
	return l.m.Close()
}

/*
Sync()
Async()

AddOutput()
RemoveOutput()
ForEachOutput()

Debug
Info
Warn
Error
Fatal

SetFlags
SetLogLevel

AddPrefix
SetPrefixes
ResetPrefixes

SetFields
ResetFields
AddFields

Lock
Unlock

Close
*/
