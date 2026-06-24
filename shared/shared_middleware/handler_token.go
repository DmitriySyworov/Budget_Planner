package shared_middleware

import (
	"context"
	"net/http"
	"shared/response"
	"shared/shared_errors"
	"shared/shared_jwt"
	"strings"
)

type DataAuth struct {
	UserUUID  string
	SessionID string
}

func HelperHandleHeader(header string) (string, error) {
	if !strings.Contains(header, "Bearer") {
		return "", shared_errors.ErrInvalidAccessToken
	}
	tokenSplit := strings.Split(header, " ")
	if len(tokenSplit) != 2 {
		return "", shared_errors.ErrInvalidAccessToken
	}
	if strings.Count(tokenSplit[1], ".") != 2 {
		return "", shared_errors.ErrInvalidAccessToken
	}
	return tokenSplit[1], nil
}
func (m *ManagerSharedMiddleware) HandlerAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(KeyContextValue)
		values, ok := ctxValues.(*ContextValues)
		if !ok {
			m.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + "middleware HandlerAuthToken")
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			m.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		header := request.Header.Get("Authorization")
		token, errToken := HelperHandleHeader(header)
		if errToken != nil {
			values.DataLog.Errors = shared_errors.ErrInvalidAccessToken.Error()
			resp.Error["auth"] = shared_errors.ErrInvalidAccessToken.Error()
			m.HandlerResponse.ResponseSend(writer, resp, http.StatusUnauthorized)
			return
		}
		j := shared_jwt.NewSharedJWT(m.Signature)
		userUUID, errParse := j.ParseAccessToken(token)
		if errParse != nil {
			values.DataLog.Errors = shared_errors.ErrInvalidAccessToken.Error()
			resp.Error["auth"] = shared_errors.ErrInvalidAccessToken.Error()
			m.HandlerResponse.ResponseSend(writer, resp, http.StatusUnauthorized)
			return
		}
		if len(userUUID) != 36 {
			values.DataLog.Errors = shared_errors.ErrInvalidAccessToken.Error()
			resp.Error["auth"] = shared_errors.ErrInvalidAccessToken.Error()
			m.HandlerResponse.ResponseSend(writer, resp, http.StatusUnauthorized)
			return
		}
		values.DataLog.UserUUID = userUUID
		values.DataAuth.UserUUID = userUUID
		newCtxValue := context.WithValue(context.Background(), KeyContextValue, values)
		ctxRequest := request.WithContext(newCtxValue)
		next.ServeHTTP(writer, ctxRequest)
	})
}
