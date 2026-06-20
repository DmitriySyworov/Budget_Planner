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
	DataAuth *DataAuth
	DataLog  *DataLog
}

const (
	KeyContextValue = "keyCtxValue"
	sizeMap         = 8
)

func NewManagerSharedMiddleware(signature string, logger *loggers.Logger, handlerResponse *response.HandlerResponse) *ManagerMiddleware {
	return &ManagerMiddleware{
		Signature:       signature,
		Logger:          logger,
		HandlerResponse: handlerResponse,
		ContextValues: &ContextValues{
			DataAuth: &DataAuth{},
			DataLog: &DataLog{
				Errors: make([]string, 10),
				MapLog: make(map[string]any, sizeMap),
			},
		},
	}
}
