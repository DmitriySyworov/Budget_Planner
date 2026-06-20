package shared_middleware

import (
	"net/http"
	"shared/response"
	"shared/shared_errors"
)

func (m *ManagerSharedMiddleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if errPanic := recover(); errPanic != nil {
				resp := &response.Response{
					Error: make([]string, 0, 1),
				}
				resp.Error = append(resp.Error, shared_errors.ErrCriticalServer.Error())
				m.ResponseSend(writer, resp, http.StatusInternalServerError)
				m.Logger.Error("critical error: ", "panic", errPanic)
			}
		}()
		next.ServeHTTP(writer, request)
	})
}
