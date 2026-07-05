package expense

import "app/budget-planner/internal/model"

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
type ResponseCreateAndUpdateExpense struct {
	*model.Expenses            `json:"expense"`
	*model.DescriptionExpenses `json:"description_expenses"`
}
