package expense

type CreateAndUpdateExpense struct {
	Category    string `validate:"oneof=health sport supermarket restaurant leisure investments savings other"`
	Expense     string
	Description string `validate:"max=250"`
}
