package domain

import "time"

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Service   string    `json:"service"`
}

// HealthUseCase defines the interface for health check operations
type HealthUseCase interface {
	CheckHealth() *HealthStatus
}
