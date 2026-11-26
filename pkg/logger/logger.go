package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level  string
	Folder string
}

type Logger struct {
	*zap.Logger
	file *os.File
}

type ctxKey struct{ name string }

var (
	loggerCtxKey    = ctxKey{"logger"}
	requestIDCtxKey = ctxKey{"request_id"}
)

func NewLogger(config Config) (*Logger, error) {
	zapLvl := zap.NewAtomicLevel()
	if err := zapLvl.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("unmarshal log level: %w", err)
	}

	if err := os.MkdirAll(config.Folder, 0o750); err != nil {
		return nil, fmt.Errorf("mkdir log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	logFilePath := filepath.Join(config.Folder, fmt.Sprintf("%s.log", timestamp))

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	defer func() {
		if err != nil {
			_ = logFile.Close()
		}
	}()

	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00")
	zapConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	fileEncoder := zapcore.NewJSONEncoder(zapConfig)

	consoleCfg := zap.NewDevelopmentEncoderConfig()
	consoleCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000Z07:00")
	consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLvl),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zapLvl),
	)

	zapLogger := zap.New(core, zap.AddCaller())

	return &Logger{Logger: zapLogger, file: logFile}, nil
}

func FromContext(ctx context.Context) *Logger {
	if log, ok := ctx.Value(loggerCtxKey).(*Logger); ok {
		return log
	}
	return &Logger{Logger: zap.NewNop()}
}

func WithContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, log)
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDCtxKey).(string)
	return id
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDCtxKey, id)
}

func (l *Logger) With(field ...zap.Field) *Logger {
	return &Logger{Logger: l.Logger.With(field...), file: l.file}
}

func (l *Logger) Close() {
	if l.file == nil {
		return
	}
	if err := l.file.Close(); err != nil {
		fmt.Println("failed to close application logger:", err)
	}
}
