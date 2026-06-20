package budget

import (
	"app/budget-planner/internal/custom_errors"
	"net/http"
	"shared/handler_request"
	"shared/loggers"
	"shared/response"
	"shared/shared_errors"
	"shared/shared_middleware"

	"github.com/go-playground/validator/v10"
)

type HandlerBudget struct {
	*ServiceBudget
	*loggers.Logger
	*response.HandlerResponse
}

func NewHandlerBudget(router *http.ServeMux, service *ServiceBudget, logger *loggers.Logger, responseHandler *response.HandlerResponse, mv *shared_middleware.ManagerSharedMiddleware) {
	budget := &HandlerBudget{
		ServiceBudget:   service,
		Logger:          logger,
		HandlerResponse: responseHandler,
	}
	router.Handle("POST /budget", mv.HandlerAccessToken(budget.CreateBudget()))
	router.Handle("PATCH /budget/{uuid}", mv.HandlerAccessToken(budget.UpdateBudget()))
	router.Handle("GET /budget/{uuid}", mv.HandlerAccessToken(budget.GetBudget()))
	router.Handle("DELETE /budget/{uuid}", mv.HandlerAccessToken(budget.DeleteBudget()))
	router.Handle("GET /budget", mv.HandlerAccessToken(budget.ListBudget()))
}
func (h *HandlerBudget) CreateBudget() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateBudget](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Amount" {
						values.DataLog.MapLog["amount"] = body.Amount
						resp.Error = append(resp.Error, "amount"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Start" {
						values.DataLog.MapLog["start"] = body.Start
						resp.Error = append(resp.Error, ErrIncorrectStart.Error())
					} else if err.Field() == "Finish" {
						values.DataLog.MapLog["finish"] = body.Finish
						resp.Error = append(resp.Error, ErrIncorrectFinish.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.HandlerResponse.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		budgetCreate, errCreate := h.ServiceBudget.CreateBudget(body, values.DataAuth.UserUUID)
		resp.Error = append(resp.Error, errCreate...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && resp.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if len(resp.Error) == 1 && resp.Error[0] == ErrFailedCreateBudget.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = budgetCreate
		h.ResponseSend(writer, resp, http.StatusCreated)
	}
}

func (h *HandlerBudget) UpdateBudget() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateBudget](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Amount" {
						values.DataLog.MapLog["amount"] = body.Amount
						resp.Error = append(resp.Error, "amount"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Start" {
						values.DataLog.MapLog["start"] = body.Start
						resp.Error = append(resp.Error, ErrIncorrectStart.Error())
					} else if err.Field() == "Finish" {
						values.DataLog.MapLog["finish"] = body.Finish
						resp.Error = append(resp.Error, ErrIncorrectFinish.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.HandlerResponse.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		budgetUpdate, errUpdate := h.ServiceBudget.UpdateBudget(body, values.DataAuth.UserUUID, budgetUUID)
		resp.Error = append(resp.Error, errUpdate...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if len(resp.Error) == 1 && resp.Error[0] == ErrFailedUpdateBudget.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = budgetUpdate
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerBudget) GetBudget() http.HandlerFunc {
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
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		budget, errGetBudget := h.ServiceBudget.GetBudget(values.DataAuth.UserUUID, budgetUUID)
		resp.Error = append(resp.Error, errGetBudget...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = budget
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerBudget) DeleteBudget() http.HandlerFunc {
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
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		typeRemove := request.URL.Query().Get("type")
		values.DataLog.MapLog["type"] = typeRemove
		errRemoveBudget := h.ServiceBudget.RemoveBudget(values.DataAuth.UserUUID, budgetUUID, typeRemove)
		resp.Error = append(resp.Error, errRemoveBudget...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if len(resp.Error) == 1 && resp.Error[0] == ErrFailedDeleteBudget.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
func (h *HandlerBudget) ListBudget() http.HandlerFunc {
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
		limit := request.URL.Query().Get("limit")
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["limit"] = limit
		values.DataLog.MapLog["offset"] = offset
		budgetList, errList := h.ServiceBudget.ListBudget(values.DataAuth.UserUUID, limit, offset)
		resp.Error = append(resp.Error, errList...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = budgetList
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
