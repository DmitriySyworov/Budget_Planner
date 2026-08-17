package listener

import (
	"app/auth-service/internal/di"
	"context"
	"shared/loggers"
	"shared/shconstant"
	"shared/shkafka"
	"shared/shprotos/event"
	"shared/storage"
	"time"

	"google.golang.org/protobuf/proto"
)

type KafkaListener struct {
	Redis                *storage.Redis
	IRepoUser            di.IRepoUser
	Logger               *loggers.Logger
	ProducerNotification *shkafka.KafkaProducer
}

func NewKafkaListener(producer *shkafka.KafkaProducer, redis *storage.Redis, iRepoUser di.IRepoUser, logger *loggers.Logger) *KafkaListener {
	return &KafkaListener{
		Redis:                redis,
		IRepoUser:            iRepoUser,
		ProducerNotification: producer,
		Logger:               logger,
	}
}

func (n *KafkaListener) HandleEventExpenseNotification(dataEvent []byte) error {
	eventLetter := &event.NotificationEvent{}
	if errUnmarshal := proto.Unmarshal(dataEvent, eventLetter); errUnmarshal != nil {
		return errUnmarshal
	}
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shconstant.CtxTimeoutRedis)
	defer cancel()
	success, errSetNX := n.Redis.SetNX(ctxTimeout, eventLetter.GetEventUUID(), 1, shconstant.EventKafkaTTL).Result()
	if errSetNX != nil {
		return errSetNX
	}
	if !success {
		return nil
	}
	userUUID := eventLetter.GetExpense().GetUserUUID()
	keyEmail := shconstant.EventKey + "email:" + userUUID
	var resEmail string
	if email, errGet := n.Redis.Get(ctxTimeout, keyEmail).Result(); errGet == nil {
		resEmail = email
		if errSet := n.Redis.Set(ctxTimeout, keyEmail, email, time.Hour*1).Err(); errSet != nil {
			n.Logger.Error("failed to cash user email: " + errSet.Error())
		}
	} else {
		emailU, errGetEmail := n.IRepoUser.GetEmailByUUID(context.Background(), userUUID)
		if errGetEmail != nil {
			return errGetEmail
		}
		resEmail = emailU
		if errSet := n.Redis.Set(ctxTimeout, keyEmail, emailU, time.Hour*1).Err(); errSet != nil {
			n.Logger.Error("failed to cash user email: " + errSet.Error())
		}
	}
	eventLetter.EmailTo = resEmail
	resDataEvent, errMarshal := proto.Marshal(eventLetter)
	if errMarshal != nil {
		return errMarshal
	}
	if errSendEvent := n.ProducerNotification.SendEvent(context.Background(), userUUID, resDataEvent); errSendEvent != nil {
		return errSendEvent
	}
	return nil
}
