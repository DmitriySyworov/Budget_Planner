package listener

import (
	"app/budget-planner/internal/di"
	"context"
	"errors"
	"shared/loggers"
	"shared/shconstant"
	"shared/shprotos/event"
	"shared/storage"

	"google.golang.org/protobuf/proto"
)

type KafkaListener struct {
	IRepoBudget di.IRepoBudget
	SharedRedis *storage.Redis
	Logger      *loggers.Logger
}

func NewKafkaListener(shRedis *storage.Redis, iRepoBudget di.IRepoBudget, logger *loggers.Logger) *KafkaListener {
	return &KafkaListener{
		IRepoBudget: iRepoBudget,
		SharedRedis: shRedis,
		Logger:      logger,
	}
}
func (k *KafkaListener) DeleteDataDeletingUser(data []byte) error {
	var eventDel event.DeleteUserDataEvent
	if errUnmarshal := proto.Unmarshal(data, &eventDel); errUnmarshal != nil {
		k.Logger.Error("failed to unmarshal delete event: " + errUnmarshal.Error())
		return errUnmarshal
	}
	ctxTimeout, cancel := context.WithTimeout(context.Background(), shconstant.CtxTimeoutRedis)
	defer cancel()
	eventUUID := eventDel.GetEventUuid()
	keyEvent := shconstant.EventKey + "user:" + eventUUID
	success, errSetNX := k.SharedRedis.SetNX(ctxTimeout, keyEvent, 1, shconstant.EventKafkaTTL).Result()
	if errSetNX != nil {
		k.Logger.Error("failed to set event delete user: " + errSetNX.Error())
		return errSetNX
	}
	if !success {
		k.Logger.Warn("duplicate event pass sent: " + eventUUID)
		return nil
	}

	listUserUUID := eventDel.GetUserUuidList()
	if len(listUserUUID) == 0 {
		k.Logger.Error("event with an empty user_uuid was sent")
		return errors.New("event with an empty user_uuid list was sent")
	}
	if errDel := k.IRepoBudget.DeleteAllUserBudgets(listUserUUID); errDel != nil {
		k.Logger.Warn("failed to delete users budgets maybe it was deleted earlier: " + errDel.Error())
		return errDel
	}
	k.Logger.Info("cascade deletion of all user data was successful")
	return nil
}
