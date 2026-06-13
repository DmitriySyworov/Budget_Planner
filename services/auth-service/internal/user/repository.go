package user

import (
	"app/budget-planner/internal/open_db"
)

type RepositoryUser struct {
	*open_db.Postgres
}

func NewRepositoryUser(postgres *open_db.Postgres) *RepositoryUser {
	return &RepositoryUser{
		Postgres: postgres,
	}
}
func (r *RepositoryUser) IsUserExistsByUUID(userUUID string) bool {
	var isExist bool
	errQuery := r.Postgres.
		Raw(`SELECT EXISTS(
				 SELECT FROM budget
				 WHERE user_uuid = ?)`, userUUID).Scan(&isExist)
	if !isExist || errQuery != nil {
		return false
	}
	return true
}
