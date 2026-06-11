package finance

import "app/budget-planner/internal/open_db"

type RepositoryFinance struct{
	*open_db.Postgres
}
func NewRepositoryFinance(postgres *open_db.Postgres)*RepositoryFinance{
	return &RepositoryFinance{
		Postgres : postgres,
	}
}
