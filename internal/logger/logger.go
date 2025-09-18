package logger

import (
	"context"
	"os"
	"runtime"
	"time"

	"Yulia-Lingo/internal/config"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, err error, fields ...Field)
	Fatal(ctx context.Context, msg string, err error, fields ...Field)
	WithFields(fields ...Field) Logger
}

type Field struct {
	Key   string
	Value interface{}
}

type logger struct {
	log    *logrus.Logger
	fields logrus.Fields
}

var defaultLogger Logger

func Initialize(cfg *config.Config) {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	switch cfg.Logging.Format {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	switch cfg.Logging.Level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	log.AddHook(&contextHook{})
	defaultLogger = &logger{log: log, fields: make(logrus.Fields)}
}

func New(fields ...Field) Logger {
	if defaultLogger == nil {
		panic("logger not initialized")
	}
	return defaultLogger.WithFields(fields...)
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.logWithContext(ctx, logrus.DebugLevel, msg, nil, fields...)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.logWithContext(ctx, logrus.InfoLevel, msg, nil, fields...)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.logWithContext(ctx, logrus.WarnLevel, msg, nil, fields...)
}

func (l *logger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	l.logWithContext(ctx, logrus.ErrorLevel, msg, err, fields...)
}

func (l *logger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	l.logWithContext(ctx, logrus.FatalLevel, msg, err, fields...)
}

func (l *logger) WithFields(fields ...Field) Logger {
	newFields := make(logrus.Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for _, field := range fields {
		newFields[field.Key] = field.Value
	}
	return &logger{log: l.log, fields: newFields}
}

func (l *logger) logWithContext(ctx context.Context, level logrus.Level, msg string, err error, fields ...Field) {
	entry := l.log.WithFields(l.fields)

	for _, field := range fields {
		entry = entry.WithField(field.Key, field.Value)
	}

	if err != nil {
		entry = entry.WithError(err)
	}

	if ctx != nil {
		entry = entry.WithContext(ctx)
	}

	entry.Log(level, msg)
}

type contextHook struct{}

func (h *contextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *contextHook) Fire(entry *logrus.Entry) error {
	if pc, file, line, ok := runtime.Caller(8); ok {
		entry.Data["caller"] = map[string]interface{}{
			"function": runtime.FuncForPC(pc).Name(),
			"file":     file,
			"line":     line,
		}
	}
	return nil
}

// Backward compatibility functions
func Info(msg string, fields ...logrus.Fields) {
	if defaultLogger != nil {
		defaultLogger.Info(context.Background(), msg, convertFields(fields...)...)
	}
}

func Error(msg string, err error, fields ...logrus.Fields) {
	if defaultLogger != nil {
		defaultLogger.Error(context.Background(), msg, err, convertFields(fields...)...)
	}
}

func Debug(msg string, fields ...logrus.Fields) {
	if defaultLogger != nil {
		defaultLogger.Debug(context.Background(), msg, convertFields(fields...)...)
	}
}

func Warn(msg string, fields ...logrus.Fields) {
	if defaultLogger != nil {
		defaultLogger.Warn(context.Background(), msg, convertFields(fields...)...)
	}
}

func Fatal(msg string, err error, fields ...logrus.Fields) {
	if defaultLogger != nil {
		defaultLogger.Fatal(context.Background(), msg, err, convertFields(fields...)...)
	}
}

func convertFields(fields ...logrus.Fields) []Field {
	var result []Field
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			result = append(result, Field{Key: k, Value: v})
		}
	}
	return result
}