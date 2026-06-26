package auth

import (
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/middleware"
	"errors"
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
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestRegister](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 3)}
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Name":
						values.DataLog.MapLog["name"] = body.Name
						mapError.Map["name"] = ErrIncorrectName.Error()
					case err.Field() == "Email":
						values.DataLog.MapLog["email"] = body.Email
						mapError.Map["email"] = custom_errors.ErrIncorrectEmail.Error()
					case err.Field() == "Password":
						mapError.Map["password"] = custom_errors.ErrIncorrectEnterPassword.Error()
					}
				}
			} else {
				mapError.Map["body"] = errBody.Error()
			}
			values.DataLog.Errors = mapError.Error()
			resp.Error = mapError.Map
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		respAuth, errAuth := h.ServiceAuth.Register(body)
		if errAuth != nil {
			values.DataLog.Errors = errAuth.Error()
			switch {
			case errors.Is(errAuth, ErrUserAlreadyExist):
				resp.Error["user"] = errAuth.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			default:
				resp.Error["global"] = errAuth.Error()
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
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestLogin](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Email":
						values.DataLog.MapLog["email"] = body.Email
						mapError.Map["email"] = custom_errors.ErrIncorrectEmail.Error()
					case err.Field() == "Password":
						mapError.Map["password"] = custom_errors.ErrIncorrectEnterPassword.Error()
					}
				}
			} else {
				mapError.Map["body"] = errBody.Error()
			}
			values.DataLog.Errors = mapError.Error()
			resp.Error = mapError.Map
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		respAuth, errLogin := h.ServiceAuth.Login(body)
		if errLogin != nil {
			values.DataLog.Errors = errLogin.Error()
			switch {
			case errors.Is(errLogin, custom_errors.ErrIncorrectPasswordOrEmail):
				resp.Error["auth"] = errLogin.Error()
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			default:
				resp.Error["global"] = errLogin.Error()
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
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestRecovery](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Email":
						values.DataLog.MapLog["email"] = body.Email
						mapError.Map["email"] = custom_errors.ErrIncorrectEmail.Error()
					case err.Field() == "Password":
						mapError.Map["password"] = custom_errors.ErrIncorrectEnterPassword.Error()
					}
				}
			} else {
				mapError.Map["body"] = errBody.Error()
			}
			values.DataLog.Errors = mapError.Error()
			resp.Error = mapError.Map
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		action := request.URL.Query().Get("action")
		values.DataLog.MapLog["action"] = action
		respAuth, errRecovery := h.ServiceAuth.Recovery(body, action)
		if errRecovery != nil {
			values.DataLog.Errors = errRecovery.Error()
			var mapError shared_errors.MapError
			if errors.As(errRecovery, &mapError) {
				resp.Error = mapError.Map
				switch {
				case custom_errors.ErrNotFoundUser.Error() == mapError.Map["email"] && len(mapError.Map) == 1:
					h.ResponseSend(writer, resp, http.StatusNotFound)
				default:
					h.ResponseSend(writer, resp, http.StatusBadRequest)
				}
				return
			}
			switch {
			case errors.Is(errRecovery, custom_errors.ErrFailedSecurity):
				resp.Error["global"] = errRecovery.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			default:
				resp.Error["auth"] = errRecovery.Error()
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
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
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestConfirm](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Code" {
						mapError.Map["code"] = custom_errors.ErrIncorrectFormatCode.Error()
					}
					if err.Field() == "NewPassword" {
						mapError.Map["new_password"] = custom_errors.ErrIncorrectEnterNewPassword.Error()
					}
				}
			} else {
				mapError.Map["body"] = errBody.Error()
			}
			values.DataLog.Errors = mapError.Error()
			resp.Error = mapError.Map
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		action := request.URL.Query().Get("action")
		values.DataLog.MapLog["action"] = action
		userAgent := request.Header.Get("User-Agent")
		values.DataLog.MapLog["user_agent"] = userAgent
		respConfirm, errConfirm := h.ServiceAuth.Confirm(body, values.DataAuth.SessionID, action, userAgent)
		if errConfirm != nil {
			values.DataLog.Errors = errConfirm.Error()
			var mapError shared_errors.MapError
			if errors.As(errConfirm, &mapError) {
				resp.Error = mapError.Map
				h.ResponseSend(writer, resp, http.StatusBadRequest)
				return
			}
			switch {
			case errors.Is(errConfirm, ErrUserAlreadyExist):
				resp.Error["user"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			case errors.Is(errConfirm, ErrNotSpecifiedNewPassword):
				resp.Error["new_password"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			case errors.Is(errConfirm, custom_errors.ErrSessionExpired), errors.Is(errConfirm, custom_errors.ErrIncorrectCode):
				resp.Error["auth"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			case errors.Is(errConfirm, custom_errors.ErrNotFoundUser):
				resp.Error["user"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			default:
				resp.Error["global"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respConfirm
		if action == ActionRegister {
			h.ResponseSend(writer, resp, http.StatusCreated)
		} else {
			h.ResponseSend(writer, resp, http.StatusOK)
		}
	}
}
func (h *HandlerAuth) Refresh() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestRefresh](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "RefreshJwt" {
						values.DataLog.Errors = ErrSentRefresh.Error()
						resp.Error["refresh_jwt"] = ErrSentRefresh.Error()
					}
				}
			} else {
				values.DataLog.Errors = errBody.Error()
				resp.Error["body"] = errBody.Error()
			}
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		userAgent := request.Header.Get("User-Agent")
		values.DataLog.MapLog["user_agent"] = userAgent
		respConfirm, errRefresh := h.ServiceAuth.Refresh(body.RefreshJwt, userAgent)
		if errRefresh != nil {
			values.DataLog.Errors = errRefresh.Error()
			switch {
			case errors.Is(errRefresh, ErrRenewalRefresh):
				resp.Error["refresh_jwt"] = errRefresh.Error()
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			default:
				resp.Error["global"] = errRefresh.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respConfirm
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
