package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"geo-notifications/internal/model"
	"geo-notifications/internal/service"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	logger             *logrus.Logger
	service            *service.IncidentService
	statsWindowMinutes int
}

func NewHandler(logger *logrus.Logger, svc *service.IncidentService, statsWindowMinutes int) *Handler {
	return &Handler{
		logger:             logger,
		service:            svc,
		statsWindowMinutes: statsWindowMinutes,
	}
}

func (h *Handler) IncidentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateIncident(w, r)
	case http.MethodGet:
		h.ListIncidents(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) IncidentByIDHandler(w http.ResponseWriter, r *http.Request) {
	// ожидаем путь вида /api/v1/incidents/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}

	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid incident id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetIncidentByID(w, r, id)
	case http.MethodPut:
		h.UpdateIncident(w, r, id)
	case http.MethodDelete:
		h.DeactivateIncident(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) LocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req model.LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("invalid request body in LocationHandler")
		http.Error(w, "invalid request body to location check", http.StatusBadRequest)
		return
	}

	locations, err := h.service.CheckLocations(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("error while checking location")
		http.Error(w, "location check error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(locations); err != nil {
		h.logger.WithError(err).Error("error while writing response to location check request")
	}
}

func (h *Handler) IncidentsStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	minutes := h.statsWindowMinutes

	count, err := h.service.GetUserStats(r.Context(), minutes)
	if err != nil {
		h.logger.WithError(err).Error("failed to get incidents stats")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		UserCount int `json:"user_count"`
	}{
		UserCount: count,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var incident model.Incident
	if err := json.NewDecoder(r.Body).Decode(&incident); err != nil {
		h.logger.WithError(err).Info("invalid request body in CreateIncident")
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateIncident(r.Context(), &incident); err != nil {
		h.logger.WithError(err).Info("error in service CreateIncident call")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(incident); err != nil {
		h.logger.WithError(err).Error("failed to write response")
	}
}

func (h *Handler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page := 1
	if v := q.Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			http.Error(w, "invalid page parameter", http.StatusBadRequest)
			h.logger.WithError(err).Info("error parsing page parameter")
			return
		}
		page = p
	}

	pageSize := 20
	if v := q.Get("page_size"); v != "" {
		ps, err := strconv.Atoi(v)
		if err != nil || ps < 1 {
			http.Error(w, "invalid page_size parameter", http.StatusBadRequest)
			h.logger.WithError(err).Info("error parsing page_size parameter")
			return
		}
		pageSize = ps
	}

	items, err := h.service.GetItemsList(r.Context(), page, pageSize)
	if err != nil {
		h.logger.WithError(err).Info("error while getting list of incidents")
		http.Error(w, "server error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetIncidentByID(w http.ResponseWriter, r *http.Request, id int64) {
	incident, err := h.service.GetIncidentByID(r.Context(), id)
	if err != nil {
		h.logger.WithError(err).Error("error getting incident by id")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if incident == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(incident)
}

func (h *Handler) UpdateIncident(w http.ResponseWriter, r *http.Request, id int64) {
	defer r.Body.Close()

	var incident model.Incident
	if err := json.NewDecoder(r.Body).Decode(&incident); err != nil {
		h.logger.WithError(err).Info("invalid request body in UpdateIncident")
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	incident.ID = id

	if err := h.service.UpdateIncident(r.Context(), &incident); err != nil {
		h.logger.WithError(err).Error("error updating incident")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(incident)
}

// DELETE /api/v1/incidents/{id} (деактивация)
func (h *Handler) DeactivateIncident(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.service.DeactivateIncident(r.Context(), id); err != nil {
		h.logger.WithError(err).Error("error deactivating incident")
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
