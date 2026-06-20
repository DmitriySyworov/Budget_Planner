package user

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/di"
	"app/auth-service/internal/model"
	"fmt"
	"shared/shared_common"
	"shared/shared_errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ServiceUser struct {
	Repo *RepositoryUser
	di.IServiceAuth
	di.IRepoAuth
	Conf *authconfig.VerifyEmail
}

func NewServiceUser(repo *RepositoryUser, iServiceAuth di.IServiceAuth, iRepoAuth di.IRepoAuth) *ServiceUser {
	return &ServiceUser{
		Repo:         repo,
		IServiceAuth: iServiceAuth,
		IRepoAuth:    iRepoAuth,
	}
}

func (s *ServiceUser) UpdateUser(userUUID string, body *RequestUpdateUser) (*model.User, *common.ResponseAuth, []string) {
	if body.NewPassword == "" && body.NewEmail == "" {
		user, errGet := s.Repo.GetUserByUUID(userUUID)
		if errGet != nil || user == nil {
			return nil, nil, []string{custom_errors.ErrNotFoundUser.Error()}
		}
		user.Name = body.NewName
		if s.Repo.UpdateUser(user, userUUID) != nil {
			return nil, nil, []string{ErrFailedUpdateUser.Error()}
		}
		return user, nil, nil
	}
	origPassword, errGetUser := s.Repo.GetPasswordByEmail(body.Email)
	if errGetUser != nil {
		return nil, nil, []string{custom_errors.ErrIncorrectPasswordOrEmail.Error()}
	}
	if bcrypt.CompareHashAndPassword([]byte(origPassword), []byte(body.Password)) != nil {
		return nil, nil, []string{custom_errors.ErrIncorrectPasswordOrEmail.Error()}
	}
	var newHashedPassword string
	if body.NewPassword != "" {
		password, errHash := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
		if errHash != nil {
			return nil, nil, []string{custom_errors.ErrFailedSecurity.Error()}
		}
		newHashedPassword = string(password)
	}
	var sendEmail string
	if body.Email != "" && body.NewEmail == "" {
		sendEmail = body.Email
	} else if body.Email == "" && body.NewEmail != "" {
		sendEmail = body.NewEmail
	} else {
		return nil, nil, []string{ErrIncorrectChoiceEmail.Error()}
	}
	const sizeUpdateDataMap = 5
	dataUser := make(map[string]string, sizeUpdateDataMap)
	dataUser[emailKey] = sendEmail
	dataUser[newEmailKey] = body.NewEmail
	dataUser[newNameKey] = body.NewName
	dataUser[newPasswordKey] = newHashedPassword
	respAuth, errAuth := s.HelperAuth(actionUpdate, dataUser)
	if errAuth != nil {
		return nil, nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return nil, respAuth, nil
}
func (s *ServiceUser) GetUser(userUUID string) (*ResponseUser, []string) {
	user, errGet := s.Repo.GetResponseUserByUUID(userUUID)
	if errGet != nil {
		return nil, []string{errGet.Error()}
	}
	return user, nil
}
func (s *ServiceUser) DeleteUser(email, typeRemove string) (*common.ResponseAuth, []string) {
	sliceError := make([]string, 0, 2)
	if typeRemove != shared_common.TypeSoftDelete && typeRemove != shared_common.TypeHardDelete {
		sliceError = append(sliceError, shared_errors.ErrIncorrectTypeRemove.Error())
	}
	if !s.Repo.UserExistsByEmail(email) {
		sliceError = append(sliceError, custom_errors.ErrNotFoundUser.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	const sizeRemoveDataMap = 2
	dataUser := make(map[string]string, sizeRemoveDataMap)
	dataUser[emailKey] = email
	respAuth, errAuth := s.HelperAuth(typeRemove, dataUser)
	if errAuth != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respAuth, nil
}

const (
	actionUpdate = "update"

	emailKey       = "email"
	newEmailKey    = "new_email"
	newNameKey     = "new_name"
	newPasswordKey = "new_password"
)

func (s *ServiceUser) ConfirmUser(codeUser int, userUUID, sessionID, action string) (*ResponseUser, []string) {
	sliceError := make([]string, 0, 2)
	if len(sessionID) != 36 {
		sliceError = append(sliceError, custom_errors.ErrIncorrectSessionID.Error())
	}
	if action != shared_common.TypeHardDelete && action != shared_common.TypeSoftDelete && action != actionUpdate {
		sliceError = append(sliceError, ErrIncorrectAction.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	dataSession, errGetCode := s.IRepoAuth.GetUserSession(sessionID, action)
	if errGetCode != nil {
		return nil, []string{custom_errors.ErrSessionExpired.Error()}
	}
	if fmt.Sprint(codeUser) != dataSession[common.CodeKey] {
		return nil, []string{custom_errors.ErrIncorrectCode.Error()}
	}
	if action == actionUpdate {
		if !s.Repo.UserExistsByUserUUID(userUUID) {
			return nil, []string{custom_errors.ErrNotFoundUser.Error()}
		}
		user := &model.User{
			Name:     dataSession[newNameKey],
			Email:    dataSession[newEmailKey],
			Password: dataSession[newPasswordKey],
		}
		if s.Repo.UpdateUser(user, userUUID) != nil {
			return nil, []string{ErrFailedUpdateUser.Error()}
		}
		return &ResponseUser{
			CreatedAt: user.CreatedAt.Format(time.DateOnly),
			UpdatedAt: user.UpdatedAt.Format(time.DateOnly),
			Name:      user.Name,
			Email:     user.Email,
			UserUUID:  user.UserUUID,
		}, nil
	}
	if action == shared_common.TypeHardDelete {
		if s.Repo.DeleteUser(userUUID) != nil {
			return nil, []string{ErrFailedDeleteUser.Error()}
		}
	} else if action == shared_common.TypeSoftDelete {
		if s.Repo.RemoveUser(userUUID) != nil {
			return nil, []string{ErrFailedRemoveUser.Error()}
		}
	} else {
		return nil, []string{custom_errors.ErrSessionExpired.Error()}
	}
	return nil, nil
}
