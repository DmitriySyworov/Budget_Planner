package middleware

import (
	"app/auth-service/internal/JWT"
	"context"
	"net/http"
	"shared/response"
	"shared/shared_errors"
	"shared/shared_middleware"
)

func (m *ManagerMiddleware) HandlerSessionToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make([]string, 0, 1),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			m.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + "middleware HandlerSessionToken")
			resp.Error = append(resp.Error, shared_errors.ErrCriticalServer.Error())
			m.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		header := request.Header.Get("X-Session-Token")
		token, errToken := shared_middleware.HelperHandleHeader(header)
		if errToken != nil {
			resp.Error = append(resp.Error, shared_errors.ErrInvalidSessionToken.Error())
			m.HandlerResponse.ResponseSend(writer, resp, http.StatusUnauthorized)
			return
		}
		j := JWT.NewJWT(m.Signature, m.Logger)
		sessionID, errParse := j.ParseSessionToken(token)
		if errParse != nil {
			resp.Error = append(resp.Error, shared_errors.ErrInvalidSessionToken.Error())
			m.HandlerResponse.ResponseSend(writer, resp, http.StatusUnauthorized)
			return
		}
		if len(sessionID) != 36 {
			resp.Error = append(resp.Error, shared_errors.ErrInvalidSessionToken.Error())
			m.HandlerResponse.ResponseSend(writer, resp, http.StatusUnauthorized)
			return
		}
		values.DataLog.MapLog["session_id"] = sessionID
		values.DataAuth.SessionID = sessionID
		newCtxValue := context.WithValue(context.Background(), shared_middleware.KeyContextValue, values)
		ctxRequest := request.WithContext(newCtxValue)
		next.ServeHTTP(writer, ctxRequest)
	})
}
