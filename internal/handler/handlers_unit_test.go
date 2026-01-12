package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"geo-notifications/internal/model"
	"geo-notifications/internal/service"

	"github.com/sirupsen/logrus"
)

type fakeIncidentService struct {
	healthErr       *service.HealthError
	createdIncident *model.Incident
	listItems       []model.Incident
}

func (f *fakeIncidentService) HealthCheck(ctx context.Context) *service.HealthError {
	return f.healthErr
}

func (f *fakeIncidentService) CreateIncident(ctx context.Context, inc *model.Incident) error {
	f.createdIncident = inc
	return nil
}

func (f *fakeIncidentService) GetItemsList(ctx context.Context, page, pageSize int) ([]model.Incident, error) {
	return f.listItems, nil
}

func (f *fakeIncidentService) GetIncidentByID(ctx context.Context, id int64) (*model.Incident, error) {
	return nil, nil
}

func (f *fakeIncidentService) GetUserStats(ctx context.Context, minutes int) (int, error) {
	return 0, nil
}

func (f *fakeIncidentService) UpdateIncident(ctx context.Context, in *model.Incident) error {
	return nil
}

func (f *fakeIncidentService) DeactivateIncident(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeIncidentService) CheckLocations(ctx context.Context, req model.LocationRequest) (model.LocationResponse, error) {
	return model.LocationResponse{}, nil
}

func TestHealthHandler_OK(t *testing.T) {
	logger := logrus.New()
	svc := &fakeIncidentService{
		healthErr: nil,
	}
	h := NewHandler(logger, svc, 5)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil)
	w := httptest.NewRecorder()

	h.HealthHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
		DB     string `json:"db"`
		Redis  string `json:"redis"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body.Status != "ok" || body.DB != "ok" || body.Redis != "ok" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestHealthHandler_Degraded(t *testing.T) {
	logger := logrus.New()
	svc := &fakeIncidentService{
		healthErr: &service.HealthError{
			DBError:    errors.New("db error"),
			RedisError: nil,
		},
	}
	h := NewHandler(logger, svc, 5)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil)
	w := httptest.NewRecorder()

	h.HealthHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var body struct {
		Status string `json:"status"`
		DB     string `json:"db"`
		Redis  string `json:"redis"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body.Status != "degraded" {
		t.Fatalf("expected status degraded, got %s", body.Status)
	}
	if body.DB != "error" || body.Redis != "ok" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestIncidentsHandler_CreateIncident(t *testing.T) {
	logger := logrus.New()
	svc := &fakeIncidentService{}
	h := NewHandler(logger, svc, 5)

	body := `{"title":"test incident","description":"desc","latitude":10.5,"longitude":20.5}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h.IncidentsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	// проверим, что сервис получил инцидент
	if svc.createdIncident == nil || svc.createdIncident.Title != "test incident" {
		t.Fatalf("incident was not passed correctly to service: %+v", svc.createdIncident)
	}
}

func TestIncidentsHandler_ListIncidents(t *testing.T) {
	logger := logrus.New()
	svc := &fakeIncidentService{
		listItems: []model.Incident{
			{ID: 1, Title: "i1"},
			{ID: 2, Title: "i2"},
		},
	}
	h := NewHandler(logger, svc, 5)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/incidents?page=2&page_size=10", nil)
	w := httptest.NewRecorder()

	h.IncidentsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var body struct {
		Items    []model.Incident `json:"items"`
		Page     int              `json:"page"`
		PageSize int              `json:"page_size"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body.Page != 2 || body.PageSize != 10 {
		t.Fatalf("unexpected pagination: page=%d page_size=%d", body.Page, body.PageSize)
	}
	if len(body.Items) != 2 || body.Items[0].ID != 1 {
		t.Fatalf("unexpected items: %+v", body.Items)
	}
}
