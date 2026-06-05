package loggers

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug})),
	}
}
func (l *Logger) LoggerHandler(method, msg, userUUID string, status int, errors []string, data map[string]any) {
	if status >= 200 && status < 300 {
	}
}
