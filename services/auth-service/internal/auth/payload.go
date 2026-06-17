package auth

type RequestRegister struct {
	Name     string `validate:"required,min=2,max=64"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=24"`
}
type RequestLogin struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=24"`
}
type ResponseConfirm struct {
	AccessJwt  string `json:"access_jwt"`
	RefreshJwt string `json:"refresh_jwt"`
}
type RequestRefresh struct {
	RefreshJwt string `json:"refresh_jwt" validate:"required"`
}
