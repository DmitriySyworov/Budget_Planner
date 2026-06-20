package user

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
	router.Handle("DELETE /api/v1/user", sharedMv.HandlerAccessToken(user.DeleteUser()))
	router.Handle("POST /api/v1/user/confirm", sharedMv.HandlerAccessToken(mv.HandlerSessionToken(user.ConfirmUser())))
}
func (h *HandlerUser) UpdateUser() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[RequestUpdateUser](request.Body)
		if errBody != nil {
			if errValidate, okErrValidate := errBody.(validator.ValidationErrors); okErrValidate {
				for _, err := range errValidate {
					if err.Field() == "Email" {
						values.DataLog.MapLog["email"] = body.Email
						resp.Error = append(resp.Error, custom_errors.ErrIncorrectEmail.Error())
					}
					if err.Field() == "NewEmail" {
						values.DataLog.MapLog["new_email"] = body.NewEmail
						resp.Error = append(resp.Error, ErrIncorrectNewEmail.Error())
					}
					if err.Field() == "NewName" {
						values.DataLog.MapLog["new_name"] = body.NewName
						resp.Error = append(resp.Error, ErrIncorrectNewName.Error())
					}
					if err.Field() == "NewPassword" {
						resp.Error = append(resp.Error, ErrIncorrectEnterNewPassword.Error())
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
		userUpdate, respAuth, errUpdateAuth := h.ServiceUser.UpdateUser(values.DataAuth.UserUUID, body)
		resp.Error = append(resp.Error, errUpdateAuth...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if resp.Error[0] == custom_errors.ErrIncorrectPasswordOrEmail.Error() {
				h.ResponseSend(writer, resp, http.StatusUnauthorized)
			} else if resp.Error[0] == custom_errors.ErrFailedSecurity.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
				return
			}
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
		respUser, errGetUser := h.ServiceUser.GetUser(values.DataAuth.UserUUID)
		resp.Error = append(resp.Error, errGetUser...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = respUser
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerUser) DeleteUser() http.HandlerFunc {
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
		typeRemove := request.URL.Query().Get("type")
		values.DataLog.MapLog["type"] = typeRemove
		body, errBody := handler_request.HandlerRequest[RequestRemoveUser](request.Body)
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
		respAuth, errRemoveAuth := h.ServiceUser.DeleteUser(body.Email, typeRemove)
		resp.Error = append(resp.Error, errRemoveAuth...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if resp.Error[0] == custom_errors.ErrFailedSecurity.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
				return
			}
		}
		resp.Success = true
		resp.Data = respAuth
		h.ResponseSend(writer, resp, http.StatusAccepted)
	}
}
func (h *HandlerUser) ConfirmUser() http.HandlerFunc {
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
		respUserUpdate, errConfirm := h.ServiceUser.ConfirmUser(body.Code, values.DataAuth.UserUUID, values.DataAuth.SessionID, action)
		resp.Error = append(resp.Error, errConfirm...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if resp.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if resp.Error[0] == ErrFailedRemoveUser.Error() || resp.Error[0] == ErrFailedDeleteUser.Error() || resp.Error[0] == ErrFailedUpdateUser.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else if resp.Error[0] == custom_errors.ErrIncorrectSessionID.Error() || resp.Error[0] == ErrIncorrectAction.Error() {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			} else {
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
