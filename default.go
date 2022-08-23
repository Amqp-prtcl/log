package log

import "os"

var DefaultLogger = NewLogger()

func init() {
	DefaultLogger.AddOutput(NewTextOutput(os.Stdout, F_Std, false))
}

func Sync() {
	DefaultLogger = DefaultLogger.Sync()
}

func Async() {
	DefaultLogger = DefaultLogger.Async()
}

func AddOutput(o Output) {
	DefaultLogger.AddOutput(o)
}

func RemoveOutput(o Output) {
	DefaultLogger.RemoveOutput(o)
}

func ForEachOutput(fn func(int, Output)) {
	DefaultLogger.ForEachOutput(fn)
}

func Debug(format string, a ...any) (int, error) {
	return DefaultLogger.Debug(format, a...)
}

func Info(format string, a ...any) (int, error) {
	return DefaultLogger.Info(format, a...)
}

func Warn(format string, a ...any) (int, error) {
	return DefaultLogger.Warn(format, a...)
}

func Error(format string, a ...any) (int, error) {
	return DefaultLogger.Error(format, a...)
}

func Fatal(format string, a ...any) (int, error) {
	return DefaultLogger.Fatal(format, a...)
}

func SetLogLevel(level LogLevel) {
	DefaultLogger.SetLogLevel(level)
}

func SetFlags(flags int) {
	DefaultLogger.SetFlags(flags)
}

func SetPrefix(prefix string) {
	DefaultLogger = DefaultLogger.SetPrefix(prefix)
}

func ResetPrefix() {
	DefaultLogger = DefaultLogger.ResetPrefix()
}

func AddPrefix(prefix ...string) {
	DefaultLogger = DefaultLogger.AddPrefix(prefix...)
}

func SetFields(m M) {
	DefaultLogger = DefaultLogger.SetFields(m)
}

func ResetFields() {
	DefaultLogger = DefaultLogger.ResetFields()
}

func AddFields(m M) {
	DefaultLogger = DefaultLogger.AddFields(m)
}

func Lock() {
	DefaultLogger.Lock()
}

func Unlock() {
	DefaultLogger.Unlock()
}

func Close() error {
	return DefaultLogger.Close()
}
