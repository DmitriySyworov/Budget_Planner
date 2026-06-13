package main

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/user"
	"context"
	"net/http"
	"os"
	"shared/loggers"
	"shared/open_db"
	"time"
)

func main() {
	logger := loggers.NewLogger()
	//
	conf := authconfig.Config{}
	//
	postgres := open_db.OpenPostgres(conf.DSN, logger)
	redis := open_db.OpenRedis(conf.RedisAddress, conf.RedisPassword)
	//
	router := http.NewServeMux()
	//
	repoUser := user.NewRepositoryUser(postgres)
	//
	serviceUser := user.NewServiceUser(repoUser)
	//
	router.HandleFunc("GET /health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to writer health check: ", errWrite)
		}
	})
	router.HandleFunc("GET /ready", func(writer http.ResponseWriter, request *http.Request) {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		sqlDb, errDb := postgres.DB.DB()
		if errDb != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Postgres init): ", errDb)
			return
		}
		if errPingPostgres := sqlDb.PingContext(ctxTimeout); errPingPostgres != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Postgres ping): ", errPingPostgres)
			return
		}
		if errPingRedis := redis.Ping(ctxTimeout).Err(); errPingRedis != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Redis ping): ", errPingRedis)
			return
		}
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("READY")); errWrite != nil {
			logger.Error("failed to write ready check: " + errWrite.Error())
			return
		}
	})
	user.NewHandlerUser(router, serviceUser)
	//
	service := http.Server{
		Addr:    ":" + conf.ApiPort,
		Handler: router,
	}
	if errApi := service.ListenAndServe(); errApi != nil {
		logger.Error("critical error on the server: ", errApi)
		os.Exit(1)
	}
}
