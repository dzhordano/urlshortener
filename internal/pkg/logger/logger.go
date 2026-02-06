package logger

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	Sync() error
	Level() Level
}

type Level = int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Field = any
