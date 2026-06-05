package di

type IRepoUser interface {
	IsUserExistsByUUID(userUUID string) bool
}
