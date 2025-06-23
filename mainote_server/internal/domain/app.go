package domain

import (
	"time"

	"github.com/google/uuid"
)

// App represents an application that can be integrated with the system.
// These are available apps like Notion, TickTick, Jira that customers can install.
// swagger:model App
type App struct {
	AppID       uuid.UUID  `json:"app_id" db:"app_id"`
	Provider    string     `json:"provider" db:"provider"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// AppRepository defines the interface for app data operations.
type AppRepository interface {
	GetAll() ([]App, error)
	GetByID(id uuid.UUID) (*App, error)
	GetByProvider(provider string) (*App, error)
}
