package service

import (
	"context"
	"fmt"

	"geo-notifications/internal/model"
	"geo-notifications/internal/repository"

	"github.com/sirupsen/logrus"
)

type IncidentService struct {
	storage *repository.Storage
	logger  *logrus.Logger
}

func NewIncidentService(storage *repository.Storage, logger *logrus.Logger) *IncidentService {
	return &IncidentService{
		storage: storage,
		logger:  logger,
	}
}

func (is *IncidentService) CreateIncident(ctx context.Context, req *model.Incident) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}
	if req.RadiusM < 0 {
		return fmt.Errorf("radius must be positive")
	}

	req.Active = true

	_, err := is.storage.Create(ctx, req)
	if err != nil {
		is.logger.WithError(err).Warn("failed to create incident")
		return err
	}
	return nil
}

func (is *IncidentService) GetItemsList(ctx context.Context, page, pageSize int) ([]model.Incident, error) {
	if page < 1 || pageSize < 1 {
		return nil, fmt.Errorf("invalid pagination parameters: page=%d, pageSize=%d", page, pageSize)
	}

	results, err := is.storage.GetList(ctx, page, pageSize)
	if err != nil {
		is.logger.WithError(err).Info("error while getting list of incidents")
		return nil, err
	}

	return results, nil
}

func (is *IncidentService) GetIncidentByID(ctx context.Context, id int64) (*model.Incident, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id: %d", id)
	}

	incident, err := is.storage.GetByID(ctx, id)
	if err != nil {
		is.logger.WithError(err).Error("error getting incident by id")
		return nil, err
	}
	return incident, nil
}

func (is *IncidentService) UpdateIncident(ctx context.Context, in *model.Incident) error {
	if in.ID <= 0 {
		return fmt.Errorf("invalid id: %d", in.ID)
	}
	if in.Title == "" {
		return fmt.Errorf("title is required")
	}
	if in.RadiusM < 0 {
		return fmt.Errorf("radius must be positive")
	}

	if err := is.storage.Update(ctx, in); err != nil {
		is.logger.WithError(err).Error("failed to update incident")
		return err
	}
	return nil
}

func (is *IncidentService) DeactivateIncident(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %d", id)
	}

	if err := is.storage.Deactivate(ctx, id); err != nil {
		is.logger.WithError(err).Error("failed to deactivate incident")
		return err
	}
	return nil
}

func (is *IncidentService) CheckLocations(ctx context.Context, req model.LocationRequest) (model.LocationResponse, error) {
	if req.UserID <= 0 {
		return model.LocationResponse{}, fmt.Errorf("invalid user_id: %d", req.UserID)
	}
	locations, err := is.storage.GetLocations(ctx, req)
	if err != nil {
		is.logger.WithError(err).Error("failed to get locations")
		return model.LocationResponse{}, err
	}
	return locations, nil
}
