package budget

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/handler_request"
	"app/budget-planner/internal/loggers"
	"app/budget-planner/internal/middleware"
	"app/budget-planner/internal/response"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type HandlerBudget struct {
	*ServiceBudget
	*loggers.Logger
	*response.HandlerResponse
}

func NewHandlerBudget(router *http.ServeMux, service *ServiceBudget, logger *loggers.Logger, responseHandler *response.HandlerResponse, mv *middleware.ManagerMiddleware) {
	budget := &HandlerBudget{
		ServiceBudget:   service,
		Logger:          logger,
		HandlerResponse: responseHandler,
	}
	router.Handle("POST /budget", mv.HandlerAuthToken(budget.CreateBudget()))
	router.Handle("PATCH /budget/{uuid}", mv.HandlerAuthToken(budget.UpdateBudget()))
	router.Handle("GET /budget/{uuid}", mv.HandlerAuthToken(budget.GetBudget()))
	router.Handle("DELETE /budget/{uuid}", mv.HandlerAuthToken(budget.DeleteBudget()))
	router.Handle("GET /budget", mv.HandlerAuthToken(budget.ListBudget()))
}
func (h *HandlerBudget) CreateBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateBudget](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Amount" {
						values.DataLog.MapLog["amount"] = body.Amount
						h.Response.Error = append(h.Response.Error, "amount"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Start" {
						values.DataLog.MapLog["start"] = body.Start
						h.Response.Error = append(h.Response.Error, ErrIncorrectStart.Error())
					} else if err.Field() == "Finish" {
						values.DataLog.MapLog["finish"] = body.Finish
						h.Response.Error = append(h.Response.Error, ErrIncorrectFinish.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
		}
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.HandlerResponse.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		budgetCreate, errCreate := h.ServiceBudget.CreateBudget(body, values.DataAuth.UserUUID)
		h.Response.Error = append(h.Response.Error, errCreate...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Response.Error) == 1 && h.Response.Error[0] == ErrFailedCreateBudget.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = budgetCreate
		h.ResponseSend(writer, http.StatusCreated)
	}
}

func (h *HandlerBudget) UpdateBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateBudget](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Amount" {
						values.DataLog.MapLog["amount"] = body.Amount
						h.Response.Error = append(h.Response.Error, "amount"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Start" {
						values.DataLog.MapLog["start"] = body.Start
						h.Response.Error = append(h.Response.Error, ErrIncorrectStart.Error())
					} else if err.Field() == "Finish" {
						values.DataLog.MapLog["finish"] = body.Finish
						h.Response.Error = append(h.Response.Error, ErrIncorrectFinish.Error())
					}
				}
			} else {
				h.Response.Error = append(h.Response.Error, errBody.Error())
			}
		}
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.HandlerResponse.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		budgetUpdate, errUpdate := h.ServiceBudget.UpdateBudget(body, values.DataAuth.UserUUID, budgetUUID)
		h.Response.Error = append(h.Response.Error, errUpdate...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Response.Error) == 1 && h.Response.Error[0] == ErrFailedUpdateBudget.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = budgetUpdate
		h.ResponseSend(writer, http.StatusOK)
	}
}
func (h *HandlerBudget) GetBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		budget, errGetBudget := h.ServiceBudget.GetBudget(values.DataAuth.UserUUID, budgetUUID)
		h.Response.Error = append(h.Response.Error, errGetBudget...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = budget
		h.ResponseSend(writer, http.StatusOK)
	}
}
func (h *HandlerBudget) DeleteBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		typeRemove := request.URL.Query().Get("type")
		values.DataLog.MapLog["type"] = typeRemove
		errRemoveBudget := h.ServiceBudget.RemoveBudget(values.DataAuth.UserUUID, budgetUUID, typeRemove)
		h.Response.Error = append(h.Response.Error, errRemoveBudget...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Response.Error) == 1 && h.Response.Error[0] == ErrFailedDeleteBudget.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
func (h *HandlerBudget) ListBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		limit := request.URL.Query().Get("limit")
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["limit"] = limit
		values.DataLog.MapLog["offset"] = offset
		budgetList, errList := h.ServiceBudget.ListBudget(values.DataAuth.UserUUID, limit, offset)
		h.Response.Error = append(h.Response.Error, errList...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = budgetList
		h.ResponseSend(writer, http.StatusOK)
	}
}
