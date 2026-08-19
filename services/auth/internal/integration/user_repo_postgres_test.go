package integration

import (
	"app/auth-service/internal/model"
	"app/auth-service/internal/user"
	"context"
	"shared/storage"
	"shared/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	password = "$2a$10$wCq6LLOuoRrZhfTM5W714.YXGhG3TmA33IE4Irs9UJfAdspns6E2i"
)

func TestCreateUser_Success(t *testing.T) {
	if errInitial != nil {
		t.Fatal(errInitial)
	}
	testutil.CleanPostgres(t, suiteContainer.Postgres, []string{"users"})
	userRepo := user.NewRepositoryUser(&storage.Postgres{
		DB: suiteContainer.Postgres,
	}, logger)
	userUUID := uuid.New().String()
	name := "Bob"
	email := "example@gmail.com"
	if errCreateUser := userRepo.CreateUser(context.Background(), &model.Users{
		Name:     name,
		Email:    email,
		Password: password,
		UserUUID: userUUID,
	}); errCreateUser != nil {
		t.Fatal("failed to create user: ", errCreateUser)
	}
	userData, errGetUser := userRepo.GetUserByUUID(context.Background(), userUUID)
	if errGetUser != nil {
		t.Fatal("failed to get user: ", errGetUser)
	}
	if userData.Name != name {
		t.Errorf("expected name %s got %s", name, userData.Name)
	}
	if userData.Email != email {
		t.Errorf("expected email %s got %s", email, userData.Email)
	}
}
func TestUpdateUser_BorderlineCase(t *testing.T) {
	if errInitial != nil {
		t.Fatal(errInitial)
	}
	testutil.CleanPostgres(t, suiteContainer.Postgres, []string{"users"})
	userRepo := user.NewRepositoryUser(&storage.Postgres{
		DB: suiteContainer.Postgres,
	}, logger)
	userUUID := uuid.New().String()

	caseUpdateUser := []struct {
		Name          string
		UpdateUser    *model.Users
		ExpectedError bool
	}{
		{
			Name: "Success - update name only",
			UpdateUser: &model.Users{
				UserUUID: userUUID,
				Name:     "Alice",
				Email:    "example@gmail.com",
			},
			ExpectedError: false,
		},
		{
			Name: "Success - update email to valid format",
			UpdateUser: &model.Users{
				UserUUID: userUUID,
				Name:     "Bob",
				Email:    "new-email@gmail.com",
			},
			ExpectedError: false,
		},
		{
			Name: "Success - update password hash",
			UpdateUser: &model.Users{
				UserUUID: userUUID,
				Name:     "Bob",
				Email:    "example@gmail.com",
				Password: "$2a$10$wCq6LLOuoRrZhfTM5W714.YXGhG3TmA33IE4IrsmnJfAdspns6E2i",
			},
			ExpectedError: false,
		},
		{
			Name: "Success - no fields changed",
			UpdateUser: &model.Users{
				UserUUID: userUUID,
				Name:     "Bob",
				Email:    "example@gmail.com",
			},
			ExpectedError: false,
		},
		{
			Name: "Error - empty user UUID",
			UpdateUser: &model.Users{
				UserUUID: "",
				Name:     "Alice",
			},
			ExpectedError: true,
		},
		{
			Name: "Error - name too long",
			UpdateUser: &model.Users{
				UserUUID: userUUID,
				Name: func() string {
					tooLongName := "aa"
					for i := 0; i < 50; i++ {
						tooLongName += tooLongName
					}
					return tooLongName
				}(),
				Email: "example@gmail.com",
			},
			ExpectedError: true,
		},
		{
			Name: "Error - invalid email format",
			UpdateUser: &model.Users{
				UserUUID: userUUID,
				Name:     "Bob",
				Email:    "invalid-email-format",
			},
			ExpectedError: true,
		},
		{
			Name: "Error - empty name string",
			UpdateUser: &model.Users{
				UserUUID: userUUID,
				Name:     "",
				Email:    "example@gmail.com",
			},
			ExpectedError: true,
		},
	}
	for _, tc := range caseUpdateUser {
		t.Run(tc.Name, func(t *testing.T) {
			testutil.CleanPostgres(t, suiteContainer.Postgres, []string{"users"})
			if errCreateUser := userRepo.CreateUser(context.Background(), &model.Users{
				Name:     "Bob",
				Email:    "example@gmail.com",
				Password: password,
				UserUUID: userUUID,
			}); errCreateUser != nil {
				t.Fatal("failed to create user: ", errCreateUser)
			}
			errUpdate := userRepo.UpdateUser(context.Background(), tc.UpdateUser, userUUID)
			if errUpdate != nil && !tc.ExpectedError {
				t.Fatal("failed to update user: ", errUpdate)
			} else if errUpdate == nil && tc.ExpectedError {
				t.Fatal("expected error update")
			}
			if !tc.ExpectedError {
				resUser, errGetUser := userRepo.GetUserByUUID(context.Background(), userUUID)
				if errGetUser != nil {
					t.Fatal("failed to get update user: ", errGetUser)
				}
				if resUser.Name != tc.UpdateUser.Name {
					t.Errorf("expected name %s got %s", tc.UpdateUser.Name, resUser.Name)
				}
				if resUser.Email != tc.UpdateUser.Email {
					t.Errorf("expected email %s got %s", resUser.Email, tc.UpdateUser.Email)
				}
				if resUser.Password != tc.UpdateUser.Password {
					t.Errorf("expected password %s got %s", resUser.Password, tc.UpdateUser.Password)
				}
			}
		})
	}

}
func TestRemoveUser_Success(t *testing.T) {
	if errInitial != nil {
		t.Fatal(errInitial)
	}
	testutil.CleanPostgres(t, suiteContainer.Postgres, []string{"users"})
	userRepo := user.NewRepositoryUser(&storage.Postgres{
		DB: suiteContainer.Postgres,
	}, logger)
	userUUID := uuid.New().String()
	name := "Bob"
	email := "example@gmail.com"
	if errCreateUser := userRepo.CreateUser(context.Background(), &model.Users{
		Name:     name,
		Email:    email,
		Password: password,
		UserUUID: userUUID,
	}); errCreateUser != nil {
		t.Fatal("failed to create user: ", errCreateUser)
	}
	if errRemoveUser := userRepo.RemoveUser(context.Background(), userUUID); errRemoveUser != nil {
		t.Fatal("failed to remove user: ", errRemoveUser)
	}
	userData, errGetUser := userRepo.GetUserByUUID(context.Background(), userUUID)
	if errGetUser == nil {
		t.Fatal("failed to remove user: ", userData)
	}
	removeUser := &model.Users{}
	if errGetRemove := suiteContainer.
		Postgres.Unscoped().
		Where("user_uuid = ?", userUUID).
		Take(removeUser).Error; errGetRemove != nil {
		t.Fatal("failed to get remove user: ", errGetRemove)
	}
	if removeUser.Name != name {
		t.Errorf("expected name remove %s got %s", name, removeUser.Name)
	}
	if removeUser.Email != email {
		t.Errorf("expected email remove %s got %s", email, removeUser.Email)
	}
}
func TestDeleteUser_Success(t *testing.T) {
	if errInitial != nil {
		t.Fatal(errInitial)
	}
	testutil.CleanPostgres(t, suiteContainer.Postgres, []string{"users"})
	userRepo := user.NewRepositoryUser(&storage.Postgres{
		DB: suiteContainer.Postgres,
	}, logger)
	userUUID := uuid.New().String()
	name := "Bob"
	email := "example@gmail.com"
	if errCreateUser := userRepo.CreateUser(context.Background(), &model.Users{
		Name:     name,
		Email:    email,
		Password: password,
		UserUUID: userUUID,
	}); errCreateUser != nil {
		t.Fatal("failed to create user: ", errCreateUser)
	}
	if errDeleteUser := userRepo.DeleteUser(context.Background(), userUUID); errDeleteUser != nil {
		t.Fatal("failed to delete user: ", errDeleteUser)
	}
	exist, errCheckUser := userRepo.UserExistsByUserUUID(context.Background(), userUUID)
	if errCheckUser == nil {
		t.Fatal("failed to check delete user: ", errCheckUser)
	}
	if exist {
		t.Fatal("failed to delete user")
	}
}

