package send_letter

import (
	authconfig "app/auth-service/config"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"shared/loggers"
	"time"

	"github.com/jordan-wright/email"
)

type SendLetter struct {
	ApiEmail        string
	ApiPassword     string
	SmtpAddress     string
	SmtpAddressHost string
	Logger          *loggers.Logger
}

const (
	ErrSendEmail       = "failed to send email letter to "
	ErrSendTimeExpired = "the time to send the email letter has expired: "
)

func NewSendLetter(conf *authconfig.VerifyEmail, logger *loggers.Logger) *SendLetter {
	return &SendLetter{
		ApiEmail:        conf.ApiEmail,
		ApiPassword:     conf.ApiPassword,
		SmtpAddress:     conf.SmtpAddress,
		SmtpAddressHost: conf.SmtpAddressHost,
		Logger:          logger,
	}
}
func (l *SendLetter) SendEmailLetter(userEmail string, code int) error {
	after := time.After(time.Second * 10)
	for {
		select {
		case <-after:
			l.Logger.Warn(ErrSendTimeExpired + userEmail)
			return errors.New(ErrSendTimeExpired + userEmail)
		default:
			send := email.NewEmail()
			send.From = l.ApiEmail
			send.To = []string{userEmail}
			send.Subject = "Verification letter from the budget-planner service"
			send.Text = []byte(fmt.Sprint("If you are performing an action on budget-planner, please use the following authorization code: ", code))

			var auth smtp.Auth
			if l.ApiPassword != "test" {
				auth = smtp.PlainAuth("", l.ApiEmail, l.ApiPassword, l.SmtpAddress)
			}
			if errSend := send.Send(l.SmtpAddressHost, auth); errSend != nil {
				l.Logger.Warn(ErrSendEmail + userEmail + errSend.Error())
				return errors.New(ErrSendEmail + userEmail)
			}
			return nil
		}
	}
}
func GenerateCode() (int, error) {
	num, errRand := rand.Int(rand.Reader, big.NewInt(900000))
	if errRand != nil {
		return 0, errRand
	}
	return int(num.Int64() + 100000), nil
}
