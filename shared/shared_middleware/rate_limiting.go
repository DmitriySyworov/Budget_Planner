package shared_middleware

import (
	"net/http"
	"shared/rate_limiter"
	"shared/response"
	"shared/shared_errors"
	"strings"
)

func (m *ManagerSharedMiddleware) RateLimiting(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.Contains(request.URL.Path, "/ready") || strings.Contains(request.URL.Path, "/health") {
			next.ServeHTTP(writer, request)
			return
		}
		resp := &response.Response{Error: make(map[string]string, 1)}
		ctxValues := request.Context().Value(KeyContextValue)
		values, ok := ctxValues.(*ContextValues)
		if !ok {
			m.Logger.Error(shared_errors.ErrFailedAssertionContextValues.Error() + request.Pattern)
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			m.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		counterRequest, headers, errRateLimiting := m.RateLimit.RateLimiter(values.DataAuth.UserUUID)
		if errRateLimiting != nil {
			resp.Error["global"] = shared_errors.ErrCriticalServer.Error()
			m.ResponseSend(writer, resp, http.StatusInternalServerError)
			return
		}
		if counterRequest > rate_limiter.RequestLimit {
			writer.Header().Set("Retry-After", "60")
			for key, value := range headers {
				writer.Header().Set(key, value)
			}
			resp.Error["rate_limit"] = shared_errors.ErrRateLimiting.Error()
			m.ResponseSend(writer, resp, http.StatusTooManyRequests)
			return
		}
		rateLimitWrapperWriter := &RateLimitWrapperWriter{
			ResponseWriter: writer,
			Headers:        headers,
		}
		next.ServeHTTP(rateLimitWrapperWriter, request)
	})
}

type RateLimitWrapperWriter struct {
	http.ResponseWriter
	Headers map[string]string
}

func (w *RateLimitWrapperWriter) WriteHeader(code int) {
	for key, value := range w.Headers {
		w.ResponseWriter.Header().Set(key, value)
	}
	w.ResponseWriter.WriteHeader(code)
}
