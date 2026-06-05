package middleware

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/response"
	"net/http"
	"strings"
)

type DataAuth struct {
	UserUUID  string
	SessionID string
}

func helperHandleHeader(header string) (string, error) {
	if !strings.Contains(header, "Bearer") {
		return "", custom_errors.ErrIncorrectToken
	}
	tokenSplit := strings.Split(header, " ")
	if len(tokenSplit) != 2 {
		return "", custom_errors.ErrIncorrectToken
	}
	if strings.Count(tokenSplit[1], ".") != 2 {
		return "", custom_errors.ErrIncorrectToken
	}
	return tokenSplit[1], nil
}
func (m *ManagerMiddleware) HandlerAuthToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		header := request.Header.Get("Authorization")
		token, errToken := helperHandleHeader(header)
		if errToken != nil {
			m.HandlerResponse.ResponseSend(writer, &response.Response{Error: []string{custom_errors.ErrIncorrectToken.Error()}}, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
