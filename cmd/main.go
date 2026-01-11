package main

import (
	"context"
	"geo-notifications/internal/config"
	"geo-notifications/internal/handler"
	"geo-notifications/internal/repository"
	"geo-notifications/internal/service"
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
	redisCfg := config.GetRedisConfig()
	if dbURL == "" {
		logger.Fatal("DATABASE_URL is empty")
	}
	if redisCfg.Addr == "" {
		logger.Fatal("REDIS_ADDR is empty")
	}

	// общий контекст с сигналами
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// init storage (Postgres + Redis)
	storage, err := repository.NewStorage(dbURL, redisCfg)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize storage")
	}

	if err := storage.CreateTables(ctx); err != nil {
		logger.WithError(err).Fatal("failed to create tables")
	}

	// init service
	incidentService := service.NewIncidentService(storage, logger)

	// init handler
	h := handler.NewHandler(logger, incidentService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/incidents", h.IncidentsHandler)
	mux.HandleFunc("/api/v1/incidents/", h.IncidentByIDHandler)
	mux.HandleFunc("/api/v1/location/check", h.LocationHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("server ListenAndServe error")
		}
	}()

	logger.Info("Server started on :8080")

	worker := service.NewWebhookWorker(storage, logger, os.Getenv("WEBHOOK_URL"))
	go worker.Run(ctx)

	// ждём сигнал
	<-ctx.Done()
	logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Warn("server forced to shutdown")
	} else {
		logger.Info("Server stopped gracefully")
	}

	if err := storage.Close(); err != nil {
		logger.WithError(err).Warn("storage close error")
	} else {
		logger.Info("Storage closed")
	}
}
