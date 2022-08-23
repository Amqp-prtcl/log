package log

import (
	"context"
	"os"
)

type ctxKey int
type Action int

const (
	loggerKey ctxKey = iota
)

// NewContext wraps logger into parent context
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext unwraps logger from context or return false if not present
func FromContext(ctx context.Context) (Logger, bool) {
	log, ok := ctx.Value(loggerKey).(Logger)
	return log, ok
}

func SyncCtx(ctx context.Context) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.Sync())
}

func AsyncCtx(ctx context.Context) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.Async())
}

func LogFromContext(ctx context.Context, level LogLevel, format string, a ...any) (int, error) {
	log, ok := FromContext(ctx)
	if !ok {
		return 0, nil
	}
	return log.Log(level, format, a...)
}

func DebugCtx(ctx context.Context, format string, a ...any) (int, error) {
	return LogFromContext(ctx, L_Debug, format, a...)
}

func InfoCtx(ctx context.Context, format string, a ...any) (int, error) {
	return LogFromContext(ctx, L_Info, format, a...)
}

func WarnCtx(ctx context.Context, format string, a ...any) (int, error) {
	return LogFromContext(ctx, L_Warn, format, a...)
}

func ErrorCtx(ctx context.Context, format string, a ...any) (int, error) {
	return LogFromContext(ctx, L_Error, format, a...)
}

func FatalCtx(ctx context.Context, format string, a ...any) (int, error) {
	log, ok := FromContext(ctx)
	if !ok {
		return 0, nil
	}
	log.Sync().Log(L_Fatal, format, a...)
	os.Exit(1)
	return 0, nil
}

func SetFlagsCtx(ctx context.Context, flags int) {
	log, ok := FromContext(ctx)
	if !ok {
		return
	}
	log.SetFlags(flags)
}

func SetLogLevelCtx(ctx context.Context, level LogLevel) {
	log, ok := FromContext(ctx)
	if !ok {
		return
	}
	log.SetLogLevel(level)
}

func AddPrefixCtx(ctx context.Context, prefix ...string) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.AddPrefix(prefix...))
}

func SetPrefixCtx(ctx context.Context, prefix string) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.SetPrefix(prefix))
}

func ResetPrefixCtx(ctx context.Context) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.ResetPrefix())
}

func SetFieldsCtx(ctx context.Context, m M) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.SetFields(m))
}

func ResetFieldsCtx(ctx context.Context) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.ResetFields())
}

func AddFieldsCtx(ctx context.Context, m M) context.Context {
	log, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	return NewContext(ctx, log.AddFields(m))
}

func AddOutputCtx(ctx context.Context, o Output) {
	log, ok := FromContext(ctx)
	if !ok {
		return
	}
	log.AddOutput(o)
}

func RemoveOutputCtx(ctx context.Context, o Output) {
	log, ok := FromContext(ctx)
	if !ok {
		return
	}
	log.RemoveOutput(o)
}

func ForEachOutputCtx(ctx context.Context, fn func(int, Output)) {
	log, ok := FromContext(ctx)
	if !ok {
		return
	}
	log.ForEachOutput(fn)
}

func LockCtx(ctx context.Context) {
	log, ok := FromContext(ctx)
	if !ok {
		return
	}
	log.Lock()
}

func UnlockCtx(ctx context.Context) {
	log, ok := FromContext(ctx)
	if !ok {
		return
	}
	log.Unlock()
}

func CloseCtx(ctx context.Context) error {
	log, ok := FromContext(ctx)
	if !ok {
		return nil
	}
	return log.Close()
}
