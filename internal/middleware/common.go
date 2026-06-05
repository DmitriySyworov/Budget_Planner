package middleware

import (
	"app/budget-planner/internal/loggers"
	"app/budget-planner/internal/response"
)

type ManagerMiddleware struct {
	Logger *loggers.Logger
	response.Response
	*response.HandlerResponse
	*ContextValues
}

type ContextValues struct {
	*DataAuth
	*DataLog
}

var (
	KeyContextValue = "keyCtxValue"
)

func NewManagerMiddleware(logger *loggers.Logger, handlerResponse *response.HandlerResponse) *ManagerMiddleware {
	return &ManagerMiddleware{
		Logger: logger,
		ContextValues: &ContextValues{
			DataAuth: &DataAuth{},
			DataLog: &DataLog{
				MapLog: make(map[string]any, 1),
			},
		},
	}
}
