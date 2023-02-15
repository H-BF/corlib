package logger

import (
	"context"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	//LogLevel os alias tp zapcore.Level
	LogLevel = zapcore.Level

	//TypeOfLogger logger type
	TypeOfLogger struct {
		LevelEnabler
		*zap.SugaredLogger
	}

	//LevelEnabler is alias to zapcore.LevelEnabler
	LevelEnabler = zapcore.LevelEnabler
)

var (
	//global logger instance.
	global       TypeOfLogger
	defaultLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
)

func init() {
	SetLogger(New(defaultLevel))
}

// New creates new logger with standard EncoderConfig
// if lvl == nil, global AtomicLevel will be used
func New(level LevelEnabler, options ...zap.Option) TypeOfLogger {
	return NewWithSink(level, os.Stdout, options...)
}

// NewWithSink ...
func NewWithSink(level LevelEnabler, sink io.Writer, options ...zap.Option) TypeOfLogger {
	if level == nil {
		level = defaultLevel
	}
	return TypeOfLogger{
		LevelEnabler: level,
		SugaredLogger: zap.New(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zapcore.EncoderConfig{
					TimeKey:        "ts",
					LevelKey:       "lvl",
					NameKey:        "log-of",
					CallerKey:      "at",
					MessageKey:     "msg",
					StacktraceKey:  "stack",
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeLevel:    zapcore.LowercaseLevelEncoder,
					EncodeTime:     zapcore.ISO8601TimeEncoder,
					EncodeDuration: zapcore.SecondsDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}),
				zapcore.AddSync(sink),
				level,
			),
			options...,
		).Sugar(),
	}
}

// Named ...
func (lg TypeOfLogger) Named(n string) TypeOfLogger {
	ret := lg
	ret.SugaredLogger = lg.SugaredLogger.Named(n)
	return ret
}

// WithField adds additional field
func (lg TypeOfLogger) WithField(k, v interface{}) TypeOfLogger {
	ret := lg
	ret.SugaredLogger = lg.SugaredLogger.With(k, v)
	return ret
}

// Level returns current global logger level
func Level() LogLevel {
	return defaultLevel.Level()
}

// SetLevel sets level for global logger
func SetLevel(l LogLevel) {
	defaultLevel.SetLevel(l)
}

//Global returns current global logger.
func Global() TypeOfLogger {
	return global
}

// SetLogger sets global used logger. This function is not thread-safe.
func SetLogger(l TypeOfLogger) {
	global = l
}

// Below listed all logging functions
// Suffix meaning:
// * No suffix, e.g. Debug()   - log concatenated args
// * f,         e.g. Debugf()  - log using format string
// * KV,        e.g. DebugKV() - log key-values, odd args are keys, even – values

// Debug ...
func Debug(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Debug(args...)
}

// Debugf ...
func Debugf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Debugf(format, args...)
}

// DebugKV ...
func DebugKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Debugw(message, kvs...)
}

// Info ...
func Info(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Info(args...)
}

// Infof ...
func Infof(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Infof(format, args...)
}

// InfoKV ...
func InfoKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Infow(message, kvs...)
}

// Warn ...
func Warn(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Warn(args...)
}

// Warnf ...
func Warnf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Warnf(format, args...)
}

// WarnKV ...
func WarnKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Warnw(message, kvs...)
}

// Error ...
func Error(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Error(args...)
}

// Errorf ...
func Errorf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Errorf(format, args...)
}

// ErrorKV ...
func ErrorKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Errorw(message, kvs...)
}

// Fatal ...
func Fatal(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Fatal(args...)
}

// Fatalf ...
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Fatalf(format, args...)
}

// FatalKV ...
func FatalKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Fatalw(message, kvs...)
}

// Panic ...
func Panic(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Panic(args...)
}

// Panicf ...
func Panicf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Panicf(format, args...)
}

// PanicKV ...
func PanicKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Panicw(message, kvs...)
}
