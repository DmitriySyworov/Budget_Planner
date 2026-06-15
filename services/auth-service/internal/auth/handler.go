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
		HandlerResponse : handlerResponse,
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
				for _, err := range errValidate {
					if err.Field() == "Name" {
						h.Response.Error = append(h.Response.Error, ErrIncorrectName.Error())
					}
					if err.Field() == "Email" {
						h.Response.Error = append(h.Response.Error, ErrIncorrectEmail.Error())
					}
					if err.Field() == "Password" {
						h.Response.Error = append(h.Response.Error, ErrIncorrectEnterPassword.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		respAuth, errAuth := h.ServiceAuth.Register(body)
		h.Response.Error = append(h.Response.Error, errAuth...)
		if len(h.Response.Error) != 0 {
			if h.Response.Error[0] == ErrUserAlreadyExist.Error() {
				h.ResponseSend(writer, http.StatusBadRequest)
			} else {
				h.ResponseSend(writer, http.StatusInternalServerError)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = respAuth
		h.ResponseSend(writer, http.StatusAccepted)
	}
}
func (h *HandlerAuth) Login() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, errBody := handler_request.HandlerRequest[RequestLogin](request.Body)
		if errBody != nil {
			if errValidate, ok := errBody.(validator.ValidationErrors); ok {
				for _, err := range errValidate {
					if err.Field() == "Email" {
						h.Response.Error = append(h.Response.Error, ErrIncorrectEmail.Error())
					}
					if err.Field() == "Password" {
						h.Response.Error = append(h.Response.Error, ErrIncorrectEnterPassword.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		respAuth, errLogin := h.ServiceAuth.Login(body)
		h.Response.Error = append(h.Response.Error, errLogin...)
		if len(h.Response.Error) != 0 {
			if h.Response.Error[0] == ErrIncorrectPasswordOrEmail.Error() {
				h.ResponseSend(writer, http.StatusUnauthorized)
			} else if h.Response.Error[0] == ErrFailedSecurity.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = respAuth
		h.ResponseSend(writer, http.StatusAccepted)
	}
}
func (h *HandlerAuth) Recovery() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}
func (h *HandlerAuth) Confirm() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, errBody := handler_request.HandlerRequest[RequestConfirm](request.Body)
		if errBody != nil {
			if errValidate, ok := errBody.(validator.ValidationErrors); ok {
				for _, err := range errValidate {
					if err.Field() == "Code" {
						h.Response.Error = append(h.Response.Error, ErrIncorrectFormatCode.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		action := request.URL.Query().Get("action")
		userAgent := request.Header.Get("User-Agent")
		respConfirm, errConfirm := h.ServiceAuth.Confirm(body.Code, , action, userAgent)
		h.Response.Error = append(h.Response.Error, errConfirm...)
		if len(h.Response.Error) != 0 {
			switch h.Response.Error[0] {
			case ErrUserAlreadyExist.Error(), ErrIncorrectAction.Error():
				h.ResponseSend(writer, http.StatusBadRequest)
			case ErrSessionExpired.Error(), ErrIncorrectCode.Error(), ErrIncorrectSessionID.Error():
				h.ResponseSend(writer, http.StatusUnauthorized)
			default:
				h.ResponseSend(writer, http.StatusInternalServerError)
			}
		}
		h.Response.Success = true
		h.Response.Data = respConfirm
		if action == actionRegister{
			h.ResponseSend(writer, http.StatusCreated)
		} else {
			h.ResponseSend(writer, http.StatusOK)
		}
	}
}
func (h *HandlerAuth) Refresh() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, errBody := handler_request.HandlerRequest[RequestRefresh](request.Body)
		if errBody != nil {
			if errValidate, ok := errBody.(validator.ValidationErrors); ok {
				for _, err := range errValidate {
					if err.Field() == "RefreshJwt" {
					h.Response.Error = append(h.Response.Error, ErrSentRefresh.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		userAgent := request.Header.Get("User-Agent")
		respConfirm, errConfirm := h.ServiceAuth.Refresh(body.RefreshJwt, userAgent)
				h.Response.Error = append(h.Response.Error, errConfirm...)
				if len(h.Response.Error) != 0 {
					if h.Response.Error[0] == ErrRenewalRefresh.Error() {
						h.ResponseSend(writer, http.StatusUnauthorized)
					} else {
						h.ResponseSend(writer, http.StatusInternalServerError)
					}
					return
				}
				h.Response.Success = true
				h.Response.Data = respConfirm
				h.ResponseSend(writer, http.StatusOK)
	}
}
