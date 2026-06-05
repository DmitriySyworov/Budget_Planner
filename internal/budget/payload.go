package budget

type CreateAndUpdateBudget struct {
	Amount      float64 `validate:"min=1,max=999999999999999"`
	Start       string  `validate:"datetime=2006-01-01"`
	Finish      string  `validate:"datetime=2006-01-01"`
	Description string
}