func TestRecoveryUser_Success(t *testing.T) {
	if errInitial != nil {
		t.Fatal(errInitial)
	}
	testutil.CleanPostgres(t, suiteContainer.Postgres, []string{"users"})
	userRepo := user.NewRepositoryUser(&storage.Postgres{
		DB: suiteContainer.Postgres,
	}, logger)
	userUUID := uuid.New().String()
	name := "Bob"
	email := "example@gmail.com"
	if errCreateUser := userRepo.CreateUser(context.Background(), &model.Users{
		Name:     name,
		Email:    email,
		Password: password,
		UserUUID: userUUID,
	}); errCreateUser != nil {
		t.Fatal("failed to create user: ", errCreateUser)
	}
	if errRemoveUser := userRepo.RemoveUser(context.Background(), userUUID); errRemoveUser != nil {
		t.Fatal("failed to remove user: ", errRemoveUser)
	}
	userData, errGetUser := userRepo.GetUserByUUID(context.Background(), userUUID)
	if errGetUser == nil {
		t.Fatal("failed to remove user: ", userData)
	}
	if errRecoveryUser := userRepo.RecoveryUser(context.Background(), userUUID); errRecoveryUser != nil {
		t.Fatal("failed to recovery user: ", errRecoveryUser)
	}
	userData, errGetUserRecovery := userRepo.GetUserByUUID(context.Background(), userUUID)
	if errGetUserRecovery != nil {
		t.Fatal("failed to recovery user: ", errGetUserRecovery)
	}
	if userData.Name != name {
		t.Errorf("expected name %s got %s", name, userData.Name)
	}
	if userData.Email != email {
		t.Errorf("expected email %s got %s", email, userData.Email)
	}
}
func TestDeleteUserByTimer_Success(t *testing.T) {
	if errInitial != nil {
		t.Fatal(errInitial)
	}
	testutil.CleanPostgres(t, suiteContainer.Postgres, []string{"users"})
	userRepo := user.NewRepositoryUser(&storage.Postgres{
		DB: suiteContainer.Postgres,
	}, logger)
	firstUserUUID := uuid.New().String()
	if errCreateFirst := userRepo.CreateUser(context.Background(), &model.Users{
		CreatedAt: time.Now().Add(-time.Hour * 1000),
		UpdatedAt: time.Now().Add(-time.Hour * 1000),
		DeletedAt: gorm.DeletedAt{Time: time.Now().Add(-time.Hour * 726)},
		Name:      "Bob",
		Email:     "exampleemail@gmail.com",
		Password:  password,
		UserUUID:  firstUserUUID,
	}); errCreateFirst != nil {
		t.Fatal("failed to create first user: ", errCreateFirst)
	}
	secondUserUUID := uuid.New().String()
	if errCreateSecond := userRepo.CreateUser(context.Background(), &model.Users{
		CreatedAt: time.Now().Add(-time.Hour * 1000),
		UpdatedAt: time.Now().Add(-time.Hour * 1000),
		DeletedAt: gorm.DeletedAt{Time: time.Now().Add(-time.Hour * 725)},
		Name:      "Jhon",
		Email:     "exampleemailsecond@gmail.com",
		Password:  password,
		UserUUID:  secondUserUUID,
	}); errCreateSecond != nil {
		t.Fatal("failed to create second user: ", errCreateSecond)
	}
	sliceUserUUID, errDeleteExpireUser := userRepo.DeleteUsersByTimer()
	if errDeleteExpireUser != nil {
		t.Fatal("failed to delete users: ", errDeleteExpireUser)
	}
	var returnFirstUUID, returnSecondUUID bool
	for _, u := range sliceUserUUID {
		if u == firstUserUUID {
			returnFirstUUID = true
		}
		if u == secondUserUUID {
			returnSecondUUID = true
		}
	}
	if !returnFirstUUID {
		t.Error("first user uuid doesn't return")
	}
	if !returnSecondUUID {
		t.Error("second user uuid doesn't return")
	}
	existFirst, errCheckExistFirst := userRepo.UserExistsByUserUUID(context.Background(), firstUserUUID)
	if errCheckExistFirst != nil {
		t.Fatal("failed to check exist first user: ", errCheckExistFirst)
	}
	if existFirst {
		t.Error("failed to delete first user")
	}
	existSecond, errCheckExistSecond := userRepo.UserExistsByUserUUID(context.Background(), secondUserUUID)
	if errCheckExistSecond != nil {
		t.Fatal("failed to check exist second user: ", errCheckExistFirst)
	}
	if existSecond {
		t.Error("failed to delete second user")
	}
}
