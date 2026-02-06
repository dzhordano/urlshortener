package logger

import (
	"context"
	"log/slog"
	"os"
)

var _ Logger = (*SlogLogger)(nil)

type SlogLogger struct {
	log *slog.Logger
	lvl Level
}

func NewSlogLogger(dev bool, levelStr string) (*SlogLogger, error) {
	lvl := strToSlogLevel(levelStr)

	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	var h slog.Handler
	if dev {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &SlogLogger{log: slog.New(h)}, nil
}

func (s *SlogLogger) Log(ctx context.Context, level Level, msg string, fields ...Field) {
	s.log.Log(ctx, toSlogLevel(level), msg, fields...)
}

func (s *SlogLogger) Debug(msg string, fields ...Field) {
	s.log.Debug(msg, fields...)
}

func (s *SlogLogger) Info(msg string, fields ...Field) {
	s.log.Info(msg, fields...)
}

func (s *SlogLogger) Warn(msg string, fields ...Field) {
	s.log.Warn(msg, fields...)
}

func (s *SlogLogger) Error(msg string, fields ...Field) {
	s.log.Error(msg, fields...)
}

func (s *SlogLogger) With(fields ...Field) Logger {
	return &SlogLogger{log: s.log.With(fields...)}
}

func (s *SlogLogger) Sync() error {
	return nil
}

func (s *SlogLogger) Level() Level {
	return s.lvl
}

func toSlogLevel(levelStr Level) slog.Level {
	switch levelStr {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func strToSlogLevel(str string) slog.Level {
	switch str {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
