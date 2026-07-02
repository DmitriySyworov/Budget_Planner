package expense

type RequestCreateDescriptionExpense struct {
	Category    string `validate:"required,oneof=health sport supermarket restaurant leisure investments savings other"`
	Expense     string `validate:"required"`
	Description string `validate:"omitempty,max=250"`
}
type RequestUpdateDescriptionExpense struct {
	Category    string `validate:"omitempty,oneof=health sport supermarket restaurant leisure investments savings other"`
	Expense     string
	Description string `validate:"omitempty,max=250"`
}
