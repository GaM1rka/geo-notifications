package main

import (
	"context"
	"geo-notifications/internal/config"
	"geo-notifications/internal/handler"
	"geo-notifications/internal/repository"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	dbURL := config.GetDBURL()
	redisAddr := config.GetRedisConfig()
	if dbURL == "" {
		logger.Fatal("Database URL is empty")
	}
	if redisAddr.Addr == "" {
		logger.Fatal("Redis address is empty")
	}
	storage, err := repository.NewStorage(dbURL, redisAddr)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize storage")
	}
	h := handler.NewHandler(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/incidents", h.IncidentsHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server ListenAndServe error")
		}
	}()

	logger.Info("Server started on :8080 port")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	logger.Info("Shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Warn("Server forced to shutdown")
	} else {
		logger.Info("Server stopped gracefully")
	}

	if err := storage.Close(); err != nil {
		logger.WithError(err).Warn("Database close error")
	} else {
		logger.Info("Database connection closed")
	}
}
