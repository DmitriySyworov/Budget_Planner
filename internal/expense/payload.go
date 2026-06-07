package expense

type CreateAndUpdateExpense struct {
	Health      string `json:"health,omitempty"`
	Sport       string `json:"sport,omitempty"`
	Supermarket string `json:"supermarket,omitempty"`
	Restaurant  string `json:"restaurant,omitempty"`
	Leisure     string `json:"leisure,omitempty"`
	Investments string `json:"investments,omitempty"`
	Savings     string `json:"savings,omitempty"`
	Other       string `json:"other,omitempty"`
}
