package handler

import (
	"encoding/json"
	"net/http"

	"mainote-backend/internal/domain"
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
func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	healthStatus := h.healthUseCase.CheckHealth()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
