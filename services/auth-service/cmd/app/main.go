package main

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/auth"
	"app/auth-service/internal/user"
	"context"
	"net/http"
	"os"
	"shared/loggers"
	"shared/open_db"
	"shared/response"
	"shared/shared_middleware"
	"time"
)

func main() {
	logger := loggers.NewLogger()
	//
	conf := authconfig.NewConfig(logger)
	//
	responseHandler := response.NewHandlerResponse(logger)
	//
	sharedMv := shared_middleware.NewManagerSharedMiddleware(conf.Signature, logger, responseHandler)
	//
	postgres := open_db.OpenPostgres(conf.DSN, logger)
	redis := open_db.OpenRedis(conf.RedisAddress, conf.RedisPassword)
	//
	router := http.NewServeMux()
	//
	repoAuth := auth.NewRepository(redis, logger)
	repoUser := user.NewRepositoryUser(postgres, logger)
	//
	serviceAuth := auth.NewServiceAuth(repoAuth, repoUser, conf.VerifyEmail, logger)
	serviceUser := user.NewServiceUser(repoUser, serviceAuth, repoAuth)
	//
	router.HandleFunc("GET /health", health(logger))
	router.HandleFunc("GET /ready", ready(postgres, redis, logger))
	auth.NewHandlerAuth(router, serviceAuth, responseHandler, logger, sharedMv)
	user.NewHandlerUser(router, serviceUser, responseHandler, logger, sharedMv)
	//
	chainMv := shared_middleware.Chain(
		sharedMv.Recovery,
		sharedMv.Logging,
	)
	service := http.Server{
		Addr:    ":" + conf.ApiPort,
		Handler: chainMv(router),
	}
	if errApi := service.ListenAndServe(); errApi != nil {
		logger.Error("critical error on the server: ", errApi)
		os.Exit(1)
	}
}
func health(logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to writer health check: ", errWrite)
		}
	}
}
func ready(postgres *open_db.Postgres, redis *open_db.Redis, logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
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
	}
}
