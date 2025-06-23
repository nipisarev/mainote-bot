package handler

import (
	"context"
	"mainote-server/internal/domain"
	"mainote-server/pkg/generated/api"
)

type HealthHandler struct {
	healthUseCase domain.HealthUseCase
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthUseCase domain.HealthUseCase) *HealthHandler {
	return &HealthHandler{
		healthUseCase: healthUseCase,
	}
}

// CheckHealth handles health check requests
func (h *HealthHandler) CheckHealth(ctx context.Context) (api.ImplResponse, error) {
	healthStatus := h.healthUseCase.CheckHealth()
	return api.Response(200, healthStatus), nil
}
