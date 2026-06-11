package finance


type ServiceFinance struct{
	Repo *RepositoryFinance
}
func NewServiceFinance(repo *RepositoryFinance)*ServiceFinance{
	return &ServiceFinance{
		Repo : repo,
	}
}
