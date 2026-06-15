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
type RequestConfirm struct {
	Code int `validate:"required,min=100000,max=999999"`
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
