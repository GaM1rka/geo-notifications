package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	logger *logrus.Logger
}

func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) IncidentsHandler(w http.ResponseWriter, r *http.Request) {

}
