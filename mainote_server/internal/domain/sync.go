package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SyncDirection represents the direction of synchronization
type SyncDirection int

const (
	SyncDirectionInbound  SyncDirection = iota // External → Internal
	SyncDirectionOutbound                      // Internal → External
	SyncDirectionBoth                          // Bidirectional
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Success    bool      `json:"success"`
	ExternalID string    `json:"external_id,omitempty"`
	Error      string    `json:"error,omitempty"`
	SyncedAt   time.Time `json:"synced_at"`
}

// SyncProvider defines the interface for different sync providers
type SyncProvider interface {
	// GetProviderName returns the name of the provider (e.g., "notion", "ticktick")
	GetProviderName() string

	// SyncNote synchronizes a note with the external system
	SyncNote(ctx context.Context, note *Note, integration *Integration) (*SyncResult, error)

	// FetchNotes retrieves notes from the external system
	FetchNotes(ctx context.Context, integration *Integration, lastSync *time.Time) ([]*ExternalNote, error)

	// UpdateNote updates an existing note in the external system
	UpdateNote(ctx context.Context, note *Note, integration *Integration, externalID string) (*SyncResult, error)

	// DeleteNote deletes a note from the external system
	DeleteNote(ctx context.Context, integration *Integration, externalID string) error

	// ValidateIntegration validates the integration configuration
	ValidateIntegration(ctx context.Context, integration *Integration) error
}

// ExternalNote represents a note from an external system
type ExternalNote struct {
	ExternalID string                 `json:"external_id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Category   string                 `json:"category"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NoteIntegration represents the relationship between a note and an integration
type NoteIntegration struct {
	NoteIntegrationID int        `json:"note_integration_id" db:"note_integration_id"`
	NoteID            uuid.UUID  `json:"note_id" db:"note_id"`
	IntegrationID     uuid.UUID  `json:"integration_id" db:"integration_id"`
	ExternalID        string     `json:"external_id" db:"external_id"`
	SyncStatus        string     `json:"sync_status" db:"sync_status"` // 'synced', 'pending', 'error', 'conflict'
	SyncedAt          time.Time  `json:"synced_at" db:"synced_at"`
	RetryCount        int        `json:"retry_count" db:"retry_count"`
	LastError         *string    `json:"last_error,omitempty" db:"last_error"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// SyncService defines the interface for sync service operations
type SyncService interface {
	// RegisterProvider registers a new sync provider
	RegisterProvider(provider SyncProvider)

	// GetProvider returns a sync provider by name
	GetProvider(providerName string) (SyncProvider, error)

	// SyncNoteToIntegrations synchronizes a note to all active integrations for a user
	SyncNoteToIntegrations(ctx context.Context, note *Note) error

	// SyncFromIntegrations synchronizes notes from external systems to internal
	SyncFromIntegrations(ctx context.Context, chatID string) error

	// ResyncFailedNotes retries failed sync operations
	ResyncFailedNotes(ctx context.Context) error

	// GetSyncStatus returns the sync status for a note
	GetSyncStatus(ctx context.Context, noteID uuid.UUID) ([]*NoteIntegration, error)
}
