package expense

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

type HandlerExpense struct {
	*ServiceExpense
	*loggers.Logger
	Resp *response.HandlerResponse
	*shared_middleware.ManagerMiddleware
}

func NewHandlerExpense(router *http.ServeMux, service *ServiceExpense, logger *loggers.Logger, responseHandler *response.HandlerResponse, mv *shared_middleware.ManagerMiddleware) {
	expense := &HandlerExpense{
		Logger:         logger,
		ServiceExpense: service,
		Resp:           responseHandler,
	}
	router.Handle("POST /api/v1/expense/{budget_uuid}", mv.HandlerAuthToken(expense.CreateExpense()))
	router.Handle("PATCH /api/v1/expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAuthToken(expense.UpdateExpense()))
	router.Handle("GET /api/v1/expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAuthToken(expense.GetExpense()))
	router.Handle("DELETE /api/v1/expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAuthToken(expense.RemoveExpense()))
	router.Handle("GET /api/v1/expense/{budget_uuid}", mv.HandlerAuthToken(expense.ListExpense()))
}

func (h *HandlerExpense) CreateExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Resp.Error = append(h.Resp.Error, shared_errors.ErrCriticalServer.Error())
			h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Category" {
						values.DataLog.MapLog["category"] = body.Category
						h.Resp.Error = append(h.Resp.Error, ErrIncorrectCategory.Error())
					} else if err.Field() == "Expense" {
						values.DataLog.MapLog["expense"] = body.Expense
						h.Resp.Error = append(h.Resp.Error, "expense"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Description" {
						values.DataLog.MapLog["description"] = body.Description
						h.Resp.Error = append(h.Resp.Error, ErrIncorrectDescription.Error())
					}
				}
			} else {
				h.Resp.Error = append(h.Resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.Resp.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseCreate, errCreate := h.ServiceExpense.CreateExpense(body, values.DataAuth.UserUUID, budgetUUID)
		h.Resp.Error = append(h.Resp.Error, errCreate...)
		if len(h.Resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Resp.Error...)
			if len(h.Resp.Error) == 1 && (h.Resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.Resp.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Resp.Error) == 1 && h.Resp.Error[0] == ErrFailedCreateExpense.Error() {
				h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.Resp.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Resp.Success = true
		h.Resp.Data = expenseCreate
		h.Resp.ResponseSend(writer, http.StatusCreated)
	}
}

func (h *HandlerExpense) UpdateExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Resp.Error = append(h.Resp.Error, shared_errors.ErrCriticalServer.Error())
			h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Category" {
						values.DataLog.MapLog["category"] = body.Category
						h.Resp.Error = append(h.Resp.Error, ErrIncorrectCategory.Error())
					} else if err.Field() == "Expense" {
						values.DataLog.MapLog["expense"] = body.Expense
						h.Resp.Error = append(h.Resp.Error, "expense"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Description" {
						values.DataLog.MapLog["description"] = body.Description
						h.Resp.Error = append(h.Resp.Error, ErrIncorrectDescription.Error())
					}
				}
			} else {
				h.Resp.Error = append(h.Resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.Resp.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expenseUpdate, errUpdate := h.ServiceExpense.UpdateExpense(body, values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		h.Resp.Error = append(h.Resp.Error, errUpdate...)
		if len(h.Resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Resp.Error...)
			if len(h.Resp.Error) == 1 && (h.Resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.Resp.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Resp.Error) == 1 && h.Resp.Error[0] == ErrFailedUpdateExpense.Error() {
				h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.Resp.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Resp.Success = true
		h.Resp.Data = expenseUpdate
		h.Resp.ResponseSend(writer, http.StatusOK)
	}
}
func (h *HandlerExpense) GetExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Resp.Error = append(h.Resp.Error, shared_errors.ErrCriticalServer.Error())
			h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expense, errGet := h.ServiceExpense.GetExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		h.Resp.Error = append(h.Resp.Error, errGet...)
		if len(h.Resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Resp.Error...)
			if len(h.Resp.Error) == 1 && (h.Resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.Resp.ResponseSend(writer, http.StatusNotFound)
			} else {
				h.Resp.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Resp.Success = true
		h.Resp.Data = expense
		h.Resp.ResponseSend(writer, http.StatusOK)
	}
}
func (h *HandlerExpense) RemoveExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Resp.Error = append(h.Resp.Error, shared_errors.ErrCriticalServer.Error())
			h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		errDelete := h.ServiceExpense.DeleteExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		h.Resp.Error = append(h.Resp.Error, errDelete...)
		if len(h.Resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Resp.Error...)
			if len(h.Resp.Error) == 1 && (h.Resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.Resp.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Resp.Error) == 1 && (h.Resp.Error[0] == ErrFailedRemoveExpense.Error() || h.Resp.Error[0] == ErrFailedDeleteExpense.Error()) {
				h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.Resp.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
func (h *HandlerExpense) ListExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(shared_middleware.KeyContextValue)
		values, ok := ctxValues.(*shared_middleware.ContextValues)
		if !ok {
			h.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Resp.Error = append(h.Resp.Error, shared_errors.ErrCriticalServer.Error())
			h.Resp.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		limit := request.URL.Query().Get("limit")
		values.DataLog.MapLog["limit"] = limit
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["offset"] = offset
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseList, errList := h.ServiceExpense.ListExpense(budgetUUID, limit, offset)
		h.Resp.Error = append(h.Resp.Error, errList...)
		if len(h.Resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Resp.Error...)
			if len(h.Resp.Error) == 1 && (h.Resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.Resp.ResponseSend(writer, http.StatusNotFound)
			} else {
				h.Resp.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Resp.Success = true
		h.Resp.Data = expenseList
		h.Resp.ResponseSend(writer, http.StatusOK)
	}
}
