package swaggerdocs

import (
	"net/http"
	"shared/response"
)

type HandlerSwaggerDocs struct {
	Service *ServiceSwaggerDocs
	*response.HandlerResponse
}

func NewHandlerSwaggerDocs(router *http.ServeMux, service *ServiceSwaggerDocs, respHandler *response.HandlerResponse) {
	docs := HandlerSwaggerDocs{
		Service:         service,
		HandlerResponse: respHandler,
	}
	router.HandleFunc("GET /api/v1/docs", docs.GetDocs())
	router.HandleFunc("GET /api/v1/docs/api", docs.GetDocsAPI())
	router.HandleFunc("PUT /api/v1/docs", docs.UpdateDocs())
}

func (h *HandlerSwaggerDocs) GetDocs() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{Error: make(map[string]string)}
		infoService := h.Service.GetDocs()
		resp.Success = true
		resp.Data = infoService
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}

func (h *HandlerSwaggerDocs) GetDocsAPI() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		resp := &response.Response{Error: make(map[string]string)}
		service := request.URL.Query().Get("service")
		dataDocs, errGetDocs := h.Service.GetDocsAPI(service)
		if errGetDocs != nil {
			resp.Error["docs"] = errGetDocs.Error()
			h.ResponseSend(writer, resp, http.StatusNotFound)
			return
		}
		resp.Success = true
		resp.Data = dataDocs
		h.ResponseSend(writer, resp, http.StatusOK)
	}
}

func (h *HandlerSwaggerDocs) UpdateDocs() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		h.Service.UpdateDocs()
		writer.WriteHeader(http.StatusNoContent)
	}
}
