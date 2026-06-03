package budget

import "app/budget-planner/internal/open_db"

type RepositoryBudget struct {
	*open_db.OpenDB
}

func NewRepositoryBudget(db *open_db.OpenDB) *RepositoryBudget {
	return &RepositoryBudget{
		OpenDB: db,
	}
}
