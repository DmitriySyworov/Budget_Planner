package main

import (
	authconfig "app/auth-service/config"
	"app/auth-service/internal/auth"
	"app/auth-service/internal/middleware"
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
	apiPort, logging, handlers := App()
	service := http.Server{
		Addr:    ":" + apiPort,
		Handler: handlers,
	}
	if errApi := service.ListenAndServe(); errApi != nil {
		logging.Error("critical error on the server: " + errApi.Error())
		os.Exit(1)
	}
}
func App() (string, *loggers.Logger, http.Handler) {
	logging := loggers.NewLogger()
	//
	conf := authconfig.NewConfig(logging)
	//
	responseHandler := response.NewHandlerResponse(logging)
	//
	sharedMv := shared_middleware.NewManagerSharedMiddleware(conf.Signature, logging, responseHandler)
	mv := middleware.NewManagerMiddleware(conf.Signature, logging, responseHandler)
	//
	postgres := open_db.OpenPostgres(conf.DSN, logging)
	redis := open_db.OpenRedis(conf.RedisAddress, conf.RedisPassword)
	//
	router := http.NewServeMux()
	//
	repoAuth := auth.NewRepository(redis, logging)
	repoUser := user.NewRepositoryUser(postgres, logging)
	//
	serviceAuth := auth.NewServiceAuth(repoAuth, repoUser, conf.VerifyEmail, logging)
	serviceUser := user.NewServiceUser(repoUser, serviceAuth, repoAuth)
	//
	router.HandleFunc("GET /health", health(logging))
	router.HandleFunc("GET /ready", ready(postgres, redis, logging))
	auth.NewHandlerAuth(router, serviceAuth, responseHandler, logging, mv)
	user.NewHandlerUser(router, serviceUser, responseHandler, logging, mv, sharedMv)
	//
	chainMv := shared_middleware.Chain(
		sharedMv.Logging,
		sharedMv.Recovery,
	)
	return conf.ApiPort, logging, chainMv(router)
}
func health(logger *loggers.Logger) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("OK")); errWrite != nil {
			logger.Error("failed to writer health check: " + errWrite.Error())
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
			logger.Error("ready check failed (Postgres init): " + errDb.Error())
			return
		}
		if errPingPostgres := sqlDb.PingContext(ctxTimeout); errPingPostgres != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Postgres ping): " + errPingPostgres.Error())
			return
		}
		if errPingRedis := redis.Ping(ctxTimeout).Err(); errPingRedis != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			logger.Error("ready check failed (Redis ping): " + errPingRedis.Error())
			return
		}
		writer.WriteHeader(http.StatusOK)
		if _, errWrite := writer.Write([]byte("READY")); errWrite != nil {
			logger.Error("failed to write ready check: " + errWrite.Error())
			return
		}
	}
}
