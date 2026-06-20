package finance

import (
	"app/budget-planner/internal/custom_errors"
	"shared/response"
	"shared/shared_errors"
	"shared/shared_middleware"

	"net/http"
	"shared/loggers"
)

type HandlerFinance struct {
	*ServiceFinance
	*response.HandlerResponse
	Logger *loggers.Logger
}

func NewHandlerFinance(router *http.ServeMux, service *ServiceFinance, response *response.HandlerResponse, logger *loggers.Logger, mv *shared_middleware.ManagerSharedMiddleware) {
	finance := &HandlerFinance{
		ServiceFinance:  service,
		HandlerResponse: response,
		Logger:          logger,
	}
	router.Handle("GET /api/v1/finance/{budget_uuid}/{expense_uuid}", mv.HandlerAccessToken(finance.Finance()))
}
func (h HandlerFinance) Finance() http.HandlerFunc {
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
		budgetUUID := request.PathValue("budget_uuid")
		expenseUUID := request.PathValue("expense_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		values.DataLog.MapLog["expense_uuid"] = expenseUUID
		finance, errGetFinance := h.ServiceFinance.Finance(values.DataAuth.UserUUID, budgetUUID, expenseUUID)
		resp.Error = append(resp.Error, errGetFinance...)
		if len(resp.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, errGetFinance...)
			if resp.Error[0] == custom_errors.ErrNotFoundUser.Error() || resp.Error[0] == custom_errors.ErrNotFoundBudget.Error() || resp.Error[0] == custom_errors.ErrNotFoundExpense.Error() {
				h.ResponseSend(writer, resp, http.StatusNotFound)
			} else if resp.Error[0] == ErrFailedGetFinance.Error() {
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
		}
		resp.Success = true
		resp.Data = finance
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
