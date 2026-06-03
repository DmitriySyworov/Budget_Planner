package budget

import "net/http"

type HandlerBudget struct {
	*ServiceBudget
}

func NewHandlerBudget(router *http.ServeMux, service *ServiceBudget) {
	budget := &HandlerBudget{
		ServiceBudget: service,
	}
	router.Handle("POST /budget", budget.CreateBudget())
	router.Handle("PATCH /budget/{hash}", budget.CreateBudget())
	router.Handle("GET /budget/{hash}", budget.CreateBudget())
	router.Handle("DELETE /budget/{hash}", budget.CreateBudget())
	router.Handle("GET /budget/{period}", budget.CreateBudget())
}
func (h *HandlerBudget) CreateBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}

func (h *HandlerBudget) UpdateBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}
func (h *HandlerBudget) GetBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}
func (h *HandlerBudget) DeleteBudget() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}
