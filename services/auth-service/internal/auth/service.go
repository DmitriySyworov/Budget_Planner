package auth

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/JWT"
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/di"
	"app/auth-service/internal/model"
	"app/auth-service/internal/send_letter"
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
	respAuth, sessionID, errAuth := s.HelperAuth(body.Email, s.Conf)
	if errAuth != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	if s.Repo.CreateDataUserSession(body.Name, body.Email, string(hashPassword), sessionID) != nil {
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
	respAuth, _, errAuth := s.HelperAuth(body.Email, s.Conf)
	if errAuth != nil {
		return nil, []string{custom_errors.ErrFailedSecurity.Error()}
	}
	return respAuth, nil
}
func (s *ServiceAuth) HelperAuth(userEmail string, conf *authconfig.VerifyEmail) (*common.ResponseAuth, string, error) {
	sender := send_letter.NewSendLetter(s.Conf, s.Logger)
	sessionID := uuid.New().String()
	code, errCode := send_letter.GenerateCode()
	if errCode != nil {
		s.Logger.Error("failed rand: ", errCode)
		return nil, "", custom_errors.ErrFailedSecurity
	}
	if sender.SendEmailLetter(userEmail, code) != nil {
		return nil, "", custom_errors.ErrFailedSecurity
	}
	if s.Repo.CreateSession(sessionID, code) != nil {
		return nil, "", custom_errors.ErrFailedSecurity
	}
	j := JWT.NewJWT(conf.Signature, s.Logger)
	token, errJwtSession := j.CreateSessionJWT(sessionID)
	if errJwtSession != nil {
		return nil, "", custom_errors.ErrFailedSecurity
	}
	return &common.ResponseAuth{
		Message:    "we have sent a confirmation code to the following email address: " + userEmail,
		JwtSession: token,
	}, sessionID, nil
}

const (
	actionRegister = "register"
	actionLogin    = "login"
	actionRecovery = "recovery"
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
	code, errGetCode := s.Repo.GetSession(sessionID)
	if errGetCode != nil {
		return nil, []string{custom_errors.ErrSessionExpired.Error()}
	}
	if codeUser != code {
		return nil, []string{custom_errors.ErrIncorrectCode.Error()}
	}
	var userUUID string
	switch action {
	case actionRegister:
		dataUser, errGetDataUser := s.Repo.GetDataUserSession(sessionID)
		if errGetDataUser != nil {
			return nil, []string{custom_errors.ErrSessionExpired.Error()}
		}
		if s.IRepoUser.UserExistsByEmail(dataUser.Email) {
			return nil, []string{ErrUserAlreadyExist.Error()}
		}
		userUUID = uuid.New().String()
		if s.IRepoUser.CreateUser(&model.User{
			Name:     dataUser.Name,
			Email:    dataUser.Email,
			Password: dataUser.Password,
			UserUUID: userUUID,
		}) != nil {
			return nil, []string{ErrCreateUser.Error()}
		}
	case actionLogin:
		userUUID = uuid.New().String()
	case actionRecovery:
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
