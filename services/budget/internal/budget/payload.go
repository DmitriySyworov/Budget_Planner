package budget

type RequestCreateBudget struct {
	Amount      string
	Start       string `validate:"datetime=2006-01-02"`
	Finish      string `validate:"datetime=2006-01-02"`
	Description string
}
type RequestUpdateBudget struct {
	Amount      string
	Start       string `validate:"omitempty,datetime=2006-01-02"`
	Finish      string `validate:"omitempty,datetime=2006-01-02"`
	Description string
}
