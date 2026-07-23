package shared_middleware

import (
	"shared/loggers"
	"shared/rate_limiter"
	"shared/response"
)

type ManagerSharedMiddleware struct {
	Logger    *loggers.Logger
	Signature string
	*response.HandlerResponse
	RateLimit *rate_limiter.Limiter
}

type ContextValues struct {
	DataAuth *DataAuth
	DataLog  *DataLog
}

const (
	KeyContextValue = "keyCtxValue"
)

func NewManagerSharedMiddleware(signature string, rateLimiter *rate_limiter.Limiter, logger *loggers.Logger, handlerResponse *response.HandlerResponse) *ManagerSharedMiddleware {
	return &ManagerSharedMiddleware{
		Signature:       signature,
		Logger:          logger,
		HandlerResponse: handlerResponse,
		RateLimit:       rateLimiter,
	}
}
