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
