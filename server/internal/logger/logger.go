// Package logger provides structured logging for the Alignment game server
package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

// InitLogger initializes the structured logger with JSON format
func InitLogger() {
	// Configure JSON handler for structured logging
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	
	Logger = slog.New(handler).With(
		"service", "alignment-server",
	)
	
	// Set as default logger
	slog.SetDefault(Logger)
}

// GetLogger returns the global logger instance
func GetLogger() *slog.Logger {
	if Logger == nil {
		InitLogger()
	}
	return Logger
}

// WithField adds a key-value pair to the logger context
func WithField(key string, value interface{}) *slog.Logger {
	return GetLogger().With(key, value)
}

// WithFields adds multiple key-value pairs to the logger context
func WithFields(fields map[string]interface{}) *slog.Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return GetLogger().With(args...)
}