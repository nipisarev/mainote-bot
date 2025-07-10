package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Integration represents an integration between a user and an external app.
type Integration struct {
	IntegrationID uuid.UUID        `json:"integration_id" db:"integration_id"`
	AppID         uuid.UUID        `json:"app_id" db:"app_id"`
	ChatID        string           `json:"chat_id" db:"chat_id"`
	Status        string           `json:"status" db:"status"`
	Config        *json.RawMessage `json:"config" db:"config"`
	AuthType      string           `json:"auth_type" db:"auth_type"`
	AuthData      *json.RawMessage `json:"auth_data" db:"auth_data"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time       `json:"deleted_at,omitempty" db:"deleted_at"`
}

// IntegrationWithApp represents an integration with its associated app information.
type IntegrationWithApp struct {
	Integration
	App App `json:"app"`
}

// IntegrationsListResult represents paginated integrations result.
type IntegrationsListResult struct {
	Integrations []IntegrationWithApp `json:"integrations"`
	Total        int                  `json:"total"`
	Limit        int                  `json:"limit"`
	Offset       int                  `json:"offset"`
	HasMore      bool                 `json:"has_more"`
}
