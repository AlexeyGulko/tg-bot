package logger

import (
	"log"

	"go.uber.org/zap"
)

var logger *zap.Logger

type Config interface {
	DevMode() string
}

func Init(cfg Config) {
	var localLogger *zap.Logger
	var err error
	switch cfg.DevMode() {
	case "production":
		cfg := zap.NewProductionConfig()
		cfg.DisableCaller = true
		cfg.DisableStacktrace = false
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		localLogger, err = cfg.Build()
	default:
		localLogger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatal("logger init", err)
	}
	logger = localLogger
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
