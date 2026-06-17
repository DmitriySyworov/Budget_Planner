package shared_middleware

import (
	"context"
	"net/http"
	"shared/loggers"
)

type DataLog struct {
	Method   string
	Path     string
	UserUUID string
	MapLog   map[string]any
	Errors   []string
}

func (m *ManagerMiddleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			m.ContextValues = &ContextValues{
				DataAuth: &DataAuth{
					UserUUID:  "",
					SessionID: "",
				},
				DataLog: &DataLog{
					UserUUID: "",
					Errors:   make([]string, 10),
					MapLog:   make(map[string]any, 8),
				},
			}
		}()
		wrapperWriter := &WrapperWriter{
			Status:         http.StatusOK,
			ResponseWriter: writer,
		}
		ctxValue := context.WithValue(context.Background(), KeyContextValue, m.ContextValues)
		ctxRequest := request.WithContext(ctxValue)
		next.ServeHTTP(wrapperWriter, ctxRequest)
		m.DataLog.Method = request.Method
		m.DataLog.Path = request.Pattern
		dataLoggerHandler := &loggers.DataLoggerHandler{
			Status:      wrapperWriter.Status,
			Method:      m.DataLog.Method,
			Path:        m.DataLog.Path,
			UserUUID:    m.ContextValues.DataLog.UserUUID,
			Errors:      m.ContextValues.DataLog.Errors,
			DataRequest: m.ContextValues.DataLog.MapLog,
		}
		if wrapperWriter.Status >= 200 && wrapperWriter.Status < 400 {
			dataLoggerHandler.Msg = "successful operation"
			m.Logger.LoggerHandler(dataLoggerHandler)
		} else {
			dataLoggerHandler.Msg = "unsuccessful operation"
			m.Logger.LoggerHandler(dataLoggerHandler)
		}
	})
}

type WrapperWriter struct {
	http.ResponseWriter
	Status int
}

func (w *WrapperWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}
