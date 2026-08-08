package budget

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

type HandlerBudget struct {
	*ServiceBudget
	*loggers.Logger
	*response.HandlerResponse
	*validator.Validate
}

func NewHandlerBudget(router *http.ServeMux, service *ServiceBudget, logger *loggers.Logger, responseHandler *response.HandlerResponse, validate *validator.Validate, mv *shmiddleware.ManagerSharedMiddleware) {
	budget := &HandlerBudget{
		ServiceBudget:   service,
		Logger:          logger,
		Validate:        validate,
		HandlerResponse: responseHandler,
	}
	router.Handle("POST /api/v1/budget", mv.HandlerAccessToken(budget.CreateBudget()))
	router.Handle("PATCH /api/v1/budget/{uuid}", mv.HandlerAccessToken(budget.UpdateBudget()))
	router.Handle("GET /api/v1/budget/{uuid}", mv.HandlerAccessToken(budget.GetBudget()))
	router.Handle("DELETE /api/v1/budget/{uuid}", mv.HandlerAccessToken(budget.RemoveBudget()))
	router.Handle("GET /api/v1/budget", mv.HandlerAccessToken(budget.ListBudget()))
}

// CreateBudget godoc
// @Summary      Create a new financial budget plan
// @Description  Creates a personal budget allocation with date boundaries. Validates time period overlaps and restricts multiple active budgets for the same time window.
// @Tags         budget
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string              true  "Bearer <access_token>"
// @Param        request        body      RequestCreateBudget true  "Budget creation payload"
// @Success      201      {object}  response.Response{data=model.Budgets,errors=nil} "Budget plan successfully created"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"amount\": \"amount must be a positive decimals and greater than 0\" } } or { \"errors\": { \"start\": \"the start parameter must be in the format YYYY-MM-DD\" } } or { \"errors\": { \"finish\": \"the finish parameter must be in the format YYYY-MM-DD\" } } or { \"errors\": { \"dates\": \"the start parameter cannot be greater than or equal to finish\" } } or { \"errors\": { \"dates\": \"the start and finish parameters are overlap with another budget time period\" } } or { \"errors\": { \"body\": \"<raw_body_error>\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"global\": \"failed to create budget\" } }"
// @Router       /budget [post]
func (h *HandlerBudget) CreateBudget() http.HandlerFunc {
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
		body, errBody := requestutil.HandlerRequest[RequestCreateBudget](request.Body, h.Validate)
		if errBody != nil {
			mapError := sherrors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Amount":
						values.DataLog.MapLog["amount"] = body.Amount
						mapError.Map["amount"] = "amount" + apperrors.ErrIncorrectDecimal.Error()
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
		budgetCreate, errCreate := h.ServiceBudget.CreateBudget(request.Context(), body, values.DataAuth.UserUUID)
		if errCreate != nil {
			values.DataLog.Errors = errCreate.Error()
			var mapError sherrors.MapError
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

// UpdateBudget godoc
// @Summary      Update an existing financial budget plan
// @Description  Updates a budget's amount, dates, or description by its UUID. Validates that new date periods do not overlap with existing budgets and partial date updates are cross-checked with current values.
// @Tags         budget
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string              true  "Bearer <access_token>"
// @Param        uuid           path      string              true  "Budget UUID (36 characters)"
// @Param        request        body      RequestUpdateBudget true  "Budget partial update payload"
// @Success      200      {object}  response.Response{data=model.Budgets,errors=nil} "Budget plan successfully updated"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"amount\": \"amount must be a positive decimals and greater than 0\" } } or { \"errors\": { \"start\": \"the start parameter must be in the format YYYY-MM-DD\" } } or { \"errors\": { \"finish\": \"the finish parameter must be in the format YYYY-MM-DD\" } } or { \"errors\": { \"dates\": \"the start parameter cannot be greater than or equal to finish\" } } or { \"errors\": { \"dates\": \"the start and finish parameters are overlap with another budget time period\" } } or { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } } or { \"errors\": { \"body\": \"<raw_body_error>\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"global\": \"failed to update budget\" } }"
// @Router       /budget/{uuid} [patch]
func (h *HandlerBudget) UpdateBudget() http.HandlerFunc {
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
		body, errBody := requestutil.HandlerRequest[RequestUpdateBudget](request.Body, h.Validate)
		if errBody != nil {
			mapError := sherrors.MapError{Map: make(map[string]string, 3)}
			if errValidate, isErrValid := errBody.(validator.ValidationErrors); isErrValid {
				for _, err := range errValidate {
					switch {
					case err.Field() == "Amount":
						values.DataLog.MapLog["amount"] = body.Amount
						mapError.Map["amount"] = "amount" + apperrors.ErrIncorrectDecimal.Error()
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
		budgetUpdate, errUpdate := h.ServiceBudget.UpdateBudget(request.Context(), body, values.DataAuth.UserUUID, budgetUUID)
		if errUpdate != nil {
			values.DataLog.Errors = errUpdate.Error()
			var mapError sherrors.MapError
			if errors.As(errUpdate, &mapError) {
				resp.Error = mapError.Map
				switch {
				case mapError.Map["budget"] == apperrors.ErrNotFoundBudget.Error() && len(mapError.Map) == 1:
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

// GetBudget godoc
// @Summary      Get budget plan details by UUID
// @Description  Retrieves the complete data of a specific financial budget plan using its unique identifier. Validates resource ownership based on the extracted user token.
// @Tags         budget
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <access_token>"
// @Param        uuid           path      string  true  "Budget UUID (36 characters)"
// @Success      200      {object}  response.Response{data=model.Budgets,errors=nil} "Budget plan details successfully retrieved"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } }"
// @Router       /budget/{uuid} [get]
func (h *HandlerBudget) GetBudget() http.HandlerFunc {
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
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		budget, errGetBudget := h.ServiceBudget.GetBudget(request.Context(), values.DataAuth.UserUUID, budgetUUID)
		if errGetBudget != nil {
			values.DataLog.Errors = errGetBudget.Error()
			resp.Error["budget"] = errGetBudget.Error()
			switch {
			case errors.Is(errGetBudget, apperrors.ErrNotFoundBudget):
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

// RemoveBudget godoc
// @Summary      Remove or hard delete a financial budget plan
// @Description  Removes a budget by its UUID. Supports soft-delete (default) or hard-delete via query parameter. Validates budget existence and user ownership before execution.
// @Tags         budget
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <access_token>"
// @Param        uuid           path      string  true  "Budget UUID (36 characters)"
// @Param        type           query     string  false "Deletion type strategy: 'soft-delete' or 'hard-delete'. Defaults to soft-delete if empty."
// @Success      204      "Budget successfully removed, no content returned"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"budget\": \"the budget uuid must be exactly 36 characters\" } } or { \"errors\": { \"type\": \"the type  must be a soft-delete or hard-delete\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"global\": \"failed to remove budget\" } } or { \"errors\": { \"global\": \"failed to delete budget\" } }"
// @Router       /budget/{uuid} [delete]
func (h *HandlerBudget) RemoveBudget() http.HandlerFunc {
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
		budgetUUID := request.PathValue("uuid")
		values.DataLog.MapLog["budget_uuid"] = budgetUUID
		typeRemove := request.URL.Query().Get("type")
		values.DataLog.MapLog["type"] = typeRemove
		errRemoveBudget := h.ServiceBudget.RemoveBudget(request.Context(), values.DataAuth.UserUUID, budgetUUID, typeRemove)
		if errRemoveBudget != nil {
			values.DataLog.Errors = errRemoveBudget.Error()
			var mapError sherrors.MapError
			if errors.As(errRemoveBudget, &mapError) {
				resp.Error = mapError.Map
				switch {
				case mapError.Map["budget"] == apperrors.ErrNotFoundBudget.Error() && len(mapError.Map) == 1:
					h.ResponseSend(writer, resp, http.StatusNotFound)
				default:
					h.ResponseSend(writer, resp, http.StatusBadRequest)
				}
				return
			}
			switch {
			case errors.Is(errRemoveBudget, sherrors.ErrIncorrectTypeRemove):
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

// ListBudget godoc
// @Summary      Get a paginated list of budget plans
// @Description  Retrieves a collection of budget plans for the authenticated user. Supports pagination via limit and offset query parameters.
// @Tags         budget
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer <access_token>"
// @Param        limit          query     string  false "Maximum number of records to return (positive integer, max 100)"
// @Param        offset         query     string  false "Number of records to skip (positive integer)"
// @Success      200      {object}  response.Response{data=[]model.Budgets,errors=nil} "List of budget plans successfully retrieved"
// @Failure      400      {object}  response.NegativeResponse "Validation or business logic errors. Format: { \"errors\": { \"limit\": \"the limit must be a positive integer not greater than 100\" } } or { \"errors\": { \"offset\": \"the offset must be a positive integer\" } }"
// @Failure      401      {object}  response.NegativeResponse "Authentication errors. Format: { \"errors\": { \"auth\": \"invalid access token\" } } or { \"errors\": { \"auth\": \"access token has expired\" } }"
// @Failure      404      {object}  response.NegativeResponse "Data errors. Format: { \"errors\": { \"budget\": \"not found budget\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } }"
// @Router       /budget [get]
func (h *HandlerBudget) ListBudget() http.HandlerFunc {
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
		offset := request.URL.Query().Get("offset")
		values.DataLog.MapLog["limit"] = limit
		values.DataLog.MapLog["offset"] = offset
		budgetList, errList := h.ServiceBudget.ListBudget(request.Context(), values.DataAuth.UserUUID, limit, offset)
		if errList != nil {
			values.DataLog.Errors = errList.Error()
			var mapError sherrors.MapError
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
