package service

import (
	"context"
	"geo-notifications/internal/model"
	"geo-notifications/internal/repository"

	"github.com/sirupsen/logrus"
)

type IncidentService struct {
	storage repository.Storage
	logger  *logrus.Logger
}

func NewIncidentService(storage repository.Storage, logger *logrus.Logger) *IncidentService {
	return &IncidentService{
		storage: storage,
		logger:  logger,
	}
}

func (is *IncidentService) CreateIncident(req model.Incident) error {
	if req.Title == "" {
		is.logger.Error("Required incident title in CreateIncident request")
	}
	if req.RadiusM < 0 {
		is.logger.Error("Incident radius need to be positive in CreateIncident request")
	}

	_, err := is.storage.Create(context.Background(), &req)
	if err != nil {
		is.logger.Warn("Failed to create incident")
		return err
	}
	return nil
}
