package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envProd = "prod"
)

type Logger struct {
	*zap.SugaredLogger
}

func setupLogger(env string) *zap.SugaredLogger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "Logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    colorLevelEncoder, // custom encoder
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var logger *zap.Logger
	var core zapcore.Core

	switch env {
	case envProd:
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
	default:
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
	}

	logger = zap.New(core)
	return logger.Sugar()
}

// colorLevelEncoder sets colors for logging levels
func colorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		enc.AppendString("\033[35mDEBUG\033[0m") // purple
	case zapcore.InfoLevel:
		enc.AppendString("\033[34mINFO\033[0m") // blue
	case zapcore.WarnLevel:
		enc.AppendString("\033[33mWARN\033[0m") // yellow
	case zapcore.ErrorLevel:
		enc.AppendString("\033[31mERROR\033[0m") // red
	default:
		enc.AppendString(level.CapitalString())
	}
}

// Info Custom logging methods with colored output
func (l Logger) Info(args ...interface{}) {
	l.SugaredLogger.Infof("\033[1;32m%s\033[0m", fmt.Sprint(args...))
}

func (l Logger) Infof(format string, args ...interface{}) {
	l.SugaredLogger.Infof("\033[1;32m"+format+"\033[0m", args...)
}

func (l Logger) Debug(args ...interface{}) {
	l.SugaredLogger.Debugf("\033[1;32m%s\033[0m", fmt.Sprint(args...))
}

func (l Logger) Debugf(format string, args ...interface{}) {
	l.SugaredLogger.Debugf("\033[1;32m"+format+"\033[0m", args...)
}

func (l Logger) Warn(args ...interface{}) {
	l.SugaredLogger.Warnf("\033[1;32m%s\033[0m", fmt.Sprint(args...))
}

func (l Logger) Warnf(format string, args ...interface{}) {
	l.SugaredLogger.Warnf("\033[1;32m"+format+"\033[0m", args...)
}

func (l Logger) Error(args ...interface{}) {
	l.SugaredLogger.Errorf("\033[1;32m%s\033[0m", fmt.Sprint(args...))
}

func (l Logger) Errorf(format string, args ...interface{}) {
	l.SugaredLogger.Errorf("\033[1;32m"+format+"\033[0m", args...)
}

// NewLogger creates a new Logger instance
func NewLogger(env string) *Logger {
	return &Logger{
		SugaredLogger: setupLogger(env),
	}
}
