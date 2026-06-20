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
	*response.HandlerResponse
	*shared_middleware.ManagerSharedMiddleware
}

func NewHandlerExpense(router *http.ServeMux, service *ServiceExpense, logger *loggers.Logger, responseHandler *response.HandlerResponse, mv *shared_middleware.ManagerSharedMiddleware) {
	expense := &HandlerExpense{
		Logger:          logger,
		ServiceExpense:  service,
		HandlerResponse: responseHandler,
	}
	router.Handle("POST /api/v1/expense/{budget_uuid}", mv.HandlerAccessToken(expense.CreateExpense()))
	router.Handle("PATCH /api/v1/expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAccessToken(expense.UpdateExpense()))
	router.Handle("GET /api/v1/expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAccessToken(expense.GetExpense()))
	router.Handle("DELETE /api/v1/expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAccessToken(expense.RemoveExpense()))
	router.Handle("GET /api/v1/expense/{budget_uuid}", mv.HandlerAccessToken(expense.ListExpense()))
}

func (h *HandlerExpense) CreateExpense() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Category" {
						values.DataLog.MapLog["category"] = body.Category
						resp.Error = append(resp.Error, ErrIncorrectCategory.Error())
					} else if err.Field() == "Expense" {
						values.DataLog.MapLog["expense"] = body.Expense
						resp.Error = append(resp.Error, "expense"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Description" {
						values.DataLog.MapLog["description"] = body.Description
						resp.Error = append(resp.Error, ErrIncorrectDescription.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseCreate, errCreate := h.ServiceExpense.CreateExpense(body, values.DataAuth.UserUUID, budgetUUID)
		resp.Error = append(resp.Error, errCreate...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if len(resp.Error) == 1 && resp.Error[0] == ErrFailedCreateExpense.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = expenseCreate
		h.ResponseSend(writer, resp, http.StatusCreated)
	}
}

func (h *HandlerExpense) UpdateExpense() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					if err.Field() == "Category" {
						values.DataLog.MapLog["category"] = body.Category
						resp.Error = append(resp.Error, ErrIncorrectCategory.Error())
					} else if err.Field() == "Expense" {
						values.DataLog.MapLog["expense"] = body.Expense
						resp.Error = append(resp.Error, "expense"+custom_errors.ErrIncorrectDecimal.Error())
					} else if err.Field() == "Description" {
						values.DataLog.MapLog["description"] = body.Description
						resp.Error = append(resp.Error, ErrIncorrectDescription.Error())
					}
				}
			} else {
				resp.Error = append(resp.Error, errBody.Error())
			}
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			h.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expenseUpdate, errUpdate := h.ServiceExpense.UpdateExpense(body, values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		resp.Error = append(resp.Error, errUpdate...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() || resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if len(resp.Error) == 1 && resp.Error[0] == ErrFailedUpdateExpense.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = expenseUpdate
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerExpense) GetExpense() http.HandlerFunc {
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
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expense, errGet := h.ServiceExpense.GetExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		resp.Error = append(resp.Error, errGet...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() || resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = expense
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerExpense) RemoveExpense() http.HandlerFunc {
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
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		errDelete := h.ServiceExpense.DeleteExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		resp.Error = append(resp.Error, errDelete...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() || resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if len(resp.Error) == 1 && (resp.Error[0] == ErrFailedRemoveExpense.Error() || resp.Error[0] == ErrFailedDeleteExpense.Error()) {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
func (h *HandlerExpense) ListExpense() http.HandlerFunc {
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
		values.DataLog.MapLog["limit"] = limit
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["offset"] = offset
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseList, errList := h.ServiceExpense.ListExpense(budgetUUID, limit, offset)
		resp.Error = append(resp.Error, errList...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, resp.Error...)
			if len(resp.Error) == 1 && (resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() || resp.Error[0] == ErrNotFoundDescriptionExpense.Error()) {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = expenseList
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
