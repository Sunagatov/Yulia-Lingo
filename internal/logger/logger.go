package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})

	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

func Info(msg string, fields ...logrus.Fields) {
	entry := log.WithFields(getFields(fields...))
	entry.Info(msg)
}

func Error(msg string, err error, fields ...logrus.Fields) {
	entry := log.WithFields(getFields(fields...))
	if err != nil {
		entry = entry.WithError(err)
	}
	entry.Error(msg)
}

func Debug(msg string, fields ...logrus.Fields) {
	entry := log.WithFields(getFields(fields...))
	entry.Debug(msg)
}

func Warn(msg string, fields ...logrus.Fields) {
	entry := log.WithFields(getFields(fields...))
	entry.Warn(msg)
}

func Fatal(msg string, err error, fields ...logrus.Fields) {
	entry := log.WithFields(getFields(fields...))
	if err != nil {
		entry = entry.WithError(err)
	}
	entry.Fatal(msg)
}

func getFields(fields ...logrus.Fields) logrus.Fields {
	if len(fields) == 0 {
		return logrus.Fields{}
	}
	if len(fields) == 1 {
		return fields[0]
	}

	result := make(logrus.Fields)
	for _, field := range fields {
		for k, v := range field {
			result[k] = v
		}
	}
	return result
}