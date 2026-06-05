package user

import "net/http"

type HandlerUser struct{
	*ServiceUser
}

func NewHandlerUser(router *http.ServeMux, service *ServiceUser){

}