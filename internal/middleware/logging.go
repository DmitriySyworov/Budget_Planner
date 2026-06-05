package middleware

import "net/http"

type DataLog struct {
	UserUUID string
	MapLog   map[string]any
	Errors   []string
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		wrapperWrite := &WrapperWriter{
			Status:         http.StatusOK,
			ResponseWriter: writer,
		}
		next.ServeHTTP(wrapperWrite)

	})
}

type WrapperWriter struct {
	http.ResponseWriter
	Status int
}

func (w *WrapperWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}
