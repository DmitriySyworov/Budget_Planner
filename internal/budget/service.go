package budget

type ServiceBudget struct {
	*RepositoryBudget
}

func NewServiceBudget(repo *RepositoryBudget) *ServiceBudget {
	return &ServiceBudget{
		RepositoryBudget: repo,
	}
}
