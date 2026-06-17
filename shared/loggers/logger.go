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

type DataLoggerHandler struct {
	Msg         string
	Status      int
	Method      string
	Path        string
	UserUUID    string
	Errors      []string
	DataRequest map[string]any
}

func (l *Logger) LoggerHandler(data *DataLoggerHandler) {
	if data.Status >= 200 && data.Status < 400 {
		l.Logger.Info(data.Msg, "status", data.Status, "method", data.Method, "path", data.Path, "user_uuid", data.UserUUID)
	} else if data.Status >= 400 && data.Status < 500 {
		l.Logger.Info(data.Msg, "status", data.Status, "method", data.Method, "path", data.Path, "user_uuid", data.UserUUID, "errors", data.Errors, "data", data.DataRequest)
	} else {
		l.Logger.Error(data.Msg, "status", data.Status, "method", data.Method, "path", data.Path, "user_uuid", data.UserUUID, "errors", data.Errors, "data", data.DataRequest)
	}
}
