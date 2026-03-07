package logger

import (
	"context"
	"log"
	"log/slog"
	"os"

	"Yulia-Lingo/internal/config"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, err error, fields ...Field)
}

type Field struct {
	Key   string
	Value any
}

type logger struct{ log *slog.Logger }

func New(cfg *config.Config) Logger {
	level := slog.LevelInfo
	switch cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.Logging.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	sl := slog.New(handler)
	// redirect stdlib log (used by tgbotapi) into structured logger
	log.SetOutput(&stdlibWriter{sl})
	log.SetFlags(0)
	return &logger{log: sl}
}

// stdlibWriter forwards stdlib log lines to slog at WARN level.
type stdlibWriter struct{ sl *slog.Logger }

func (w *stdlibWriter) Write(p []byte) (int, error) {
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.sl.Warn(msg)
	return len(p), nil
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log.DebugContext(ctx, msg, toAttrs(fields)...)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log.InfoContext(ctx, msg, toAttrs(fields)...)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log.WarnContext(ctx, msg, toAttrs(fields)...)
}

func (l *logger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	attrs := append(toAttrs(fields), slog.Any("error", err))
	l.log.ErrorContext(ctx, msg, attrs...)
}

func toAttrs(fields []Field) []any {
	attrs := make([]any, 0, len(fields)*2)
	for _, f := range fields {
		attrs = append(attrs, f.Key, f.Value)
	}
	return attrs
}
