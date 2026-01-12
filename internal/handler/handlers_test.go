package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"geo-notifications/internal/config"
	"geo-notifications/internal/model"
	"geo-notifications/internal/repository"
	"geo-notifications/internal/service"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestIncidentsHandlerSuccessCase(t *testing.T) {
	dbURL := config.GetDBURL()
	redisCfg := config.GetRedisConfig()
	if dbURL == "" || redisCfg.Addr == "" {
		t.Skip("DATABASE_URL or REDIS_ADDR not set, skipping integration test")
	}

	logger := logrus.New()

	storage, err := repository.NewStorage(dbURL, redisCfg)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()
	if err := storage.CreateTables(ctx); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	incidentService := service.NewIncidentService(storage, logger)

	statsMinutes := 10
	if v := os.Getenv("STATS_TIME_WINDOW_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			statsMinutes = n
		}
	}
	h := NewHandler(logger, incidentService, statsMinutes)

	incidentReq := model.Incident{
		Title:       "Meteorite",
		Description: "A large meteorite fell in the forest.",
		Latitude:    52.0,
		Longitude:   52.0,
		RadiusM:     100,
		Active:      true,
	}

	body, err := json.Marshal(&incidentReq)
	if err != nil {
		t.Fatalf("failed to marshal incident: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.IncidentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, resp.StatusCode, string(data))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var created model.Incident
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if created.ID == 0 {
		t.Errorf("expected non-zero ID in created incident")
	}
	if created.Title != incidentReq.Title {
		t.Errorf("expected title %q, got %q", incidentReq.Title, created.Title)
	}
}
