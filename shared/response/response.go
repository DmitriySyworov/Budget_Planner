package response

import (
	"encoding/json"
	"shared/loggers"

	"net/http"
)

type Response struct {
	Success bool
	Data    any      `json:"data,omitempty"`
	Error   []string `json:"errors,omitempty"`
}
type HandlerResponse struct {
	*Response
	*loggers.Logger
}

func NewHandlerResponse(logger *loggers.Logger) *HandlerResponse {
	return &HandlerResponse{
		Logger: logger,
		Response: &Response{
			Error: make([]string, 0, 10),
		},
	}
}

func (hr *HandlerResponse) ResponseSend(writer http.ResponseWriter, status int) {
	defer func() {
		hr.Response = &Response{
			Error: make([]string, 0, 10),
		}
	}()
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	errEncode := json.NewEncoder(writer).Encode(hr.Response)
	if errEncode != nil {
		hr.Logger.Error("failed to process the response: ", errEncode)
	}
}
