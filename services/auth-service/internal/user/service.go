package user

type ServiceUser struct {
	Repo *RepositoryUser
}

func NewServiceUser(repo *RepositoryUser) *ServiceUser {
	return &ServiceUser{
		Repo: repo,
	}
}
