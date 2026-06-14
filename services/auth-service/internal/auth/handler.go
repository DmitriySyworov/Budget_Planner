package auth

import (
	"net/http"
	"shared/handler_request"
	"shared/response"

	"github.com/go-playground/validator/v10"
)

type HandlerAuth struct {
	response.Response
	*response.HandlerResponse
	*ServiceAuth
}

func NewHandlerAuth(router *http.ServeMux, service *ServiceAuth, handlerResponse *response.HandlerResponse) {
	auth := &HandlerAuth{
		ServiceAuth: service,
	}
	router.HandleFunc("POST /api/v1/register", auth.Register())
	router.HandleFunc("POST /api/v1/login", auth.Login())
	router.HandleFunc("POST /api/v1/recovery", auth.Recovery())
	router.HandleFunc("POST /api/v1/confirm", auth.Confirm())
	router.HandleFunc("POST /api/v1/refresh", auth.Refresh())
}
func (h *HandlerAuth) Register() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, errBody := handler_request.HandlerRequest[RequestRegister](request.Body)
		if errBody != nil {
			if errValidate, ok := errBody.(validator.ValidationErrors); ok {

			} else {

			}
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}

	}
}
func (h *HandlerAuth) Login() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, errBody := handler_request.HandlerRequest[RequestLogin](request.Body)
		if errBody != nil {
			if errValidate, ok := errBody.(validator.ValidationErrors); ok {

			} else {

			}
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
	}
}
func (h *HandlerAuth) Recovery() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}
func (h *HandlerAuth) Confirm() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		request.Header.Get("User-Agent")
	}
}
func (h *HandlerAuth) Refresh() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, errBody := handler_request.HandlerRequest[RequestRefresh](request.Body)
		if errBody != nil {
			if errValidate, ok := errBody.(validator.ValidationErrors); ok {
				for _, err := range errValidate {
					if err.Field() == "RefreshJwt" {

					}
				}
			} else {

			}
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		userAgent := request.Header.Get("User-Agent")
		
	}
}
