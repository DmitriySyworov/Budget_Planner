package user

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

type HandlerUser struct {
	*response.HandlerResponse
	*loggers.Logger
	*ServiceUser
}

func NewHandlerUser(router *http.ServeMux,
	service *ServiceUser,
	handlerResponse *response.HandlerResponse,
	logger *loggers.Logger,
	mv *middleware.ManagerMiddleware,
	sharedMv *shared_middleware.ManagerSharedMiddleware) {
	user := HandlerUser{
		ServiceUser:     service,
		HandlerResponse: handlerResponse,
		Logger:          logger,
	}
	router.Handle("PATCH /api/v1/user", sharedMv.HandlerAccessToken(user.UpdateUser()))
	router.Handle("GET /api/v1/user", sharedMv.HandlerAccessToken(user.GetUser()))
	router.Handle("DELETE /api/v1/user", sharedMv.HandlerAccessToken(user.RemoveUser()))
	router.Handle("POST /api/v1/user/confirm", sharedMv.HandlerAccessToken(mv.HandlerSessionToken(user.ConfirmUser())))
}
func (h *HandlerUser) UpdateUser() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[RequestUpdateUser](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 5)}
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Email":
						values.DataLog.MapLog["email"] = body.Email
						mapError.Map["email"] = custom_errors.ErrIncorrectEmail.Error()
					case err.Field() == "NewEmail":
						values.DataLog.MapLog["new_email"] = body.NewEmail
						mapError.Map["new_email"] = ErrIncorrectNewEmail.Error()
					case err.Field() == "NewName":
						values.DataLog.MapLog["new_name"] = body.NewName
						mapError.Map["new_name"] = ErrIncorrectNewName.Error()
					case err.Field() == "NewPassword":
						mapError.Map["new_password"] = custom_errors.ErrIncorrectEnterNewPassword.Error()
					case err.Field() == "Password":
						mapError.Map["password"] = custom_errors.ErrIncorrectEnterPassword.Error()
					}
				}
			} else {
				mapError.Map["body"] = errBody.Error()
			}
			resp.Error = mapError.Map
			values.DataLog.Errors = mapError.Error()
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		userUpdate, respAuth, errUpdateAuth := h.ServiceUser.UpdateUser(values.DataAuth.UserUUID, body)
		if errUpdateAuth != nil {
			values.DataLog.Errors = errUpdateAuth.Error()
			switch {
			case errors.Is(errUpdateAuth, custom_errors.ErrNotFoundUser):
				resp.Error["user"] = errUpdateAuth.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errUpdateAuth, custom_errors.ErrIncorrectPasswordOrEmail):
				resp.Error["auth"] = errUpdateAuth.Error()
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			case errors.Is(errUpdateAuth, ErrIncorrectChoiceEmail):
				resp.Error["email"] = errUpdateAuth.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			default:
				resp.Error["global"] = errUpdateAuth.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		if respAuth != nil {
			resp.Data = respAuth
			h.ResponseSend(writer, resp, http.StatusAccepted)
			return
		}
		if userUpdate != nil {
			resp.Data = userUpdate
			h.ResponseSend(writer, resp, http.StatusOK)
			return
		}
	}
}
func (h *HandlerUser) GetUser() http.HandlerFunc {
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
		respUser, errGetUser := h.ServiceUser.GetUser(values.DataAuth.UserUUID)
		if errGetUser != nil {
			values.DataLog.Errors = errGetUser.Error()
			if errors.Is(errGetUser, custom_errors.ErrNotFoundUser) {
				resp.Error["user"] = errGetUser.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else {
				resp.Error["global"] = errGetUser.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respUser
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerUser) RemoveUser() http.HandlerFunc {
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
		typeRemove := request.URL.Query().Get("type")
		values.DataLog.MapLog["type"] = typeRemove
		body, errBody := handler_request.HandlerRequest[RequestRemoveUser](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Email" {
						values.DataLog.MapLog["email"] = body.Email
						values.DataLog.Errors = custom_errors.ErrIncorrectEmail.Error()
						resp.Error["email"] = custom_errors.ErrIncorrectEmail.Error()
					}
					if err.Field() == "Password" {
						values.DataLog.Errors = custom_errors.ErrIncorrectEnterPassword.Error()
						resp.Error["password"] = custom_errors.ErrIncorrectEnterPassword.Error()
					}
				}
			} else {
				values.DataLog.Errors = errBody.Error()
				resp.Error["body"] = errBody.Error()
			}
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		respAuth, errRemoveAuth := h.ServiceUser.RemoveUser(body, typeRemove)
		if errRemoveAuth != nil {
			values.DataLog.Errors = errRemoveAuth.Error()
			switch {
			case errors.Is(errRemoveAuth, custom_errors.ErrIncorrectPasswordOrEmail):
				resp.Error["auth"] = errRemoveAuth.Error()
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			case errors.Is(errRemoveAuth, shared_errors.ErrIncorrectTypeRemove):
				resp.Error["action"] = errRemoveAuth.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			default:
				resp.Error["global"] = errRemoveAuth.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respAuth
		h.ResponseSend(writer, resp, http.StatusAccepted)
	}
}
func (h *HandlerUser) ConfirmUser() http.HandlerFunc {
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
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Code" {
						values.DataLog.Errors = custom_errors.ErrIncorrectFormatCode.Error()
						resp.Error["auth"] = custom_errors.ErrIncorrectFormatCode.Error()
					}
				}
			} else {
				values.DataLog.Errors = errBody.Error()
				resp.Error["body"] = errBody.Error()
			}
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		action := request.URL.Query().Get("action")
		values.DataLog.MapLog["action"] = action
		respUserUpdate, errConfirm := h.ServiceUser.ConfirmUser(body.Code, values.DataAuth.UserUUID, values.DataAuth.SessionID, action)
		if errConfirm != nil {
			values.DataLog.Errors = errConfirm.Error()
			var mapError shared_errors.MapError
			if errors.As(errConfirm, &mapError) {
				resp.Error = mapError.Map
				h.ResponseSend(writer, resp, http.StatusBadRequest)
				return
			}
			switch {
			case errors.Is(errConfirm, custom_errors.ErrNotFoundUser):
				resp.Error["user"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errConfirm, ErrFailedRemoveUser), errors.Is(errConfirm, ErrFailedDeleteUser), errors.Is(errConfirm, ErrFailedUpdateUser):
				resp.Error["global"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			default:
				resp.Error["auth"] = errConfirm.Error()
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			}
			return
		}
		if respUserUpdate != nil {
			resp.Success = true
			resp.Data = respUserUpdate
			h.ResponseSend(writer, resp, http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusNoContent)
		}
	}
}
