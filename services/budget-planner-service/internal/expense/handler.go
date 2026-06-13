package expense

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/handler_request"
	"app/budget-planner/internal/middleware"
	"app/budget-planner/internal/response"
	"net/http"
	"shared/loggers"

	"github.com/go-playground/validator/v10"
)

type HandlerExpense struct {
	*ServiceExpense
	*loggers.Logger
	Resp *response.HandlerResponse
	*middleware.ManagerMiddleware
}

func NewHandlerExpense(router *http.ServeMux, service *ServiceExpense, logger *loggers.Logger, responseHandler *response.HandlerResponse, mv *middleware.ManagerMiddleware) {
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
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, custom_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Category" {
						values.DataLog.MapLog["category"] = body.Category
						h.Response.Error = append(h.Response.Error, ErrIncorrectCategory.Error())
					} else if err.Field() == "Expense" {
						values.DataLog.MapLog["expense"] = body.Expense
						h.Response.Error = append(h.Response.Error, "expense"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Description" {
						values.DataLog.MapLog["description"] = body.Description
						h.Response.Error = append(h.Response.Error, ErrIncorrectDescription.Error())
					}
				}
			} else {
				h.HandlerResponse.Response.Error = append(h.Response.Error, errBody.Error())
			}
		}
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.HandlerResponse.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		expenseCreate, errCreate := h.ServiceExpense.CreateExpense(body, values.DataAuth.UserUUID, budgetUUID)
		h.Response.Error = append(h.Response.Error, errCreate...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Response.Error) == 1 && h.Response.Error[0] == ErrFailedCreateExpense.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = expenseCreate
		h.ResponseSend(writer, http.StatusCreated)
	}
}

func (h *HandlerExpense) UpdateExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, custom_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Category" {
						values.DataLog.MapLog["category"] = body.Category
						h.Response.Error = append(h.Response.Error, ErrIncorrectCategory.Error())
					} else if err.Field() == "Expense" {
						values.DataLog.MapLog["expense"] = body.Expense
						h.Response.Error = append(h.Response.Error, "expense"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Description" {
						values.DataLog.MapLog["description"] = body.Description
						h.Response.Error = append(h.Response.Error, ErrIncorrectDescription.Error())
					}
				}
			} else {
				h.HandlerResponse.Response.Error = append(h.Response.Error, errBody.Error())
			}
		}
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			h.HandlerResponse.ResponseSend(writer, http.StatusBadRequest)
			return
		}
		expenseUpdate, errUpdate := h.ServiceExpense.UpdateExpense(body, values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		h.Response.Error = append(h.Response.Error, errUpdate...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Response.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Response.Error) == 1 && h.Response.Error[0] == ErrFailedUpdateExpense.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = expenseUpdate
		h.ResponseSend(writer, http.StatusOK)
	}
}
func (h *HandlerExpense) GetExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, custom_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expense, errGet := h.ServiceExpense.GetExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		h.Response.Error = append(h.Response.Error, errGet...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Response.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = expense
		h.ResponseSend(writer, http.StatusOK)
	}
}
func (h *HandlerExpense) RemoveExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, custom_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		errDelete := h.ServiceExpense.DeleteExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		h.Response.Error = append(h.Response.Error, errDelete...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Response.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if len(h.Response.Error) == 1 && (h.Response.Error[0] == ErrFailedRemoveExpense.Error() || h.Response.Error[0] == ErrFailedDeleteExpense.Error()) {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
func (h *HandlerExpense) ListExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctxValues := request.Context().Value(middleware.KeyContextValue)
		values, ok := ctxValues.(*middleware.ContextValues)
		if !ok {
			h.Logger.Error(custom_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			h.Response.Error = append(h.Response.Error, custom_errors.ErrCriticalServer.Error())
			h.ResponseSend(writer, http.StatusInternalServerError)
			return
		}
		values.DataLog.UserUUID = values.DataAuth.UserUUID
		limit := request.URL.Query().Get("limit")
		values.DataLog.MapLog["limit"] = limit
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["offset"] = offset
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseList, errList := h.ServiceExpense.ListExpense(values.DataAuth.UserUUID, budgetUUID, limit, offset)
		h.Response.Error = append(h.Response.Error, errList...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, h.Response.Error...)
			if len(h.Response.Error) == 1 && (h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundExpense.Error() || h.Response.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
			return
		}
		h.Response.Success = true
		h.Response.Data = expenseList
		h.ResponseSend(writer, http.StatusOK)
	}
}
