package main

import (
	"app/gatewaydocs/internal/swaggerdocs"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"shared/loggers"
	"shared/response"
	"shared/shmiddleware"
	"syscall"
	"time"

	_ "github.com/swaggo/http-swagger/v2"
)

func main() {
	logger := loggers.NewLogger()
	//
	apiPort := os.Getenv("EXTERNAL_API_PORT")
	if apiPort == "" {
		logger.Warn("environment variable 'EXTERNAL_API_PORT' not found. Default value = 8080")
		apiPort = "8080"
	}
	//
	respHandler := response.NewHandlerResponse(logger)
	//
	serviceDocs := swaggerdocs.NewServiceSwaggerDocs(logger)
	serviceDocs.UpdateDocs()
	ctxCancel, cancelCancel := context.WithCancel(context.Background())
	go serviceDocs.PlanUpdateDocs(ctxCancel)
	//
	sharedMv := shmiddleware.NewManagerSharedMiddleware("", nil, logger, respHandler)
	//
	router := http.NewServeMux()
	//
	chainMv := shmiddleware.Chain(
		sharedMv.Logging,
		sharedMv.Recovery,
	)
	server := http.Server{
		Addr:    ":" + apiPort,
		Handler: chainMv(router),
	}
	//
	swaggerdocs.NewHandlerSwaggerDocs(router, serviceDocs, respHandler)
	//
	serverError := make(chan error, 1)
	stopSignal := make(chan os.Signal, 1)
	go func() {
		if errServer := server.ListenAndServe(); errServer != nil && !errors.Is(errServer, http.ErrServerClosed) {
			serverError <- errServer
		}
	}()
	signal.Notify(stopSignal, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-stopSignal:
		logger.Info("received stop signal, starting graceful shutdown")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if errShutdown := server.Shutdown(ctxShutdown); errShutdown != nil {
			logger.Error("failed graceful shutdown: " + errShutdown.Error())
		}
		cancel()
	case err := <-serverError:
		logger.Error("error on the server: " + err.Error())
	}
	cancelCancel()
}
