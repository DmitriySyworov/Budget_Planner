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

func (s *ServiceUser) UpdateUser(userUUID string, body *RequestUpdateUser) (*model.Users, *common.ResponseAuth, error) {
	if body.NewPassword == "" && body.NewEmail == "" {
		user, errGet := s.Repo.GetUserByUUID(userUUID)
		if errGet != nil || user == nil {
			return nil, nil, custom_errors.ErrNotFoundUser
		}
		user.Name = body.NewName
		if s.Repo.UpdateUser(user, userUUID) != nil {
			return nil, nil, ErrFailedUpdateUser
		}
		return user, nil, nil
	}
	origPassword, errGetUser := s.Repo.GetPasswordByEmail(body.Email)
	if errGetUser != nil {
		return nil, nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	if bcrypt.CompareHashAndPassword([]byte(origPassword), []byte(body.Password)) != nil {
		return nil, nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	var newHashedPassword string
	if body.NewPassword != "" {
		password, errHash := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
		if errHash != nil {
			return nil, nil, custom_errors.ErrFailedSecurity
		}
		newHashedPassword = string(password)
	}
	var sendEmail string
	if body.Email != "" && body.NewEmail == "" {
		sendEmail = body.Email
	} else if body.Email == "" && body.NewEmail != "" {
		sendEmail = body.NewEmail
	} else {
		return nil, nil, ErrIncorrectChoiceEmail
	}
	const sizeUpdateDataMap = 5
	dataUser := make(map[string]string, sizeUpdateDataMap)
	dataUser[emailKey] = sendEmail
	dataUser[newEmailKey] = body.NewEmail
	dataUser[newNameKey] = body.NewName
	dataUser[newPasswordKey] = newHashedPassword
	respAuth, errAuth := s.HelperAuth(actionUpdate, dataUser)
	if errAuth != nil {
		return nil, nil, custom_errors.ErrFailedSecurity
	}
	return nil, respAuth, nil
}
func (s *ServiceUser) GetUser(userUUID string) (*ResponseUser, error) {
	user, errGet := s.Repo.GetResponseUserByUUID(userUUID)
	if errGet != nil {
		return nil, errGet
	}
	return user, nil
}
func (s *ServiceUser) DeleteUser(email, typeRemove string) (*common.ResponseAuth, error) {
	mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
	if typeRemove != shared_common.TypeSoftDelete && typeRemove != shared_common.TypeHardDelete {
		mapError.Map["type"] = shared_errors.ErrIncorrectTypeRemove.Error()
	}
	if !s.Repo.UserExistsByEmail(email) {
		mapError.Map["user"] = custom_errors.ErrNotFoundUser.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	const sizeRemoveDataMap = 2
	dataUser := make(map[string]string, sizeRemoveDataMap)
	dataUser[emailKey] = email
	respAuth, errAuth := s.HelperAuth(typeRemove, dataUser)
	if errAuth != nil {
		return nil, custom_errors.ErrFailedSecurity
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

func (s *ServiceUser) ConfirmUser(codeUser int, userUUID, sessionID, action string) (*ResponseUser, error) {
	mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
	if len(sessionID) != 36 {
		mapError.Map["auth"] = custom_errors.ErrIncorrectSessionID.Error()
	}
	if action != shared_common.TypeHardDelete && action != shared_common.TypeSoftDelete && action != actionUpdate {
		mapError.Map["action"] = ErrIncorrectAction.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	dataSession, errGetCode := s.IRepoAuth.GetUserSession(sessionID, action)
	if errGetCode != nil {
		return nil, custom_errors.ErrSessionExpired
	}
	if fmt.Sprint(codeUser) != dataSession[common.CodeKey] {
		return nil, custom_errors.ErrIncorrectCode
	}
	if action == actionUpdate {
		if !s.Repo.UserExistsByUserUUID(userUUID) {
			return nil, custom_errors.ErrNotFoundUser
		}
		user := &model.Users{
			Name:     dataSession[newNameKey],
			Email:    dataSession[newEmailKey],
			Password: dataSession[newPasswordKey],
		}
		if s.Repo.UpdateUser(user, userUUID) != nil {
			return nil, ErrFailedUpdateUser
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
			return nil, ErrFailedDeleteUser
		}
	} else if action == shared_common.TypeSoftDelete {
		if s.Repo.RemoveUser(userUUID) != nil {
			return nil, ErrFailedRemoveUser
		}
	} else {
		return nil, custom_errors.ErrSessionExpired
	}
	return nil, nil
}
