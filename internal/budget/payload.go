package budget

type CreateAndUpdateBudget struct {
	Amount      string
	Start       string `validate:"datetime=2006-01-01"`
	Finish      string `validate:"datetime=2006-01-01"`
	Description string
}
