package send_letter

import (
	authconfig "app/auth-service/config"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"shared/loggers"
	"strconv"
	"time"

	"github.com/jordan-wright/email"
	"github.com/sony/gobreaker"
)

type SendLetter struct {
	ApiEmail        string
	ApiPassword     string
	SmtpAddress     string
	SmtpAddressHost string
	Logger          *loggers.Logger
	CircuitBreaker  *gobreaker.CircuitBreaker
}

var (
	ErrSendTimeExpired = errors.New("the time to send the email letter has expired: ")
)

type ISendLetter interface {
	SendEmailLetter(serviceName, userEmail string, code int) error
}
type MockSendLetter struct{}

func (ml *MockSendLetter) SendEmailLetter(serviceName, userEmail string, code int) error { return nil }

func NewSendLetter(conf *authconfig.SMTP, logger *loggers.Logger) ISendLetter {
	if conf.ApiEmail == "" || conf.ApiEmail == "example@gmail.com" {
		return &MockSendLetter{}
	}
	circuitBreaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "SMTP-breaker",
		MaxRequests: 3,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn(fmt.Sprintf("Circuit Breaker [%s] changed condition: %s -> %s", name, from, to))
		},
	})
	return &SendLetter{
		ApiEmail:        conf.ApiEmail,
		ApiPassword:     conf.ApiPassword,
		SmtpAddress:     conf.SmtpAddress,
		SmtpAddressHost: conf.SmtpAddressHost,
		Logger:          logger,
		CircuitBreaker:  circuitBreaker,
	}
}
func (l *SendLetter) SendEmailLetter(serviceName, userEmail string, code int) error {
	after := time.After(time.Second * 10)
	send := email.NewEmail()
	_, errSendLetter := l.CircuitBreaker.Execute(func() (interface{}, error) {
		send.From = l.ApiEmail
		send.To = []string{userEmail}
		send.Subject = "Verification letter from the " + serviceName + " service"
		send.Text = []byte("If you are performing an action on " + serviceName + ", please use the following authorization code: " + strconv.Itoa(code))

		errSend := make(chan error)
		go func() {
			errSend <- send.Send(l.SmtpAddressHost, smtp.PlainAuth("", l.ApiEmail, l.ApiPassword, l.SmtpAddress))
		}()
		select {
		case <-after:
			return nil, ErrSendTimeExpired
		case errSmtp := <-errSend:
			if errSmtp != nil {
				return nil, errors.New(errSmtp.Error() + ":")
			}
			return nil, nil
		}
	})
	if errSendLetter != nil {
		l.Logger.Error(errSendLetter.Error() + userEmail)
		return errSendLetter
	}
	return nil
}

func GenerateCode() (int, error) {
	maxValue := big.NewInt(900000)
	code, errRand := rand.Int(rand.Reader, maxValue)
	if errRand != nil {
		return 0, errRand
	}
	return int(code.Int64()) + 100000, nil
}
