package logger

import (
	"context"
	"os"
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
}

type Field struct {
	Key   string
	Value interface{}
}

type logger struct {
	log *logrus.Logger
}

var defaultLogger Logger

func Initialize(cfg *config.Config) {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	if cfg.Logging.Format == "text" {
		log.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339, FullTimestamp: true})
	} else {
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
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

	defaultLogger = &logger{log: log}
}

func New() Logger {
	if defaultLogger == nil {
		panic("logger not initialized")
	}
	return defaultLogger
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.entry(ctx, fields).Debug(msg)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.entry(ctx, fields).Info(msg)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.entry(ctx, fields).Warn(msg)
}

func (l *logger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	l.entry(ctx, fields).WithError(err).Error(msg)
}

func (l *logger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	l.entry(ctx, fields).WithError(err).Fatal(msg)
}

func (l *logger) entry(ctx context.Context, fields []Field) *logrus.Entry {
	entry := l.log.WithContext(ctx)
	for _, f := range fields {
		entry = entry.WithField(f.Key, f.Value)
	}
	return entry
}
