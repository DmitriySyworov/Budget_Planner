package expense

import (
	"app/budget-planner/internal/apperrors"
	"errors"
	"net/http"
	"shared/loggers"
	"shared/requestutil"
	"shared/response"
	"shared/sherrors"
	"shared/shmiddleware"

	_ "app/budget-planner/internal/model"

	"github.com/go-playground/validator/v10"
)

type HandlerExpense struct {
	*ServiceExpense
	*loggers.Logger
	*validator.Validate
	*response.HandlerResponse
	*shmiddleware.ManagerSharedMiddleware
}

func NewHandlerExpense(router *http.ServeMux, service *ServiceExpense, logger *loggers.Logger, responseHandler *response.HandlerResponse, validate *validator.Validate, mv *shmiddleware.ManagerSharedMiddleware) {
	expense := &HandlerExpense{
		Logger:          logger,
		ServiceExpense:  service,
		Validate:        validate,
		HandlerResponse: responseHandler,
	}
	router.Handle("POST /api/v1/description-expense/{budget_uuid}", mv.HandlerAccessToken(expense.CreateExpense()))
	router.Handle("PATCH /api/v1/description-expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAccessToken(expense.UpdateExpense()))
	router.Handle("GET /api/v1/description-expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAccessToken(expense.GetDescriptionExpense()))
	router.Handle("DELETE /api/v1/description-expense/{budget_uuid}/{description_expense_uuid}", mv.HandlerAccessToken(expense.DeleteDescriptionExpense()))
	router.Handle("GET /api/v1/expense/{budget_uuid}", mv.HandlerAccessToken(expense.GetExpense()))
	router.Handle("GET /api/v1/description-expense/{budget_uuid}", mv.HandlerAccessToken(expense.ListDescriptionExpense()))
}

// CreateExpense godoc
// @Summary      Create a new expense transaction under a budget
// @Description  Creates a detailed expense record and associates it with a specific budget by UUID. If the parent expense container does not exist, it initializes one automatically. Validates transaction category constraints and overall budget existence.
// @Tags         expense
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                             true  "Bearer <access_token>"
// @Param        budget_uuid    path      string                             true  "Parent Budget UUID (36 characters)"
// @Param        request        body      RequestCreateDescriptionExpense    true  "Expense creation payload"
// @Success      201      {object}  response.Response{data=ResponseCreateAndUpdateExpense,errors=nil} "Expense transaction successfully recorded"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"expense\": \"expense must be a positive decimals and greater than 0\" } } or { \"errors\": { \"category\": \"category mast be a health, sport, supermarket, restaurant, leisure, investments, savings or other\" } } or { \"errors\": { \"description\": \"description cannot be more than 250 characters\" } } or { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } } or { \"errors\": { \"body\": \"<raw_body_error>\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"global\": \"failed to create expense\" } }"
// @Router       /api/v1/description-expense/{budget_uuid} [post]
func (h *HandlerExpense) CreateExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shmiddleware.KeyContextValue)
		values, ok := ctxValues.(*shmiddleware.ContextValues)
		if !ok {
			h.Logger.Error(sherrors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = sherrors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := requestutil.HandlerRequest[RequestCreateDescriptionExpense](request.Body, h.Validate)
		if errBody != nil {
			mapError := sherrors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Category":
						values.DataLog.MapLog["category"] = body.Category
						mapError.Map["category"] = ErrIncorrectCategory.Error()
					case err.Field() == "Expense":
						values.DataLog.MapLog["expense"] = body.Expense
						mapError.Map["expense"] = "expense" + apperrors.ErrIncorrectDecimal.Error()
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
		expenseCreate, errCreate := h.ServiceExpense.CreateExpense(request.Context(), body, values.DataAuth.UserUUID, budgetUUID)
		if errCreate != nil {
			values.DataLog.Errors = errCreate.Error()
			switch {
			case errors.Is(errCreate, apperrors.ErrNotFoundBudget):
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

// UpdateExpense godoc
// @Summary      Update an existing expense transaction
// @Description  Updates an individual expense transaction's category, expense amount, or description by its UUID under a specific budget. Validates the existence of the parent budget, the main expense container, and the specific transaction record.
// @Tags         expense
// @Accept       json
// @Produce      json
// @Param        Authorization           header    string                             true  "Bearer <access_token>"
// @Param        budget_uuid             path      string                             true  "Parent Budget UUID (36 characters)"
// @Param        description_expense_uuid path      string                             true  "Specific Transaction UUID (36 characters)"
// @Param        request                 body      RequestUpdateDescriptionExpense    true  "Expense translation update payload"
// @Success      200      {object}  response.Response{data=ResponseCreateAndUpdateExpense,errors=nil} "Expense transaction successfully updated"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"expense\": \"expense must be a positive decimals and greater than 0\" } } or { \"errors\": { \"category\": \"category mast be a health, sport, supermarket, restaurant, leisure, investments, savings or other\" } } or { \"errors\": { \"description\": \"description cannot be more than 250 characters\" } } or { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } } or { \"errors\": { \"body\": \"<raw_body_error>\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } } or { \"errors\": { \"expense\": \"not found expense\" } } or { \"errors\": { \"expense\": \"not found description expense\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"global\": \"failed to update expense\" } }"
// @Router       /api/v1/description-expense/{budget_uuid}/{description_expense_uuid} [patch]
func (h *HandlerExpense) UpdateExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shmiddleware.KeyContextValue)
		values, ok := ctxValues.(*shmiddleware.ContextValues)
		if !ok {
			h.Logger.Error(sherrors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = sherrors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		body, errBody := requestutil.HandlerRequest[RequestUpdateDescriptionExpense](request.Body, h.Validate)
		if errBody != nil {
			mapError := sherrors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Category":
						values.DataLog.MapLog["category"] = body.Category
						mapError.Map["category"] = ErrIncorrectCategory.Error()
					case err.Field() == "Expense":
						values.DataLog.MapLog["expense"] = body.Expense
						mapError.Map["expense"] = "expense" + apperrors.ErrIncorrectDecimal.Error()
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
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expenseUpdate, errUpdate := h.ServiceExpense.UpdateExpense(request.Context(), body, values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		if errUpdate != nil {
			values.DataLog.Errors = errUpdate.Error()
			switch {
			case errors.Is(errUpdate, apperrors.ErrNotFoundBudget):
				resp.Error["budget"] = errUpdate.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errUpdate, apperrors.ErrNotFoundExpense), errors.Is(errUpdate, ErrNotFoundDescriptionExpense):
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

// GetDescriptionExpense godoc
// @Summary      Get a specific expense transaction details
// @Description  Retrieves the complete data of an individual expense transaction by its UUID under a specific budget. Validates the entity existence chain: parent budget, expense container, and the specific transaction.
// @Tags         expense
// @Accept       json
// @Produce      json
// @Param        Authorization           header    string  true  "Bearer <access_token>"
// @Param        budget_uuid             path      string  true  "Parent Budget UUID (36 characters)"
// @Param        description_expense_uuid path      string  true  "Specific Transaction UUID (36 characters)"
// @Success      200      {object}  response.Response{data=model.DescriptionExpenses,errors=nil} "Expense transaction details successfully retrieved"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } } or { \"errors\": { \"expense\": \"not found expense\" } } or { \"errors\": { \"expense\": \"not found description expense\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } }"
// @Router       /api/v1/description-expense/{budget_uuid}/{description_expense_uuid} [get]
func (h *HandlerExpense) GetDescriptionExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shmiddleware.KeyContextValue)
		values, ok := ctxValues.(*shmiddleware.ContextValues)
		if !ok {
			h.Logger.Error(sherrors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = sherrors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		expense, errGetExpense := h.ServiceExpense.GetDescriptionExpense(request.Context(), values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		if errGetExpense != nil {
			values.DataLog.Errors = errGetExpense.Error()
			switch {
			case errors.Is(errGetExpense, apperrors.ErrNotFoundBudget):
				resp.Error["budget"] = errGetExpense.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errGetExpense, apperrors.ErrNotFoundExpense), errors.Is(errGetExpense, ErrNotFoundDescriptionExpense):
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

// DeleteDescriptionExpense godoc
// @Summary      Delete an individual expense transaction
// @Description  Deletes a specific expense transaction record by its UUID under a parent budget. Recalculates total expense fields inside the database transaction and returns no content upon success.
// @Tags         expense
// @Accept       json
// @Produce      json
// @Param        Authorization           header    string  true  "Bearer <access_token>"
// @Param        budget_uuid             path      string  true  "Parent Budget UUID (36 characters)"
// @Param        description_expense_uuid path      string  true  "Specific Transaction UUID (36 characters)"
// @Success      204      "Expense transaction successfully deleted, no content returned"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } } or { \"errors\": { \"expense\": \"not found expense\" } } or { \"errors\": { \"expense\": \"not found description expense\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"global\": \"failed to delete expense\" } }"
// @Router       /api/v1/description-expense/{budget_uuid}/{description_expense_uuid} [delete]
func (h *HandlerExpense) DeleteDescriptionExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shmiddleware.KeyContextValue)
		values, ok := ctxValues.(*shmiddleware.ContextValues)
		if !ok {
			h.Logger.Error(sherrors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = sherrors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		descriptionExpenseUUID := request.PathValue("description_expense_uuid")
		values.DataLog.MapLog["description_expense_uuid"] = descriptionExpenseUUID
		errDelete := h.ServiceExpense.DeleteDescriptionExpense(request.Context(), values.DataAuth.UserUUID, budgetUUID, descriptionExpenseUUID)
		if errDelete != nil {
			values.DataLog.Errors = errDelete.Error()
			var mapError sherrors.MapError
			if errors.As(errDelete, &mapError) {
				resp.Error = mapError.Map
				switch {
				case mapError.Map["budget"] == apperrors.ErrIncorrectFormatBudgetUUID.Error() && len(mapError.Map) == 1:
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

// ListDescriptionExpense godoc
// @Summary      Get a paginated list of individual expense transactions
// @Description  Retrieves a collection of detailed expense transactions for a specific budget using its UUID. Supports pagination via limit and offset query parameters.
// @Tags         expense
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <access_token>"
// @Param        budget_uuid    path      string  true  "Parent Budget UUID (36 characters)"
// @Param        limit          query     string  false "Maximum number of records to return (positive integer, max 100)"
// @Param        offset         query     string  false "Number of records to skip (positive integer)"
// @Success      200      {object}  response.Response{data=[]model.DescriptionExpenses,errors=nil} "List of expense transactions successfully retrieved"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"limit\": \"the limit must be a positive integer not greater than 100\" } } or { \"errors\": { \"offset\": \"the offset must be a positive integer\" } } or { \"errors\": { \"expense\": \"not found expense\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"expense\": \"not found description expense\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } }"
// @Router       /api/v1/description-expense/{budget_uuid} [get]
func (h *HandlerExpense) ListDescriptionExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shmiddleware.KeyContextValue)
		values, ok := ctxValues.(*shmiddleware.ContextValues)
		if !ok {
			h.Logger.Error(sherrors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = sherrors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		limit := request.URL.Query().Get("limit")
		values.DataLog.MapLog["limit"] = limit
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["offset"] = offset
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expenseList, errList := h.ServiceExpense.ListDescriptionExpense(request.Context(), budgetUUID, limit, offset)
		if errList != nil {
			values.DataLog.Errors = errList.Error()
			var mapError sherrors.MapError
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

// GetExpense godoc
// @Summary      Get aggregated expenses for a specific budget
// @Description  Retrieves the top-level aggregated expenses data (total spent vs budget limits) associated with a specific budget by its UUID.
// @Tags         expense
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <access_token>"
// @Param        budget_uuid    path      string  true  "Budget UUID (36 characters)"
// @Success      200      {object}  response.Response{data=model.Expenses,errors=nil} "Aggregated expenses data successfully retrieved"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"expense\": \"not found expense\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } }"
// @Router       /api/v1/expense/{budget_uuid} [get]
func (h *HandlerExpense) GetExpense() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{
			Error: make(map[string]string),
		}
		ctxValues := request.Context().Value(shmiddleware.KeyContextValue)
		values, ok := ctxValues.(*shmiddleware.ContextValues)
		if !ok {
			h.Logger.Error(sherrors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = sherrors.ErrCriticalServer.Error()
			h.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		budgetUUID := request.PathValue("budget_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		expense, errGetExpense := h.ServiceExpense.GetExpense(request.Context(), budgetUUID)
		if errGetExpense != nil {
			values.DataLog.Errors = errGetExpense.Error()
			if errors.Is(errGetExpense, apperrors.ErrIncorrectFormatBudgetUUID) {
				resp.Error["budget"] = errGetExpense.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			} else if errors.Is(errGetExpense, apperrors.ErrNotFoundExpense) {
				resp.Error["expense"] = errGetExpense.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			}
			return
		}
		resp.Success = true
		resp.Data = expense
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
