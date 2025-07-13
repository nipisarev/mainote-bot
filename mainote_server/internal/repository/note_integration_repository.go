package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"mainote-server/internal/domain"
)

// NoteIntegrationRepository defines the interface for note_integration persistence.
type NoteIntegrationRepository interface {
	Create(ctx context.Context, noteIntegration *domain.NoteIntegration) error
	GetByID(ctx context.Context, noteIntegrationID int) (*domain.NoteIntegration, error)
	GetByNoteID(ctx context.Context, noteID uuid.UUID) ([]*domain.NoteIntegration, error)
	GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]*domain.NoteIntegration, error)
	GetByNoteAndIntegration(ctx context.Context, noteID uuid.UUID, integrationID uuid.UUID) (*domain.NoteIntegration, error)
	Update(ctx context.Context, noteIntegration *domain.NoteIntegration) error
	UpdateSyncStatus(ctx context.Context, noteIntegrationID int, status string, externalID *string, lastError *string) error
	Delete(ctx context.Context, noteIntegrationID int) error
	GetFailedSyncs(ctx context.Context, maxRetries int) ([]*domain.NoteIntegration, error)
	GetPendingSyncs(ctx context.Context) ([]*domain.NoteIntegration, error)
}

// NewNoteIntegrationRepository creates a new instance of NoteIntegrationRepository.
func NewNoteIntegrationRepository(db *sqlx.DB) NoteIntegrationRepository {
	return &noteIntegrationRepository{db: db}
}

type noteIntegrationRepository struct {
	db *sqlx.DB
}

// Create creates a new note_integration record in the database.
func (r *noteIntegrationRepository) Create(ctx context.Context, noteIntegration *domain.NoteIntegration) error {
	query := `
		INSERT INTO note_integration (note_id, integration_id, external_id, sync_status, synced_at, retry_count, last_error, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING note_integration_id`

	err := r.db.QueryRowContext(ctx, query,
		noteIntegration.NoteID, noteIntegration.IntegrationID, noteIntegration.ExternalID,
		noteIntegration.SyncStatus, noteIntegration.SyncedAt, noteIntegration.RetryCount,
		noteIntegration.LastError, noteIntegration.CreatedAt, noteIntegration.UpdatedAt).Scan(&noteIntegration.NoteIntegrationID)

	return err
}

