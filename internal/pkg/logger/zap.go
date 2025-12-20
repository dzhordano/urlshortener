package logger

// For convenience...?

type Logger interface {
	Debug(v ...any)
	Debugw(msg string, v ...any)
	Info(v ...any)
	Infow(msg string, v ...any)
	Warn(v ...any)
	Warnw(msg string, v ...any)
	Error(v ...any)
	Errorw(msg string, v ...any)
	Panic(v ...any)
	Panicw(msg string, v ...any)
}
