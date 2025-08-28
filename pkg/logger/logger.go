package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Logger представляет собой обертку над slog для логирования
type Logger struct {
	logger *slog.Logger
}

// NewLogger создает новый экземпляр логгера с указанным уровнем логирования
func NewLogger(level string) *Logger {
	// Определение уровня логирования
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Создание обработчика логов с форматированием JSON
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Создание логгера
	logger := slog.New(handler)

	return &Logger{logger: logger}
}

// Debug логирует сообщение с уровнем Debug
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info логирует сообщение с уровнем Info
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn логирует сообщение с уровнем Warn
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error логирует сообщение с уровнем Error
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Fatal логирует сообщение с уровнем Error и завершает программу
func (l *Logger) Fatal(msg string, args ...any) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}

// With возвращает новый логгер с добавленными атрибутами
func (l *Logger) With(args ...any) *Logger {
	return &Logger{logger: l.logger.With(args...)}
}