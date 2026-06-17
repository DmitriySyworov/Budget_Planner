package user

type ResponseUser struct {
	CreatedAt string
	UpdatedAt string
	Name      string
	Email     string
	UserUUID  string
}
type RequestRemoveUser struct {
	Email string `validate:"required,email"`
}
type RequestUpdateUser struct {
	NewName     string `validate:"omitempty,min=2,max=64"`
	NewEmail    string `validate:"omitempty,email"`
	NewPassword string `validate:"omitempty,min=8,max=24"`
	Password    string `validate:"omitempty,min=8,max=24,required_with=NewEmail NewPassword"`
	Email       string `validate:"omitempty,email"`
}
