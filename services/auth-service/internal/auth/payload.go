package auth

type RequestRegister struct {
	Name     string `validate:"required"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=24"`
}
type RequestLogin struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=24"`
}
type ResponseAuth struct {
	Message    string
	JwtSession string `json:"jwt_session"`
}
type ResponseConfirm struct {
	AccessJwt  string `json:"access_jwt"`
	RefreshJwt string `json:"refresh_jwt"`
}
type RequestRefresh struct {
	RefreshJwt string `json:"refresh_jwt" validate:"required"`
}
