package auth

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/middleware"
	"net/http"
	"shared/handler_request"
	"shared/loggers"
	"shared/response"
	"shared/shared_errors"
	"shared/shared_middleware"

	"github.com/go-playground/validator/v10"
)

type HandlerAuth struct {
	*response.HandlerResponse
	*loggers.Logger
	*ServiceAuth
}

func NewHandlerAuth(router *http.ServeMux, service *ServiceAuth, handlerResponse *response.HandlerResponse, logger *loggers.Logger, mv *middleware.ManagerMiddleware) {
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
		resp := &response.Response{
			Error: make([]string, 0, 5),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error = append(resp.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestRegister](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Name" {
						values.DataLog.MapLog["name"] = body.Name
						resp.Error = append(resp.Error, ErrIncorrectName.Error())
					}
					if err.Field() == "Email" {
						values.DataLog.MapLog["email"] = body.Email
						resp.Error = append(resp.Error, custom_errors.ErrIncorrectEmail.Error())
					}
					if err.Field() == "Password" {
						resp.Error = append(resp.Error, custom_errors.ErrIncorrectEnterPassword.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		respAuth, errAuth := h.ServiceAuth.Register(body)
		resp.Error = append(resp.Error, errAuth...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == ErrUserAlreadyExist.Error() {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			} else {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respAuth
		h.ResponseSend(writer, resp, http.StatusAccepted)
	}
}
func (h *HandlerAuth) Login() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make([]string, 0, 5),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error = append(resp.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestLogin](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Email" {
						values.DataLog.MapLog["email"] = body.Email
						resp.Error = append(resp.Error, custom_errors.ErrIncorrectEmail.Error())
					}
					if err.Field() == "Password" {
						resp.Error = append(resp.Error, custom_errors.ErrIncorrectEnterPassword.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		respAuth, errLogin := h.ServiceAuth.Login(body)
		resp.Error = append(resp.Error, errLogin...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == custom_errors.ErrIncorrectPasswordOrEmail.Error() {
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			} else if resp.Error[0] == custom_errors.ErrFailedSecurity.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respAuth
		h.ResponseSend(writer, resp, http.StatusAccepted)
	}
}
func (h *HandlerAuth) Recovery() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make([]string, 0, 5),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error = append(resp.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestLogin](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Email" {
						values.DataLog.MapLog["email"] = body.Email
						resp.Error = append(resp.Error, custom_errors.ErrIncorrectEmail.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		respAuth, errRecovery := h.ServiceAuth.Recovery(body.Email)
		resp.Error = append(resp.Error, errRecovery...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if resp.Error[0] == custom_errors.ErrFailedSecurity.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respAuth
		h.ResponseSend(writer, resp, http.StatusAccepted)
	}
}
func (h *HandlerAuth) Confirm() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make([]string, 0, 5),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error = append(resp.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[common.RequestConfirm](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Code" {
						resp.Error = append(resp.Error, custom_errors.ErrIncorrectFormatCode.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		action := request.URL.Query().Get("action")
		values.DataLog.MapLog["action"] = action
		userAgent := request.Header.Get("User-Agent")
		values.DataLog.MapLog["user_agent"] = userAgent
		respConfirm, errConfirm := h.ServiceAuth.Confirm(body.Code, values.DataAuth.SessionID, action, userAgent)
		resp.Error = append(resp.Error, errConfirm...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			switch resp.Error[0] {
			case ErrUserAlreadyExist.Error(), ErrIncorrectAction.Error():
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			case custom_errors.ErrSessionExpired.Error(), custom_errors.ErrIncorrectCode.Error(), custom_errors.ErrIncorrectSessionID.Error():
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			case custom_errors.ErrNotFoundUser.Error():
				h.ResponseSend(writer, resp, http.StatusNotFound)
			default:
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respConfirm
		if action == actionRegister {
			h.ResponseSend(writer, resp, http.StatusCreated)
		} else {
			h.ResponseSend(writer, resp, http.StatusOK)
		}
	}
}
func (h *HandlerAuth) Refresh() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make([]string, 0, 5),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error = append(resp.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestRefresh](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "RefreshJwt" {
						resp.Error = append(resp.Error, ErrSentRefresh.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		userAgent := request.Header.Get("User-Agent")
		values.DataLog.MapLog["user_agent"] = userAgent
		respConfirm, errConfirm := h.ServiceAuth.Refresh(body.RefreshJwt, userAgent)
		resp.Error = append(resp.Error, errConfirm...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == ErrRenewalRefresh.Error() {
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			} else {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respConfirm
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