// GetByID retrieves a note_integration by its ID.
func (r *noteIntegrationRepository) GetByID(ctx context.Context, noteIntegrationID int) (*domain.NoteIntegration, error) {
	query := `
		SELECT note_integration_id, note_id, integration_id, external_id, sync_status, synced_at, retry_count, last_error, created_at, updated_at, deleted_at
		FROM note_integration 
		WHERE note_integration_id = $1 AND deleted_at IS NULL`

	var noteIntegration domain.NoteIntegration
	err := r.db.GetContext(ctx, &noteIntegration, query, noteIntegrationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &noteIntegration, nil
}

// GetByNoteID retrieves all note_integration records for a given note ID.
func (r *noteIntegrationRepository) GetByNoteID(ctx context.Context, noteID uuid.UUID) ([]*domain.NoteIntegration, error) {
	query := `
		SELECT note_integration_id, note_id, integration_id, external_id, sync_status, synced_at, retry_count, last_error, created_at, updated_at, deleted_at
		FROM note_integration 
		WHERE note_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var noteIntegrations []*domain.NoteIntegration
	err := r.db.SelectContext(ctx, &noteIntegrations, query, noteID)
	if err != nil {
		return nil, err
	}

	return noteIntegrations, nil
}

// GetByIntegrationID retrieves all note_integration records for a given integration ID.
func (r *noteIntegrationRepository) GetByIntegrationID(ctx context.Context, integrationID uuid.UUID) ([]*domain.NoteIntegration, error) {
	query := `
		SELECT note_integration_id, note_id, integration_id, external_id, sync_status, synced_at, retry_count, last_error, created_at, updated_at, deleted_at
		FROM note_integration 
		WHERE integration_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var noteIntegrations []*domain.NoteIntegration
	err := r.db.SelectContext(ctx, &noteIntegrations, query, integrationID)
	if err != nil {
		return nil, err
	}

	return noteIntegrations, nil
}

// GetByNoteAndIntegration retrieves a note_integration record by note and integration IDs.
func (r *noteIntegrationRepository) GetByNoteAndIntegration(ctx context.Context, noteID uuid.UUID, integrationID uuid.UUID) (*domain.NoteIntegration, error) {
	query := `
		SELECT note_integration_id, note_id, integration_id, external_id, sync_status, synced_at, retry_count, last_error, created_at, updated_at, deleted_at
		FROM note_integration 
		WHERE note_id = $1 AND integration_id = $2 AND deleted_at IS NULL`

	var noteIntegration domain.NoteIntegration
	err := r.db.GetContext(ctx, &noteIntegration, query, noteID, integrationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &noteIntegration, nil
}

// Update updates an existing note_integration record.
func (r *noteIntegrationRepository) Update(ctx context.Context, noteIntegration *domain.NoteIntegration) error {
	query := `
		UPDATE note_integration 
		SET external_id = $2, sync_status = $3, synced_at = $4, retry_count = $5, last_error = $6, updated_at = $7
		WHERE note_integration_id = $1 AND deleted_at IS NULL`

	noteIntegration.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		noteIntegration.NoteIntegrationID, noteIntegration.ExternalID, noteIntegration.SyncStatus,
		noteIntegration.SyncedAt, noteIntegration.RetryCount, noteIntegration.LastError,
		noteIntegration.UpdatedAt)

	return err
}

// UpdateSyncStatus updates the sync status of a note_integration record.
func (r *noteIntegrationRepository) UpdateSyncStatus(ctx context.Context, noteIntegrationID int, status string, externalID *string, lastError *string) error {
	query := `
		UPDATE note_integration 
		SET sync_status = $2, synced_at = $3, last_error = $4, updated_at = $5`

	args := []interface{}{noteIntegrationID, status, time.Now(), lastError, time.Now()}

	if externalID != nil {
		query += `, external_id = $6`
		args = append(args, *externalID)
	}

	query += ` WHERE note_integration_id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// Delete soft deletes a note_integration record.
func (r *noteIntegrationRepository) Delete(ctx context.Context, noteIntegrationID int) error {
	query := `
		UPDATE note_integration 
		SET deleted_at = $2, updated_at = $2
		WHERE note_integration_id = $1 AND deleted_at IS NULL`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, noteIntegrationID, now)
	return err
}

// GetFailedSyncs retrieves note_integration records that failed to sync and haven't exceeded max retries.
func (r *noteIntegrationRepository) GetFailedSyncs(ctx context.Context, maxRetries int) ([]*domain.NoteIntegration, error) {
	query := `
		SELECT note_integration_id, note_id, integration_id, external_id, sync_status, synced_at, retry_count, last_error, created_at, updated_at, deleted_at
		FROM note_integration 
		WHERE sync_status = 'error' AND retry_count < $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`

	var noteIntegrations []*domain.NoteIntegration
	err := r.db.SelectContext(ctx, &noteIntegrations, query, maxRetries)
	if err != nil {
		return nil, err
	}

	return noteIntegrations, nil
}

// GetPendingSyncs retrieves note_integration records that are pending sync.
func (r *noteIntegrationRepository) GetPendingSyncs(ctx context.Context) ([]*domain.NoteIntegration, error) {
	query := `
		SELECT note_integration_id, note_id, integration_id, external_id, sync_status, synced_at, retry_count, last_error, created_at, updated_at, deleted_at
		FROM note_integration 
		WHERE sync_status = 'pending' AND deleted_at IS NULL
		ORDER BY created_at ASC`

	var noteIntegrations []*domain.NoteIntegration
	err := r.db.SelectContext(ctx, &noteIntegrations, query)
	if err != nil {
		return nil, err
	}

	return noteIntegrations, nil
}
