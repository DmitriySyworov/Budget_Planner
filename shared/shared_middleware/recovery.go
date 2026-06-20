package shared_middleware

import (
	"net/http"
	"shared/loggers"
	"shared/response"
	"shared/shared_errors"
)

func (m *ManagerMiddleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if errPanic := recover(); errPanic != nil {
				m.Response.Error = append(m.Response.Error, shared_errors.ErrCriticalServer.Error())
				m.ResponseSend(writer, http.StatusInternalServerError)
				m.Logger.Error("critical error: ", errPanic)
				m.Logger.LoggerHandler(&loggers.DataLoggerHandler{
					Msg:         "critical error",
					Status:      500,
					Method:      m.DataLog.Method,
					Path:        m.DataLog.Path,
					UserUUID:    m.DataLog.UserUUID,
					Errors:      m.DataLog.Errors,
					DataRequest: m.DataLog.MapLog,
				})
			}
			m.HandlerResponse.Response = &response.Response{
				Error: make([]string, 0, 10),
			}
			m.ContextValues = &ContextValues{
				DataAuth: &DataAuth{
					UserUUID:  "",
					SessionID: "",
				},
				DataLog: &DataLog{
					UserUUID: "",
					Errors:   make([]string, 10),
					MapLog:   make(map[string]any, sizeMap),
				},
			}
		}()
		next.ServeHTTP(writer, request)
	})
}
