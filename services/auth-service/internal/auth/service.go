package auth

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/JWT"
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/di"
	"app/auth-service/internal/model"
	"app/auth-service/internal/send_letter"
	"fmt"
	"shared/loggers"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ServiceAuth struct {
	Repo      *RepositoryAuth
	IRepoUser di.IRepoUser
	Conf      *authconfig.VerifyEmail
	Logger    *loggers.Logger
}

func NewServiceAuth(repo *RepositoryAuth, repoUser di.IRepoUser, conf *authconfig.VerifyEmail, logger *loggers.Logger) *ServiceAuth {
	return &ServiceAuth{
		Repo:      repo,
		IRepoUser: repoUser,
		Conf:      conf,
		Logger:    logger,
	}
}
func (s *ServiceAuth) Register(body *RequestRegister) (*common.ResponseAuth, []string) {
	if s.IRepoUser.UserExistsByEmail(body.Email) {
		return nil, []string{ErrUserAlreadyExist.Error()}
	}
	hashPassword, errGeneratePassword := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if errGeneratePassword != nil {
		s.Logger.Error("failed to hash the password")
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	const sizeRegisterMap = 4
	dataMap := make(map[string]string, sizeRegisterMap)
	dataMap[nameKey] = body.Name
	dataMap[passwordKey] = string(hashPassword)
	dataMap[emailKey] = body.Email
	respAuth, errAuth := s.HelperAuth(actionRecovery, dataMap, s.Conf)
	if errAuth != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respAuth, nil
}
func (s *ServiceAuth) Login(body *RequestLogin) (*common.ResponseAuth, []string) {
	hashPassword, errGetPassword := s.IRepoUser.GetPasswordByEmail(body.Email)
	if errGetPassword != nil {
		return nil, []string{custom_errors.ErrIncorrectPasswordOrEmail.Error()}
	}
	if bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(body.Password)) != nil {
		return nil, []string{custom_errors.ErrIncorrectPasswordOrEmail.Error()}
	}
	const sizeLoginMap = 2
	dataMap := make(map[string]string, sizeLoginMap)
	dataMap[emailKey] = body.Email
	respAuth, errAuth := s.HelperAuth(actionRecovery, dataMap, s.Conf)
	if errAuth != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respAuth, nil
}
func (s *ServiceAuth) Recovery(email string) (*common.ResponseAuth, []string) {
	if !s.IRepoUser.UserExistsByEmail(email) {
		return nil, []string{custom_errors.ErrNotFoundUser.Error()}
	}
	const sizeRecoveryMap = 2
	dataMap := make(map[string]string, sizeRecoveryMap)
	dataMap[emailKey] = email
	respAuth, errAuth := s.HelperAuth(actionRecovery, dataMap, s.Conf)
	if errAuth != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respAuth, nil
}
func (s *ServiceAuth) HelperAuth(action string, dataUser map[string]string, conf *authconfig.VerifyEmail) (*common.ResponseAuth, error) {
	sender := send_letter.NewSendLetter(s.Conf, s.Logger)
	sessionID := uuid.New().String()
	code, errCode := send_letter.GenerateCode()
	if errCode != nil {
		s.Logger.Error("failed rand: ", errCode)
		return nil, custom_errors.ErrFailedSecurity
	}
	if sender.SendEmailLetter(dataUser[emailKey], code) != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	dataUser[common.CodeKey] = fmt.Sprint(code)
	if s.Repo.CreateUserSession(sessionID, action, dataUser) != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	j := JWT.NewJWT(conf.Signature, s.Logger)
	token, errJwtSession := j.CreateSessionJWT(sessionID)
	if errJwtSession != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	return &common.ResponseAuth{
		Message:    "we have sent a confirmation code to the following email address: " + dataUser[emailKey],
		JwtSession: token,
	}, nil
}

const (
	actionRegister = "register"
	actionLogin    = "login"
	actionRecovery = "recovery"

	nameKey     = "name"
	emailKey    = "email"
	passwordKey = "password"
)

