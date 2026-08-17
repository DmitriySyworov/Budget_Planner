package main

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/di"
	"app/auth-service/internal/user"
	"context"
	"os"
	"shared/loggers"
	"shared/shconstant"
	"shared/shkafka"
	"shared/shprotos/event"
	"shared/storage"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func main() {
	logging := loggers.NewLogger()
	//
	conf := authconfig.NewConfig(logging)
	//
	redis := storage.OpenRedis(conf.RedisAddress, conf.RedisPassword, logging)
	defer func() {
		if errClose := redis.Close(); errClose != nil {
			logging.Error("failed to close redis auth: " + errClose.Error())
		}
	}()
	//
	postgres := storage.OpenPostgres(conf.DSN, logging)
	defer func() {
		sqlDb, errGetDriverDb := postgres.DB.DB()
		if errGetDriverDb != nil {
			logging.Error("failed to extract sql driver: " + errGetDriverDb.Error())
			return
		}
		if errClosePostgres := sqlDb.Close(); errClosePostgres != nil {
			logging.Error("failed to close Postgres: " + errClosePostgres.Error())
		}
	}()
	//
	producerDeleteEvent, errInitialProducerDelete := shkafka.NewProducer(&shkafka.ConfigProducer{
		Brokers:       []string{conf.Broker},
		KafkaUser:     conf.KafkaUser,
		KafkaPassword: conf.KafkaPassword,
		Topic:         conf.DeletedUsersTopic,
	}, logging)
	if errInitialProducerDelete != nil {
		logging.Error("failed to initial producer delete event kafka: " + errInitialProducerDelete.Error())
		os.Exit(1)
	}
	defer producerDeleteEvent.CloseProducer()
	//
	repoUser := user.NewRepositoryUser(postgres, logging)
	if DeleteExpiredUsers(producerDeleteEvent, redis, repoUser, logging) != nil {
		os.Exit(1)
	}
}

func DeleteExpiredUsers(producer *shkafka.KafkaProducer, redisAuth *storage.Redis, iRepoUser di.IRepoUser, logger *loggers.Logger) error {
	sliceDeleteUserUUID, errDelete := iRepoUser.DeleteUsersByTimer()
	if errDelete != nil {
		logger.Error("failed to delete users by timer: " + errDelete.Error())
		return errDelete
	}
	if len(sliceDeleteUserUUID) == 0 {
		logger.Warn("not found soft-deleting users")
		return nil
	}
	ctxTimeoutRedis, cancelRdb := context.WithTimeout(context.Background(), shconstant.CtxTimeoutRedis)
	defer cancelRdb()
	if errDelUserRefreshes := redisAuth.Del(ctxTimeoutRedis, sliceDeleteUserUUID...).Err(); errDelUserRefreshes != nil {
		logger.Error("failed to delete user refreshes: " + errDelUserRefreshes.Error())
	}
	eventUUID := uuid.New().String()
	dataEvent, errMarshalEvent := proto.Marshal(&event.DeleteUserDataEvent{
		UserUuidList: sliceDeleteUserUUID,
		EventUuid:    eventUUID,
	})
	if errMarshalEvent != nil {
		logger.Error("failed to marshal proto event: " + errMarshalEvent.Error())
		return errMarshalEvent
	}
	ctxTimeoutProduce, cancelProduce := context.WithTimeout(context.Background(), shconstant.CtxTimeoutSendEventKafka)
	defer cancelProduce()
	if errSendEvent := producer.SendEvent(ctxTimeoutProduce, eventUUID, dataEvent); errSendEvent != nil {
		logger.Error("failed to send event deleted user: " + errSendEvent.Error())
		return errSendEvent
	}
	return nil
}
