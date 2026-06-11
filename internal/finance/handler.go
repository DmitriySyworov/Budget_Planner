package finance

import (
	"app/budget-planner/internal/loggers"
	"app/budget-planner/internal/middleware"
	"app/budget-planner/internal/response"
	"net/http"
)

type HandlerFinance struct {
	*ServiceFinance
	Resp *response.HandlerResponse
}

func NewHandlerFinance(router *http.ServeMux, service *ServiceFinance, response *response.HandlerResponse, logger *loggers.Logger, mv *middleware.ManagerMiddleware) {
	finance := &HandlerFinance{
		ServiceFinance: service,
		Resp:           response,
	}
	router.Handle("GET /api/v1/finance")
}
