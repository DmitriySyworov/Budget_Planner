package finance

import (
	"app/budget-planner/internal/apperrors"
	"errors"
	"net/http"
	"shared/loggers"
	"shared/response"
	"shared/sherrors"
	"shared/shmiddleware"
)

type HandlerFinance struct {
	*ServiceFinance
	*response.HandlerResponse
	Logger *loggers.Logger
}

func NewHandlerFinance(router *http.ServeMux, service *ServiceFinance, response *response.HandlerResponse, logger *loggers.Logger, mv *shmiddleware.ManagerSharedMiddleware) {
	finance := &HandlerFinance{
		ServiceFinance:  service,
		HandlerResponse: response,
		Logger:          logger,
	}
	router.Handle("GET /api/v1/finance/{budget_uuid}/{expense_uuid}", mv.HandlerAccessToken(finance.Finance()))
}

// Finance godoc
// @Summary      Get comprehensive financial analytics and predictive calculations
// @Description  Calculates advanced financial analytics for a specific budget and expense container. Returns initial vs current balances, category-by-category expense tracking, percentage distribution, and a predicted average spend per day based on historical user data.
// @Tags         finance
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <access_token>"
// @Param        budget_uuid    path      string  true  "Budget UUID (36 characters)"
// @Param        expense_uuid   path      string  true  "Expense Container UUID (36 characters)"
// @Success      200      {object}  response.Response{data=Finance,errors=nil} "Comprehensive financial data successfully computed and retrieved"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } } or { \"errors\": { \"expense\": \"the expense uuid must be exactly 36 characters\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } } or { \"errors\": { \"expense\": \"not found expense\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"finance\": \"failed to get finance\" } }"
// @Router       /api/v1/finance/{budget_uuid}/{expense_uuid} [get]
func (h HandlerFinance) Finance() http.HandlerFunc {
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
		expenseUUID := request.PathValue("expense_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		values.DataLog.MapLog["expense_uuid"] = expenseUUID
		finance, errGetFinance := h.ServiceFinance.Finance(request.Context(), values.DataAuth.UserUUID, budgetUUID, expenseUUID)
		if errGetFinance != nil {
			values.DataLog.Errors = errGetFinance.Error()
			var mapError sherrors.MapError
			if errors.As(errGetFinance, &mapError) {
				resp.Error = mapError.Map
				h.ResponseSend(writer, resp, http.StatusBadRequest)
				return
			}
			switch {
			case errors.Is(errGetFinance, apperrors.ErrNotFoundBudget):
				resp.Error["budget"] = errGetFinance.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errGetFinance, apperrors.ErrNotFoundExpense):
				resp.Error["expense"] = errGetFinance.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			case errors.Is(errGetFinance, ErrFailedGetFinance):
				resp.Error["finance"] = errGetFinance.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = finance
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
