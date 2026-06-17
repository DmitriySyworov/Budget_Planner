package user

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/di"
	"app/auth-service/internal/model"
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
	respAuth, sessionId, errAuth := s.HelperAuth(sendEmail, s.Conf)
	if errAuth != nil {
		return nil, nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}

	if s.Repo.CreateUpdateDataUserSession(body.NewName, body.NewEmail, newHashedPassword, sessionId) != nil {
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
	sliceError := make([]string, 2)
	if typeRemove != shared_common.TypeSoftDelete && typeRemove != shared_common.TypeHardDelete && typeRemove != "" {
		sliceError = append(sliceError, shared_errors.ErrIncorrectTypeRemove.Error())
	}
	if !s.Repo.UserExistsByEmail(email) {
		sliceError = append(sliceError, custom_errors.ErrNotFoundUser.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	respAuth, sessionId, errAuth := s.HelperAuth(email, s.Conf)
	if errAuth != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	if s.Repo.CreateRemoveDataUserSession(typeRemove, sessionId) != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respAuth, nil
}

const (
	actionUpdate = "update"
)

func (s *ServiceUser) ConfirmUser(codeUser int, userUUID, sessionID, action string) (*ResponseUser, []string) {
	sliceError := make([]string, 2)
	if len(sessionID) != 36 {
		sliceError = append(sliceError, custom_errors.ErrIncorrectSessionID.Error())
	}
	if action != shared_common.TypeHardDelete && action != shared_common.TypeSoftDelete && action != actionUpdate {
		sliceError = append(sliceError, ErrIncorrectAction.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	code, errGetCode := s.IRepoAuth.GetSession(sessionID)
	if errGetCode != nil {
		return nil, []string{custom_errors.ErrSessionExpired.Error()}
	}
	if codeUser != code {
		return nil, []string{custom_errors.ErrIncorrectCode.Error()}
	}
	if action == actionUpdate {
		if !s.Repo.UserExistsByUserUUID(userUUID) {
			return nil, []string{custom_errors.ErrNotFoundUser.Error()}
		}
		dtoUpdateData, errGetSessionUpdate := s.Repo.GetUpdateDataUserSession(sessionID)
		if errGetSessionUpdate != nil {
			return nil, []string{custom_errors.ErrSessionExpired.Error()}
		}
		user := &model.User{
			Name:     dtoUpdateData.NewName,
			Email:    dtoUpdateData.NewEmail,
			Password: dtoUpdateData.NewPassword,
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
	typeRemove, errGetSession := s.Repo.GetRemoveDataUserSession(sessionID)
	if errGetSession != nil {
		return nil, []string{custom_errors.ErrSessionExpired.Error()}
	}
	if typeRemove ==  shared_common.TypeHardDelete {
		if s.Repo.DeleteUser(userUUID) != nil {
			return nil, []string{ErrFailedDeleteUser.Error()}
		}
	}else if typeRemove == shared_common.TypeSoftDelete {
			 if s.Repo.RemoveUser(userUUID) != nil {
				return nil, []string{ErrFailedRemoveUser.Error()}
		}
	} else {
		return nil, []string{custom_errors.ErrSessionExpired.Error()}
	}
	return nil, nil
}
