package user

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/di"
	"app/auth-service/internal/model"
	"app/auth-service/internal/validate_password"
	"context"
	"errors"
	"shared/loggers"
	"shared/shared_constant"
	"shared/shared_errors"
	"shared/shared_kafka"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ServiceUser struct {
	Repo IRepositoryUser
	di.IServiceAuth
	di.IRepoAuth
	Producer *shared_kafka.KafkaProducer
	Logger   *loggers.Logger
}

func NewServiceUser(repo IRepositoryUser, iServiceAuth di.IServiceAuth, iRepoAuth di.IRepoAuth, producer *shared_kafka.KafkaProducer, logger *loggers.Logger) *ServiceUser {
	return &ServiceUser{
		Repo:         repo,
		IServiceAuth: iServiceAuth,
		IRepoAuth:    iRepoAuth,
		Producer:     producer,
		Logger:       logger,
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
	user, errGetUser := s.Repo.GetUserByUUID(userUUID)
	if errGetUser != nil {
		return nil, nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		return nil, nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	var newHashedPassword string
	if body.NewPassword != "" {
		if body.NewEmail != "" {
			if errNewPassword := validate_password.ValidatePassword(body.NewPassword, body.NewEmail, []string{body.NewName, user.Name}); errNewPassword != nil {
				if errors.Is(errNewPassword, custom_errors.ErrPasswordIsNotStrong) {
					return nil, nil, errNewPassword
				}
				return nil, nil, ErrNewPasswordContainEmail
			}
		} else {
			if errNewPassword := validate_password.ValidatePassword(body.NewPassword, user.Email, []string{body.NewName, user.Name}); errNewPassword != nil {
				if errors.Is(errNewPassword, custom_errors.ErrPasswordIsNotStrong) {
					return nil, nil, errNewPassword
				}
				return nil, nil, ErrNewPasswordContainEmail
			}
		}
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
func (s *ServiceUser) RemoveUser(body *RequestRemoveUser, typeRemove string) (*common.ResponseAuth, error) {
	if typeRemove != shared_constant.TypeSoftDelete && typeRemove != shared_constant.TypeHardDelete {
		return nil, shared_errors.ErrIncorrectTypeRemove
	}
	password, errGetPassword := s.Repo.GetPasswordByEmail(body.Email)
	if errGetPassword != nil {
		return nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	if bcrypt.CompareHashAndPassword([]byte(password), []byte(body.Password)) != nil {
		return nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	const sizeRemoveDataMap = 2
	dataUser := make(map[string]string, sizeRemoveDataMap)
	dataUser[emailKey] = body.Email
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
	if action != shared_constant.TypeHardDelete && action != shared_constant.TypeSoftDelete && action != actionUpdate {
		mapError.Map["action"] = ErrIncorrectAction.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	dataSession, errGetDataSession := s.IRepoAuth.GetUserSession(sessionID, action)
	if errGetDataSession != nil {
		return nil, custom_errors.ErrSessionExpired
	}
	if len(dataSession) == 0 {
		return nil, custom_errors.ErrSessionExpired
	}
	if strconv.Itoa(codeUser) != dataSession[common.CodeKey] {
		return nil, custom_errors.ErrIncorrectCode
	}
	if action == actionUpdate {
		userExist, errCheckUser := s.Repo.UserExistsByUserUUID(userUUID)
		if errCheckUser != nil {
			return nil, shared_errors.ErrCriticalServer
		}
		if !userExist {
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
	if action == shared_constant.TypeHardDelete {
		if s.Repo.DeleteUser(userUUID) != nil {
			return nil, ErrFailedDeleteUser
		}
		refreshID, errGetRefresh := s.IRepoAuth.GetRefreshID(userUUID)
		if errGetRefresh != nil {
			s.Logger.Error("failed to get refreshID: " + action + " - " + errGetRefresh.Error())
		} else {
			if errDelRefresh := s.IRepoAuth.DeleteRefresh(userUUID, refreshID); errDelRefresh != nil {
				s.Logger.Error("failed to delete refresh session: " + action + " - " + errDelRefresh.Error())
			}
		}
		event := make(map[string]any, 1)
		event[shared_constant.EventDeletedUserUUID] = userUUID
		ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_kafka.SendMessageTimeout)
		defer cancel()
		if errSendEvent := s.Producer.SendEvent(ctxTimeout, userUUID, event); errSendEvent != nil {
			s.Logger.Error("failed to send event deleted_user: " + errSendEvent.Error())
			return nil, ErrFailedDeleteUser
		}
	} else if action == shared_constant.TypeSoftDelete {
		if s.Repo.RemoveUser(userUUID) != nil {
			return nil, ErrFailedRemoveUser
		}
		refreshID, errGetRefresh := s.IRepoAuth.GetRefreshID(userUUID)
		if errGetRefresh != nil {
			s.Logger.Error("failed to get refreshID: " + action + " - " + errGetRefresh.Error())
		} else {
			if errDelRefresh := s.IRepoAuth.DeleteRefresh(userUUID, refreshID); errDelRefresh != nil {
				s.Logger.Error("failed to delete refresh session: " + action + " - " + errDelRefresh.Error())
			}
		}
	} else {
		return nil, custom_errors.ErrSessionExpired
	}
	return nil, nil
}
func (s *ServiceUser) DeleteExpiredUsers() {
	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if sliceDeleteUserUUID, errDelete := s.Repo.deleteUsersByTimer(); errDelete == nil && len(sliceDeleteUserUUID) != 0 {
				for _, userUUID := range sliceDeleteUserUUID {
					event := make(map[string]any, 1)
					event[shared_constant.EventDeletedUserUUID] = userUUID
					ctxTimeout, cancel := context.WithTimeout(context.Background(), shared_kafka.SendMessageTimeout)
					if errSendEvent := s.Producer.SendEvent(ctxTimeout, userUUID, event); errSendEvent != nil {
						s.Logger.Error("failed to send event deleted_user: " + errSendEvent.Error())
					}
					cancel()
				}
			}
		}
	}
}
