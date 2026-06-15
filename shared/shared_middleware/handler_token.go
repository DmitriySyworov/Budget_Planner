package shared_middleware

import (
	"net/http"
	"shared/shared_errors"
	"shared/shared_jwt"
	"strings"
)

type DataAuth struct {
	UserUUID  string
	SessionID string
}

func helperHandleHeader(header string) (string, error) {
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
func (m *ManagerMiddleware) HandlerAuthToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		header := request.Header.Get("Authorization")
		token, errToken := helperHandleHeader(header)
		if errToken != nil {
			m.Response.Error = append(m.Response.Error, shared_errors.ErrInvalidAccessToken.Error())
			m.HandlerResponse.ResponseSend(writer, http.StatusUnauthorized)
			return
		}
		j := shared_jwt.NewSharedJWT(m.Signature)
		userUUID, errParse := j.ParseAccessToken(token)
		if errParse != nil {
			m.Response.Error = append(m.Response.Error, shared_errors.ErrInvalidAccessToken.Error())
			m.HandlerResponse.ResponseSend(writer, http.StatusUnauthorized)
			return
		}
		if len(userUUID) != 36 {
			m.Response.Error = append(m.Response.Error, shared_errors.ErrInvalidAccessToken.Error())
			m.HandlerResponse.ResponseSend(writer, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
