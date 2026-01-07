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
	db, err := repository.NewStorage(dbURL)
	if err != nil {
		logger.Fatal("Error while initialization database", err)
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
			logger.Fatal("Server ListenAndServe error", err)
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
		logger.Warn("Server forced to shutdown", err)
	} else {
		logger.Info("Server stopped gracefully")
	}

	if err := db.Close(); err != nil {
		logger.Warn("Database close error", err)
	} else {
		logger.Info("Database connection closed")
	}
}