func (s *ServiceAuth) Confirm(codeUser int, sessionID, action, userAgent string) (*ResponseConfirm, []string) {
	sliceError := make([]string, 2)
	if len(sessionID) != 36 {
		sliceError = append(sliceError, custom_errors.ErrIncorrectSessionID.Error())
	}
	if action != actionRecovery && action != actionLogin && action != actionRegister {
		sliceError = append(sliceError, ErrIncorrectAction.Error())
	}
	if len(sliceError) != 0 {
		return nil, sliceError
	}
	dataUser, errGetCode := s.Repo.GetUserSession(sessionID, action)
	if errGetCode != nil {
		return nil, []string{custom_errors.ErrSessionExpired.Error()}
	}

	if fmt.Sprint(codeUser) != dataUser[common.CodeKey] {
		return nil, []string{custom_errors.ErrIncorrectCode.Error()}
	}
	var userUUID string
	switch action {
	case actionRegister:
		if s.IRepoUser.UserExistsByEmail(dataUser[emailKey]) {
			return nil, []string{ErrUserAlreadyExist.Error()}
		}
		userUUID = uuid.New().String()
		if s.IRepoUser.CreateUser(&model.User{
			Name:     dataUser[nameKey],
			Email:    dataUser[emailKey],
			Password: dataUser[passwordKey],
			UserUUID: userUUID,
		}) != nil {
			return nil, []string{ErrCreateUser.Error()}
		}
	case actionLogin:
		if uUUID, errGetUserUUID := s.IRepoUser.GetUserUUIDByEmail(dataUser[emailKey]); errGetUserUUID != nil {
			return nil, []string{custom_errors.ErrNotFoundUser.Error()}
		} else {
			userUUID = uUUID
		}
	case actionRecovery:
		if uUUID, errGetUserUUID := s.IRepoUser.GetUserUUIDByEmail(dataUser[emailKey]); errGetUserUUID != nil {
			return nil, []string{custom_errors.ErrNotFoundUser.Error()}
		} else {
			userUUID = uUUID
			if s.IRepoUser.RecoveryUser(uUUID) != nil {
				return nil, []string{ErrFailedRecoveryUser.Error()}
			}
		}
	}
	refreshID := uuid.New().String()
	respConfirm, errConfirm := s.helperConfirm(userUUID, refreshID)
	if errConfirm != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	if s.Repo.CreateRefresh(refreshID, userUUID, userAgent) != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respConfirm, nil
}
func (s *ServiceAuth) Refresh(oldRefreshID, userAgent string) (*ResponseConfirm, []string) {
	dtoRefresh, errGetRefresh := s.Repo.GetRefresh(oldRefreshID)
	if errGetRefresh != nil {
		return nil, []string{ErrRenewalRefresh.Error()}
	}
	if userAgent != dtoRefresh.UserAgent {
		return nil, []string{ErrRenewalRefresh.Error()}
	}
	newRefreshID := uuid.New().String()
	respConfirm, errConfirm := s.helperConfirm(dtoRefresh.UserUUID, newRefreshID)
	if errConfirm != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	if s.Repo.RotationRefresh(newRefreshID, oldRefreshID, dtoRefresh.UserUUID, dtoRefresh.UserAgent) != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respConfirm, nil
}
func (s *ServiceAuth) helperConfirm(userUUID, refreshID string) (*ResponseConfirm, error) {
	j := JWT.NewJWT(s.Conf.Signature, s.Logger)
	accessJwt, errCreateAccess := j.CreateAccessJWT(uuid.New().String(), userUUID)
	if errCreateAccess != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	refreshJwt, errCreateRefresh := j.CreateRefreshJWT(refreshID)
	if errCreateRefresh != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	return &ResponseConfirm{
		AccessJwt:  accessJwt,
		RefreshJwt: refreshJwt,
	}, nil
}
