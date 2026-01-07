package handler

import (
	"encoding/json"
	"net/http"

	"geo-notifications/internal/model"
	"geo-notifications/internal/service"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	logger  *logrus.Logger
	service *service.IncidentService
}

func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

func (h *Handler) IncidentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateIncidentRequest(w, r)

	}
}

func (h *Handler) CreateIncidentRequest(w http.ResponseWriter, r *http.Request) {
	var incident model.Incident
	if err := json.NewDecoder(r.Body).Decode(&incident); err != nil {
		h.logger.WithError(err).Info("Invalid request body in IncidentsHandler")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}

	h.service.CreateIncident(incident)

}
