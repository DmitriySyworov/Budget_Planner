package shared_middleware

import (
	"shared/loggers"
	"shared/response"
)

type ManagerSharedMiddleware struct {
	Logger    *loggers.Logger
	Signature string
	*response.HandlerResponse
}

type ContextValues struct {
	DataAuth *DataAuth
	DataLog  *DataLog
}

const (
	KeyContextValue = "keyCtxValue"
)

func NewManagerSharedMiddleware(signature string, logger *loggers.Logger, handlerResponse *response.HandlerResponse) *ManagerSharedMiddleware {
	return &ManagerSharedMiddleware{
		Signature:       signature,
		Logger:          logger,
		HandlerResponse: handlerResponse,
	}
}
