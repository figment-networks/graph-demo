package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log Logger
)

type Logger struct {
	*zap.Logger
	Level zap.AtomicLevel
}

func SetLevel(logLevel string) {
	Log.Level.SetLevel(getLevel(logLevel))
}

func GetLogger() *zap.Logger {
	return Log.Logger
}

func Init(encoding, logLevel string, logOutputs []string) error {

	lLev := zap.NewAtomicLevelAt(getLevel(logLevel))
	logConfig := zap.Config{
		OutputPaths: logOutputs,
		Encoding:    "json",
		Level:       lLev,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "level",
			TimeKey:        "time",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
	}

	log, err := logConfig.Build()
	if err != nil {
		return err
	}

	Log.Logger = log
	Log.Level = lLev

	return nil
}

func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Sync() {
	Log.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(err error, fields ...zap.Field) {
	Log.Error(err.Error(), fields...)
}

func getLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	case "error":
		return zap.ErrorLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	default:
		return zap.InfoLevel
	}
}
