package finance

import (
	"app/budget-planner/internal/custom_errors"
	"app/budget-planner/internal/middleware"
	"app/budget-planner/internal/response"
	"net/http"
	"shared/loggers"
)

type HandlerFinance struct {
	*ServiceFinance
	*response.HandlerResponse
	Logger *loggers.Logger
}

func NewHandlerFinance(router *http.ServeMux, service *ServiceFinance, response *response.HandlerResponse, logger *loggers.Logger, mv *middleware.ManagerMiddleware) {
	finance := &HandlerFinance{
		ServiceFinance:  service,
		HandlerResponse: response,
		Logger:          logger,
	}
	router.Handle("GET /api/v1/finance/{budget_uuid}/{expense_uuid}", mv.HandlerAuthToken(finance.Finance()))
}
func (h HandlerFinance) Finance() http.HandlerFunc {
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
		budgetUUID := request.PathValue("budget_uuid")
		expenseUUID := request.PathValue("expense_uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		values.DataLog.MapLog["expense_uuid"] = expenseUUID
		finance, errGetFinance := h.ServiceFinance.Finance(values.DataAuth.UserUUID, budgetUUID, expenseUUID)
		h.Response.Error = append(h.Response.Error, errGetFinance...)
		if len(h.Response.Error) != 0 {
			values.DataLog.Errors = append(values.DataLog.Errors, errGetFinance...)
			if h.Response.Error[0] == custom_errors.ErrNotFoundUser.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundBudget.Error() || h.Response.Error[0] == custom_errors.ErrNotFoundExpense.Error() {
				h.ResponseSend(writer, http.StatusNotFound)
			} else if h.Response.Error[0] == ErrFailedGetFinance.Error() {
				h.ResponseSend(writer, http.StatusInternalServerError)
			} else {
				h.ResponseSend(writer, http.StatusBadRequest)
			}
		}
		h.Response.Success = true
		h.Response.Data = finance
		h.ResponseSend(writer, http.StatusOK)
	}
}
