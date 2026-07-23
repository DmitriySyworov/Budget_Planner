package auth

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/JWT"
	"app/auth-service/internal/common"
	"app/auth-service/internal/custom_errors"
	"app/auth-service/internal/di"
	"app/auth-service/internal/model"
	"app/auth-service/internal/send_letter"
	"app/auth-service/internal/validate_password"
	"errors"
	"shared/loggers"
	"shared/shared_errors"
	"strconv"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ServiceAuth struct {
	Repo      IRepositoryRedis
	IRepoUser di.IRepoUser
	Conf      *authconfig.Config
	Logger    *loggers.Logger
}

func NewServiceAuth(repo IRepositoryRedis, repoUser di.IRepoUser, conf *authconfig.Config, logger *loggers.Logger) *ServiceAuth {
	return &ServiceAuth{
		Repo:      repo,
		IRepoUser: repoUser,
		Conf:      conf,
		Logger:    logger,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Validates name length, strong password policy, and email uniqueness before triggering internal registration workflow.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RequestRegister  true  "Registration payload"
// @Success      202      {object}  response.Response{data=common.ResponseAuth,errors=nil} "User registration process accepted"
// @Failure      400      {object}  response.NegativeResponse "Client validation errors. Format: { \"errors\": { \"name\": \"name must be between 2 and 64 characters\" } } or { \"errors\": { \"email\": \"incorrect email\" } } or { \"errors\": { \"password\": \"password must be between 8 and 24 characters\" } } or { \"errors\": { \"user\": \"user already exist\" } }"
// @Failure      429      {object}  response.NegativeResponse "Too many requests. Format: { \"errors\": { \"global\": \"the limit for sending requests per minute has been exceeded\" } }"
// @Failure      500      {object}  response.NegativeResponse "Server errors. Format: { \"errors\": { \"global\": \"critical error on the server side\" } } or { \"errors\": { \"global\": \"failed to ensure security\" } }"
// @Router       /register [post]
func (s *ServiceAuth) Register(body *RequestRegister) (*common.ResponseAuth, error) {
	mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
	errValidatePassword := validate_password.ValidatePassword(body.Password, body.Email, []string{body.Name})
	if errValidatePassword != nil {
		mapError.Map["password"] = errValidatePassword.Error()
	}
	userExist, errCheckUser := s.IRepoUser.UserExistsByEmail(body.Email)
	if errCheckUser != nil {
		mapError.Map["global"] = shared_errors.ErrCriticalServer.Error()
	} else if userExist {
		mapError.Map["user"] = ErrUserAlreadyExist.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	hashPassword, errGeneratePassword := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if errGeneratePassword != nil {
		s.Logger.Error("failed to hash the password")
		return nil, custom_errors.ErrFailedSecurity
	}
	const sizeRegisterMap = 4
	dataMap := make(map[string]string, sizeRegisterMap)
	dataMap[nameKey] = body.Name
	dataMap[passwordKey] = string(hashPassword)
	dataMap[emailKey] = body.Email
	respAuth, errAuth := s.HelperAuth(ActionRegister, dataMap)
	if errAuth != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	return respAuth, nil
}
func (s *ServiceAuth) Login(body *RequestLogin) (*common.ResponseAuth, error) {
	hashPassword, errGetPassword := s.IRepoUser.GetPasswordByEmail(body.Email)
	if errGetPassword != nil {
		return nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	if bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(body.Password)) != nil {
		return nil, custom_errors.ErrIncorrectPasswordOrEmail
	}
	const sizeLoginMap = 2
	dataMap := make(map[string]string, sizeLoginMap)
	dataMap[emailKey] = body.Email
	respAuth, errAuth := s.HelperAuth(ActionLogin, dataMap)
	if errAuth != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	return respAuth, nil
}
func (s *ServiceAuth) Recovery(body *RequestRecovery, action string) (*common.ResponseAuth, error) {
	mapError := &shared_errors.MapError{
		Map: make(map[string]string, 2),
	}
	hashedPassword, errGetPassword := s.IRepoUser.GetPasswordByEmail(body.Email)
	if errGetPassword != nil {
		mapError.Map["email"] = custom_errors.ErrNotFoundUser.Error()
	}
	if action != ActionRecoveryPassword && action != ActionRecoveryUser {
		mapError.Map["action"] = ErrIncorrectActionRecovery.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	if action == ActionRecoveryUser {
		if body.Password == "" {
			return nil, ErrPasswordEmpty
		}
		if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(body.Password)) != nil {
			return nil, custom_errors.ErrIncorrectPasswordOrEmail
		}
	}
	const sizeRecoveryMap = 2
	dataMap := make(map[string]string, sizeRecoveryMap)
	dataMap[emailKey] = body.Email
	respAuth, errAuth := s.HelperAuth(action, dataMap)
	if errAuth != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	return respAuth, nil
}
func (s *ServiceAuth) HelperAuth(action string, dataUser map[string]string) (*common.ResponseAuth, error) {
	sender := send_letter.NewSendLetter(s.Conf.SMTP, s.Logger)
	sessionID := uuid.New().String()

	code, errCode := send_letter.GenerateCode()
	if errCode != nil {
		s.Logger.Error("failed create code: ", errCode)
		return nil, custom_errors.ErrFailedSecurity
	}
	if sender.SendEmailLetter(common.ServiceName, dataUser[emailKey], code) != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	dataUser[common.CodeKey] = strconv.Itoa(code)
	if s.Repo.CreateUserSession(sessionID, action, dataUser) != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	j := JWT.NewJWT(s.Conf.Signature, s.Logger)
	token, errJwtSession := j.CreateSessionJWT(sessionID)
	if errJwtSession != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	return &common.ResponseAuth{
		Message:    "we have sent a confirmation code to the following email address: " + dataUser[emailKey],
		SessionJwt: token,
	}, nil
}

const (
	ActionRegister         = "register"
	ActionLogin            = "login"
	ActionRecoveryUser     = "recovery_user"
	ActionRecoveryPassword = "recovery_password"

	nameKey     = "name"
	emailKey    = "email"
	passwordKey = "password"
)

func (s *ServiceAuth) Confirm(body *RequestConfirm, sessionID, action, userAgent string) (*ResponseConfirm, error) {
	mapError := shared_errors.MapError{Map: make(map[string]string, 2)}
	if len(sessionID) != 36 {
		mapError.Map["session"] = custom_errors.ErrIncorrectSessionID.Error()
	}
	if action != ActionRecoveryUser && action != ActionRecoveryPassword && action != ActionLogin && action != ActionRegister {
		mapError.Map["action"] = ErrIncorrectAction.Error()
	}
	if len(mapError.Map) != 0 {
		return nil, mapError
	}
	if action == ActionRecoveryPassword && body.NewPassword == "" {
		return nil, ErrNotSpecifiedNewPassword
	}
	dataUser, errGetDataSession := s.Repo.GetUserSession(sessionID, action)
	if errGetDataSession != nil {
		return nil, custom_errors.ErrSessionExpired
	}
	if len(dataUser) == 0 {
		return nil, custom_errors.ErrSessionExpired
	}
	if strconv.Itoa(body.Code) != dataUser[common.CodeKey] {
		return nil, custom_errors.ErrIncorrectCode
	}
	var userUUID string
	switch action {
	case ActionRegister:
		userExist, errCheckUser := s.IRepoUser.UserExistsByEmail(dataUser[emailKey])
		if errCheckUser != nil {
			return nil, shared_errors.ErrCriticalServer
		}
		if userExist {
			return nil, ErrUserAlreadyExist
		}
		userUUID = uuid.New().String()
		if s.IRepoUser.CreateUser(&model.Users{
			Name:     dataUser[nameKey],
			Email:    dataUser[emailKey],
			Password: dataUser[passwordKey],
			UserUUID: userUUID,
		}) != nil {
			return nil, ErrCreateUser
		}
	case ActionLogin:
		if uUUID, errGetUserUUID := s.IRepoUser.GetUserUUIDByEmail(dataUser[emailKey]); errGetUserUUID != nil {
			return nil, custom_errors.ErrNotFoundUser
		} else {
			userUUID = uUUID
		}
	case ActionRecoveryUser:
		if uUUID, errGetUserUUID := s.IRepoUser.GetUserUUIDByEmail(dataUser[emailKey]); errGetUserUUID != nil {
			return nil, custom_errors.ErrNotFoundUser
		} else {
			userUUID = uUUID
			if s.IRepoUser.RecoveryUser(uUUID) != nil {
				return nil, ErrFailedRecoveryUser
			}
		}
	case ActionRecoveryPassword:
		user, errGetUser := s.IRepoUser.GetUserByEmail(dataUser[emailKey])
		if errGetUser != nil {
			return nil, custom_errors.ErrNotFoundUser
		}
		userUUID = user.UserUUID
		hashPassword, errHashPass := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
		if errHashPass != nil {
			s.Logger.Error("failed to hashed password: " + errHashPass.Error())
			return nil, custom_errors.ErrFailedSecurity
		}
		user.Password = string(hashPassword)
		if s.IRepoUser.UpdateUser(user, user.UserUUID) != nil {
			return nil, ErrChangePassword
		}
	}
	refreshID := uuid.New().String()
	respConfirm, errConfirm := s.helperCreateToken(userUUID, refreshID)
	if errConfirm != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	if action != ActionRegister {
		oldRefreshID, errGetRefreshID := s.Repo.GetRefreshID(userUUID)
		if errGetRefreshID != nil {
			s.Logger.Warn("failed to get old refreshID in action:" + action + " - " + errGetRefreshID.Error())
		}
		if errDeleteOldRefresh := s.Repo.DeleteRefresh(userUUID, oldRefreshID); errDeleteOldRefresh != nil {
			s.Logger.Warn("failed to delete old refreshID in action:" + action + " - " + errDeleteOldRefresh.Error())
		}
	}
	if s.Repo.CreateRefresh(refreshID, userUUID, userAgent) != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	return respConfirm, nil
}
func (s *ServiceAuth) Refresh(oldRefreshToken, userAgent string) (*ResponseConfirm, error) {
	j := JWT.NewJWT(s.Conf.Signature, s.Logger)
	oldRefreshID, errParseRefresh := j.ParseRefreshToken(oldRefreshToken)
	if errParseRefresh != nil {
		return nil, errParseRefresh
	}
	refreshValues, errGetRefresh := s.Repo.GetRefresh(oldRefreshID)
	if errGetRefresh != nil {
		return nil, ErrRenewalRefresh
	}
	if userAgent != refreshValues.UserAgent {
		return nil, ErrRenewalRefresh
	}
	newRefreshID := uuid.New().String()
	respConfirm, errConfirm := s.helperCreateToken(refreshValues.UserUUID, newRefreshID)
	if errConfirm != nil {
		return nil, custom_errors.ErrFailedSecurity
	}
	if errRotation := s.Repo.RotationRefresh(refreshValues.UserUUID, newRefreshID, oldRefreshID, refreshValues.UserAgent); errRotation != nil {
		if errors.Is(errRotation, ErrRefreshExpired) {
			return nil, ErrRefreshExpired
		}
		return nil, custom_errors.ErrFailedSecurity
	}
	return respConfirm, nil
}
func (s *ServiceAuth) Logout(refreshToken string) {
	j := JWT.NewJWT(s.Conf.Signature, s.Logger)
	refreshID, errParseRefresh := j.ParseRefreshToken(refreshToken)
	if errParseRefresh != nil {
		return
	}
	refreshValues, errGetRefresh := s.Repo.GetRefresh(refreshID)
	if errGetRefresh != nil {
		return
	}
	if s.Repo.DeleteRefresh(refreshValues.UserUUID, refreshID) != nil {
		return
	}
}
func (s *ServiceAuth) helperCreateToken(userUUID, refreshID string) (*ResponseConfirm, error) {
	j := JWT.NewJWT(s.Conf.Signature, s.Logger)
	accessJwt, errCreateAccess := j.CreateAccessJWT(userUUID)
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
