package expense

import (
	"app/budget-planner/internal/custom_errors"
	"errors"
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
	router.Handle("DELETE /api/v1/expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAccessToken(expense.DeleteExpense()))
	router.Handle("GET /api/v1/expense/{budget_uuid}", mv.HandlerAccessToken(expense.ListExpense()))
}

func (h *HandlerExpense) CreateExpense() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Category":
						values.DataLog.MapLog["category"] = body.Category
						mapError.Map["category"] = ErrIncorrectCategory.Error()
					case err.Field() == "Expense":
						values.DataLog.MapLog["expense"] = body.Expense
						mapError.Map["expense"] = "expense" + custom_errors.ErrIncorrectDecimal.Error()
					case err.Field() == "Description":
						values.DataLog.MapLog["description"] = body.Description
						mapError.Map["description"] = ErrIncorrectDescription.Error()
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
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseCreate, errCreate := h.ServiceExpense.CreateExpense(body, values.DataAuth.UserUUID, budgetUUID)
		if errCreate != nil {
			values.DataLog.Errors = errCreate.Error()
			switch {
			case errors.Is(errCreate, custom_errors.ErrNotFoundBudget):
				resp.Error["budget"] = errCreate.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errCreate, ErrFailedCreateExpense):
				resp.Error["global"] = errCreate.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			default:
				resp.Error["budget"] = errCreate.Error()
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
		body, errBody := handler_request.HandlerRequest[CreateAndUpdateExpense](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Category":
						values.DataLog.MapLog["category"] = body.Category
						mapError.Map["category"] = ErrIncorrectCategory.Error()
					case err.Field() == "Expense":
						values.DataLog.MapLog["expense"] = body.Expense
						mapError.Map["expense"] = "expense" + custom_errors.ErrIncorrectDecimal.Error()
					case err.Field() == "Description":
						values.DataLog.MapLog["description"] = body.Description
						mapError.Map["description"] = ErrIncorrectDescription.Error()
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
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expenseUpdate, errUpdate := h.ServiceExpense.UpdateExpense(body, values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		if errUpdate != nil {
			values.DataLog.Errors = errUpdate.Error()
			switch {
			case errors.Is(errUpdate, custom_errors.ErrNotFoundBudget):
				resp.Error["budget"] = errUpdate.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errUpdate, custom_errors.ErrNotFoundExpense), errors.Is(errUpdate, ErrNotFoundDescriptionExpense):
				resp.Error["expense"] = errUpdate.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errUpdate, ErrFailedUpdateExpense):
				resp.Error["global"] = errUpdate.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			default:
				resp.Error["budget"] = errUpdate.Error()
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
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expense, errGetExpense := h.ServiceExpense.GetExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		if errGetExpense != nil {
			values.DataLog.Errors = errGetExpense.Error()
			switch {
			case errors.Is(errGetExpense, custom_errors.ErrNotFoundBudget):
				resp.Error["budget"] = errGetExpense.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errGetExpense, custom_errors.ErrNotFoundExpense), errors.Is(errGetExpense, ErrNotFoundDescriptionExpense):
				resp.Error["expense"] = errGetExpense.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			default:
				resp.Error["budget"] = errGetExpense.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = expense
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerExpense) DeleteExpense() http.HandlerFunc {
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
		budgetUUID := request.PathValue("budget_uid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		errDelete := h.ServiceExpense.DeleteExpense(values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		if errDelete != nil {
			values.DataLog.Errors = errDelete.Error()
			var mapError shared_errors.MapError
			if errors.As(errDelete, &mapError) {
				resp.Error = mapError.Map
				switch {
				case mapError.Map["budget"] == custom_errors.ErrIncorrectFormatBudgetUUID.Error() && len(mapError.Map) == 1:
					h.ResponseSend(writer, resp, http.StatusBadRequest)
				default:
					h.ResponseSend(writer, resp, http.StatusNotFound)
				}
			} else {
				resp.Error["global"] = errDelete.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
func (h *HandlerExpense) ListExpense() http.HandlerFunc {
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
		limit := request.URL.Query().Get("limit")
		values.DataLog.MapLog["limit"] = limit
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["offset"] = offset
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseList, errList := h.ServiceExpense.ListExpense(budgetUUID, limit, offset)
		if errList != nil {
			values.DataLog.Errors = errList.Error()
			var mapError shared_errors.MapError
			if errors.As(errList, &mapError) {
				resp.Error = mapError.Map
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			} else {
				resp.Error["expense"] = errList.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			}
			return
		}
		resp.Success = true
		resp.Data = expenseList
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
