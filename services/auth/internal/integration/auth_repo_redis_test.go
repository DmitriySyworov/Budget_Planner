package integration

import "testing"

func TestCreateUserSessionSuccess(t *testing.T) {
	if errInitial != nil {
		t.Fatal(errInitial)
	}
	repoUserRedis := user.NewRepositoryRedisUser(&storage.Redis{Client: suiteContainer.Redis}, logger)
	defer testutil.CleanRedis(suiteContainer.Redis)
}
