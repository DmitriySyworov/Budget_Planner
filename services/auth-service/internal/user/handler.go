package user

import (
	"net/http"
	"shared/response"
)

type HandlerUser struct {
	response.Response
	*response.HandlerResponse
	*ServiceUser
}

func NewHandlerUser(router *http.ServeMux, service *ServiceUser, handlerResponse *response.HandlerResponse) {
	user := HandlerUser{
		ServiceUser:     service,
		HandlerResponse: handlerResponse,
	}
	router.Handle("PATCH /api/v1/user", user)
	router.Handle("GET /api/v1/user", user)
	router.Handle("DELETE /api/v1/user", user)
}
