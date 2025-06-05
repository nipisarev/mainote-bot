package usecase

import (
	"time"

	"mainote-backend/internal/domain"
)

type healthUseCase struct{}

// NewHealthUseCase creates a new health use case
func NewHealthUseCase() domain.HealthUseCase {
	return &healthUseCase{}
}

// CheckHealth returns the current health status of the service
func (h *healthUseCase) CheckHealth() *domain.HealthStatus {
	return &domain.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Service:   "mainote-backend",
	}
}
