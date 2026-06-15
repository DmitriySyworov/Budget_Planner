package shared_middleware

import (
	"shared/loggers"
	"shared/response"
)

type ManagerMiddleware struct {
	Logger    *loggers.Logger
	Signature string
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

func NewManagerMiddleware(signature string, logger *loggers.Logger, handlerResponse *response.HandlerResponse) *ManagerMiddleware {
	return &ManagerMiddleware{
		Signature:       signature,
		Logger:          logger,
		HandlerResponse: handlerResponse,
		ContextValues: &ContextValues{
			DataAuth: &DataAuth{},
			DataLog: &DataLog{
				MapLog: make(map[string]any, 1),
			},
		},
	}
}
