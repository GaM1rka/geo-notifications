package main

import (
	"geo-notifications/internal/handler"
	"net/http"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
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
}
