package auth

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"net/http"
	"shared/handler_request"
	"shared/loggers"
	"shared/response"
	"shared/shared_errors"
	"shared/shared_middleware"

	"github.com/go-playground/validator/v10"
)

type HandlerAuth struct {
	response.Response
	*response.HandlerResponse
	*loggers.Logger
	*ServiceAuth
}

func NewHandlerAuth(router *http.ServeMux, service *ServiceAuth, handlerResponse *response.HandlerResponse, logger *loggers.Logger, mv *shared_middleware.ManagerMiddleware) {
	auth := &HandlerAuth{
		ServiceAuth:     service,
		HandlerResponse: handlerResponse,
		Logger:          logger,
	}
	router.HandleFunc("POST /api/v1/register", auth.Register())
	router.HandleFunc("POST /api/v1/login", auth.Login())
	router.HandleFunc("POST /api/v1/recovery", auth.Recovery())
	router.Handle("POST /api/v1/confirm", mv.HandlerSessionToken(auth.Confirm()))
	router.HandleFunc("POST /api/v1/refresh", auth.Refresh())
}
func (h *HandlerAuth) Register() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestRegister](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Name" {
						h.Response.Error = append(h.Response.Error, ErrIncorrectName.Error())
					}
					if err.Field() == "Email" {
						h.Response.Error = append(h.Response.Error, custom_errors.ErrIncorrectEmail.Error())
					}
					if err.Field() == "Password" {
						h.Response.Error = append(h.Response.Error, custom_errors.ErrIncorrectEnterPassword.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		respAuth, errAuth := h.ServiceAuth.Register(body)
		h.Response.Error = append(h.Response.Error, errAuth...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
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
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestLogin](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Email" {
						h.Response.Error = append(h.Response.Error, custom_errors.ErrIncorrectEmail.Error())
					}
					if err.Field() == "Password" {
						h.Response.Error = append(h.Response.Error, custom_errors.ErrIncorrectEnterPassword.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		respAuth, errLogin := h.ServiceAuth.Login(body)
		h.Response.Error = append(h.Response.Error, errLogin...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if h.Response.Error[0] == custom_errors.ErrIncorrectPasswordOrEmail.Error() {
				h.ResponseSend(writer, http.StatusUnauthorized)
			} else if h.Response.Error[0] == custom_errors.ErrFailedSecurity.Error() {
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
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[common.RequestConfirm](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Code" {
						h.Response.Error = append(h.Response.Error, custom_errors.ErrIncorrectFormatCode.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		action := request.URL.Query().Get("action")
		values.DataLog.MapLog["action"] = action
		userAgent := request.Header.Get("User-Agent")
		values.DataLog.MapLog["user_agent"] = userAgent
		respConfirm, errConfirm := h.ServiceAuth.Confirm(body.Code, values.DataAuth.SessionID, action, userAgent)
		h.Response.Error = append(h.Response.Error, errConfirm...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			switch h.Response.Error[0] {
			case ErrUserAlreadyExist.Error(), ErrIncorrectAction.Error():
				h.ResponseSend(writer, http.StatusBadRequest)
			case custom_errors.ErrSessionExpired.Error(), custom_errors.ErrIncorrectCode.Error(), custom_errors.ErrIncorrectSessionID.Error():
				h.ResponseSend(writer, http.StatusUnauthorized)
			default:
				h.ResponseSend(writer, http.StatusInternalServerError)
			}
		}
		h.Response.Success = true
		h.Response.Data = respConfirm
		if action == actionRegister {
			h.ResponseSend(writer, http.StatusCreated)
		} else {
			h.ResponseSend(writer, http.StatusOK)
		}
	}
}
func (h *HandlerAuth) Refresh() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestRefresh](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "RefreshJwt" {
						h.Response.Error = append(h.Response.Error, ErrSentRefresh.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		userAgent := request.Header.Get("User-Agent")
		values.DataLog.MapLog["user_agent"] = userAgent
		respConfirm, errConfirm := h.ServiceAuth.Refresh(body.RefreshJwt, userAgent)
		h.Response.Error = append(h.Response.Error, errConfirm...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
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
