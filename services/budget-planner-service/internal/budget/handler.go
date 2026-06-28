package budget

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

type HandlerBudget struct {
	*ServiceBudget
	*loggers.Logger
	*response.HandlerResponse
}

func NewHandlerBudget(router *http.ServeMux, service *ServiceBudget, logger *loggers.Logger, responseHandler *response.HandlerResponse, mv *shared_middleware.ManagerSharedMiddleware) {
	budget := &HandlerBudget{
		ServiceBudget:   service,
		Logger:          logger,
		HandlerResponse: responseHandler,
	}
	router.Handle("POST /api/v1/budget", mv.HandlerAccessToken(budget.CreateBudget()))
	router.Handle("PATCH /api/v1/budget/{uuid}", mv.HandlerAccessToken(budget.UpdateBudget()))
	router.Handle("GET /api/v1/budget/{uuid}", mv.HandlerAccessToken(budget.GetBudget()))
	router.Handle("DELETE /api/v1/budget/{uuid}", mv.HandlerAccessToken(budget.RemoveBudget()))
	router.Handle("GET /api/v1/budget", mv.HandlerAccessToken(budget.ListBudget()))
}
func (h *HandlerBudget) CreateBudget() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[RequestCreateBudget](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Amount":
						values.DataLog.MapLog["amount"] = body.Amount
						mapError.Map["amount"] = "amount" + custom_errors.ErrIncorrectDecimal.Error()
					case err.Field() == "Start":
						values.DataLog.MapLog["start"] = body.Start
						mapError.Map["start"] = ErrIncorrectStart.Error()
					case err.Field() == "Finish":
						values.DataLog.MapLog["finish"] = body.Finish
						mapError.Map["finish"] = ErrIncorrectFinish.Error()
					}
				}
			} else {
				mapError.Map["body"] = errBody.Error()
			}
			values.DataLog.Errors = mapError.Error()
			resp.Error = mapError.Map
			h.HandlerResponse.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		budgetCreate, errCreate := h.ServiceBudget.CreateBudget(body, values.DataAuth.UserUUID)
		if errCreate != nil {
			values.DataLog.Errors = errCreate.Error()
			var mapError shared_errors.MapError
			if errors.As(errCreate, &mapError) {
				resp.Error = mapError.Map
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			} else {
				resp.Error["global"] = errCreate.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = budgetCreate
		h.ResponseSend(writer, resp, http.StatusCreated)
	}
}

func (h *HandlerBudget) UpdateBudget() http.HandlerFunc {
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
		body, errBody := handler_request.HandlerRequest[RequestUpdateBudget](request.Body)
		if errBody != nil {
			mapError := shared_errors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Amount":
						values.DataLog.MapLog["amount"] = body.Amount
						mapError.Map["amount"] = "amount" + custom_errors.ErrIncorrectDecimal.Error()
					case err.Field() == "Start":
						values.DataLog.MapLog["start"] = body.Start
						mapError.Map["start"] = ErrIncorrectStart.Error()
					case err.Field() == "Finish":
						values.DataLog.MapLog["finish"] = body.Finish
						mapError.Map["finish"] = ErrIncorrectFinish.Error()
					}
				}
			} else {
				mapError.Map["body"] = errBody.Error()
			}
			values.DataLog.Errors = mapError.Error()
			resp.Error = mapError.Map
			h.HandlerResponse.ResponseSend(writer, resp, http.StatusBadRequest)
			return
		}
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		budgetUpdate, errUpdate := h.ServiceBudget.UpdateBudget(body, values.DataAuth.UserUUID, budgetUUID)
		if errUpdate != nil {
			values.DataLog.Errors = errUpdate.Error()
			var mapError shared_errors.MapError
			if errors.As(errUpdate, &mapError) {
				resp.Error = mapError.Map
				switch {
				case mapError.Map["budget"] == custom_errors.ErrNotFoundBudget.Error() && len(mapError.Map) == 1:
					h.ResponseSend(writer, resp, http.StatusNotFound)
				default:
					h.ResponseSend(writer, resp, http.StatusBadRequest)
				}
			} else {
				resp.Error["global"] = errUpdate.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		resp.Success = true
		resp.Data = budgetUpdate
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerBudget) GetBudget() http.HandlerFunc {
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
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		budget, errGetBudget := h.ServiceBudget.GetBudget(values.DataAuth.UserUUID, budgetUUID)
		if errGetBudget != nil {
			values.DataLog.Errors = errGetBudget.Error()
			resp.Error["budget"] = errGetBudget.Error()
			switch {
			case errors.Is(errGetBudget, custom_errors.ErrNotFoundBudget):
				h.ResponseSend(writer, resp, http.StatusNotFound)
			default:
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			}
			return
		}
		resp.Success = true
		resp.Data = budget
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
func (h *HandlerBudget) RemoveBudget() http.HandlerFunc {
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
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		typeRemove := request.URL.Query().Get("type")
		values.DataLog.MapLog["type"] = typeRemove
		errRemoveBudget := h.ServiceBudget.RemoveBudget(values.DataAuth.UserUUID, budgetUUID, typeRemove)
		if errRemoveBudget != nil {
			values.DataLog.Errors = errRemoveBudget.Error()
			var mapError shared_errors.MapError
			if errors.As(errRemoveBudget, &mapError) {
				resp.Error = mapError.Map
				switch {
				case mapError.Map["budget"] == custom_errors.ErrNotFoundBudget.Error() && len(mapError.Map) == 1:
					h.ResponseSend(writer, resp, http.StatusNotFound)
				default:
					h.ResponseSend(writer, resp, http.StatusBadRequest)
				}
				return
			}
			switch {
			case errors.Is(errRemoveBudget, shared_errors.ErrIncorrectTypeRemove):
				resp.Error["type"] = errRemoveBudget.Error()
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			default:
				resp.Error["global"] = errRemoveBudget.Error()
				h.ResponseSend(writer, resp, http.StatusInternalServerError)
			}
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}
func (h *HandlerBudget) ListBudget() http.HandlerFunc {
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
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["limit"] = limit
		values.DataLog.MapLog["offset"] = offset
		budgetList, errList := h.ServiceBudget.ListBudget(values.DataAuth.UserUUID, limit, offset)
		if errList != nil {
			values.DataLog.Errors = errList.Error()
			var mapError shared_errors.MapError
			if errors.As(errList, &mapError) {
				resp.Error = mapError.Map
				h.ResponseSend(writer, resp, http.StatusBadRequest)
			} else {
				resp.Error["budget"] = errList.Error()
				h.ResponseSend(writer, resp, http.StatusNotFound)
			}
			return
		}
		resp.Success = true
		resp.Data = budgetList
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}
