package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	case http.MethodGet:

	}
}

func (h *Handler) GetPagination(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "Invalid page parameter", http.StatusBadRequest)
		h.logger.WithError(err).Info("Error while parsing page parameter from query parameters")
		return
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil {
		http.Error(w, "Invalid page size parameter", http.StatusBadRequest)
		h.logger.WithError(err).Info("Error while parsing page size parameter from query parameters")
		return
	}

	items, err := h.service.GetItemsList(page, pageSize)

	if err != nil {
		h.logger.WithError(err).Info("Error while getting list of incidents")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Items    []model.Incident `json:"items"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
	}{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateIncidentRequest(w http.ResponseWriter, r *http.Request) {
	var incident model.Incident
	if err := json.NewDecoder(r.Body).Decode(&incident); err != nil {
		h.logger.WithError(err).Info("Invalid request body in IncidentsHandler")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.service.CreateIncident(&incident)
	if err != nil {
		h.logger.WithError(err).Info("Error in service CreateIncident call")
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(incident); err != nil {
		h.logger.WithError(err).Error("failed to write response")
	}
}
