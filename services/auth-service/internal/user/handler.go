package user

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

type HandlerUser struct {
	response.Response
	*response.HandlerResponse
	*loggers.Logger
	*ServiceUser
}

func NewHandlerUser(router *http.ServeMux, service *ServiceUser, handlerResponse *response.HandlerResponse, logger *loggers.Logger, mv *shared_middleware.ManagerMiddleware) {
	user := HandlerUser{
		ServiceUser:     service,
		HandlerResponse: handlerResponse,
		Logger:          logger,
	}
	router.Handle("PATCH /api/v1/user", mv.HandlerAuthToken(user.UpdateUser()))
	router.Handle("GET /api/v1/user", mv.HandlerAuthToken(user.GetUser()))
	router.Handle("DELETE /api/v1/user", mv.HandlerAuthToken(user.DeleteUser()))
	router.Handle("POST /api/v1/user/confirm", mv.HandlerAuthToken(mv.HandlerSessionToken(user.ConfirmUser())))
}
func (h *HandlerUser) UpdateUser() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[RequestUpdateUser](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Email" {
						values.DataLog.MapLog["email"] = body.Email
						h.Response.Error = append(h.Response.Error, custom_errors.ErrIncorrectEmail.Error())
					}
					if err.Field() == "NewEmail" {
						values.DataLog.MapLog["new_email"] = body.NewEmail
						h.Response.Error = append(h.Response.Error, ErrIncorrectNewEmail.Error())
					}
					if err.Field() == "NewName" {
						values.DataLog.MapLog["new_name"] = body.NewName
						h.Response.Error = append(h.Response.Error, ErrIncorrectNewName.Error())
					}
					if err.Field() == "NewPassword" {
						values.DataLog.MapLog["new_password"] = body.NewPassword
						h.Response.Error = append(h.Response.Error, ErrIncorrectEnterNewPassword.Error())
					}
					if err.Field() == "Password" {
						values.DataLog.MapLog["password"] = body.NewPassword
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
		userUpdate, respAuth, errUpdateAuth := h.ServiceUser.UpdateUser(values.DataAuth.UserUUID, body)
		h.Response.Error = append(h.Response.Error, errUpdateAuth...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if h.Response.Error[0] == custom_errors.ErrIncorrectPasswordOrEmail.Error() {
				h.ResponseSend(writer, http.StatusUnauthorized)
			} else if h.Response.Error[0] == custom_errors.ErrFailedSecurity.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
				return
			}
		}
		h.Response.Success = true
		if respAuth != nil {
			h.Response.Data = respAuth
			h.ResponseSend(writer, http.StatusAccepted)
			return
		}
		if userUpdate != nil {
			h.Response.Data = userUpdate
			h.ResponseSend(writer, http.StatusOK)
			return
		}
	}
}
func (h *HandlerUser) GetUser() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		respUser, errGetUser := h.ServiceUser.GetUser(values.DataAuth.UserUUID)
		h.Response.Error = append(h.Response.Error, errGetUser...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, http.StatusInternalServerError)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = respUser
		h.ResponseSend(writer, http.StatusOK)
	}
}
func (h *HandlerUser) DeleteUser() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, shared_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
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
						h.Response.Error = append(h.Response.Error, custom_errors.ErrIncorrectEmail.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		respAuth, errRemoveAuth := h.ServiceUser.DeleteUser(body.Email, typeRemove)
		h.Response.Error = append(h.Response.Error, errRemoveAuth...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if h.Response.Error[0] == custom_errors.ErrFailedSecurity.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
				return
			}
		}
		h.Response.Success = true
		h.Response.Data = respAuth
		h.ResponseSend(writer, http.StatusAccepted)
	}
}
func (h *HandlerUser) ConfirmUser() http.HandlerFunc {
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
						values.DataLog.MapLog["code"] = body.Code
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
		respUserUpdate, errConfirm := h.ServiceUser.ConfirmUser(body.Code, values.DataAuth.UserUUID, values.DataAuth.SessionID, action)
		h.Response.Error = append(h.Response.Error, errConfirm...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if h.Response.Error[0] == ErrFailedRemoveUser.Error() || h.Response.Error[0] == ErrFailedDeleteUser.Error() || h.Response.Error[0] == ErrFailedUpdateUser.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else if h.Response.Error[0] == custom_errors.ErrIncorrectSessionID.Error() || h.Response.Error[0] == ErrIncorrectAction.Error() {
				h.ResponseSend(writer, http.StatusBadRequest)
			} else {
				h.ResponseSend(writer, http.StatusUnauthorized)
			}
			return
		}
		if respUserUpdate != nil {
			h.Response.Success = true
			h.Response.Data = respUserUpdate
			h.ResponseSend(writer, http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusNoContent)
		}
	}
}
