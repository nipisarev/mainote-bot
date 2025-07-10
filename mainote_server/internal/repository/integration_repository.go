package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"mainote-server/internal/domain"
)

// IntegrationRepository defines the interface for integration persistence.
type IntegrationRepository interface {
	Create(ctx context.Context, integration *domain.Integration) error
	GetByID(ctx context.Context, integrationID uuid.UUID) (*domain.Integration, error)
	GetByIDForUser(ctx context.Context, integrationID uuid.UUID, chatID string) (*domain.IntegrationWithApp, error)
	GetIntegrationsForUser(ctx context.Context, chatID string, status *string, appID *uuid.UUID, limit, offset int) (*domain.IntegrationsListResult, error)
	Update(ctx context.Context, integration *domain.Integration) error
	UpdateForUser(ctx context.Context, integrationID uuid.UUID, chatID string, integration *domain.Integration) error
	DeleteForUser(ctx context.Context, integrationID uuid.UUID, chatID string) error
}

// NewIntegrationRepository creates a new instance of IntegrationRepository.
func NewIntegrationRepository(db *sqlx.DB) IntegrationRepository {
	return &integrationRepository{db: db}
}

type integrationRepository struct {
	db *sqlx.DB
}

// Create creates a new integration in the database.
func (r *integrationRepository) Create(ctx context.Context, integration *domain.Integration) error {
	query := `
		INSERT INTO integration (integration_id, app_id, chat_id, status, config, auth_type, auth_data, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		integration.IntegrationID, integration.AppID, integration.ChatID, integration.Status,
		integration.Config, integration.AuthType, integration.AuthData,
		integration.CreatedAt, integration.UpdatedAt)

	return err
}

// GetByID retrieves an integration by its ID.
func (r *integrationRepository) GetByID(ctx context.Context, integrationID uuid.UUID) (*domain.Integration, error) {
	var integration domain.Integration
	query := `
		SELECT integration_id, app_id, chat_id, status, config, auth_type, auth_data, created_at, updated_at, deleted_at 
		FROM integration 
		WHERE integration_id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &integration, query, integrationID)
	if err != nil {
		return nil, err
	}

	return &integration, nil
}

// GetByIDForUser retrieves an integration by ID for a specific user (scoped by chat_id).
func (r *integrationRepository) GetByIDForUser(ctx context.Context, integrationID uuid.UUID, chatID string) (*domain.IntegrationWithApp, error) {
	query := `
		SELECT i.integration_id, i.app_id, i.chat_id, i.status, i.config, i.auth_type, i.auth_data, 
		       i.created_at, i.updated_at, i.deleted_at,
		       a.provider, a.name, a.description, a.created_at as app_created_at, a.updated_at as app_updated_at, a.deleted_at as app_deleted_at
		FROM integration i
		JOIN app a ON i.app_id = a.app_id
		WHERE i.integration_id = $1 AND i.chat_id = $2 AND i.deleted_at IS NULL AND a.deleted_at IS NULL`

	var result struct {
		IntegrationID uuid.UUID        `db:"integration_id"`
		AppID         uuid.UUID        `db:"app_id"`
		ChatID        string           `db:"chat_id"`
		Status        string           `db:"status"`
		Config        *json.RawMessage `db:"config"`
		AuthType      string           `db:"auth_type"`
		AuthData      *json.RawMessage `db:"auth_data"`
		CreatedAt     time.Time        `db:"created_at"`
		UpdatedAt     time.Time        `db:"updated_at"`
		DeletedAt     *time.Time       `db:"deleted_at"`
		Provider      string           `db:"provider"`
		Name          string           `db:"name"`
		Description   *string          `db:"description"`
		AppCreatedAt  time.Time        `db:"app_created_at"`
		AppUpdatedAt  time.Time        `db:"app_updated_at"`
		AppDeletedAt  *time.Time       `db:"app_deleted_at"`
	}

	err := r.db.GetContext(ctx, &result, query, integrationID, chatID)
	if err != nil {
		return nil, err
	}

	return &domain.IntegrationWithApp{
		Integration: domain.Integration{
			IntegrationID: result.IntegrationID,
			AppID:         result.AppID,
			ChatID:        result.ChatID,
			Status:        result.Status,
			Config:        result.Config,
			AuthType:      result.AuthType,
			AuthData:      result.AuthData,
			CreatedAt:     result.CreatedAt,
			UpdatedAt:     result.UpdatedAt,
			DeletedAt:     result.DeletedAt,
		},
		App: domain.App{
			AppID:       result.AppID,
			Provider:    result.Provider,
			Name:        result.Name,
			Description: result.Description,
			CreatedAt:   result.AppCreatedAt,
			UpdatedAt:   result.AppUpdatedAt,
			DeletedAt:   result.AppDeletedAt,
		},
	}, nil
}

// GetIntegrationsForUser retrieves integrations for a specific user with optional filtering and pagination.
func (r *integrationRepository) GetIntegrationsForUser(ctx context.Context, chatID string, status *string, appID *uuid.UUID, limit, offset int) (*domain.IntegrationsListResult, error) {
	// Build the WHERE clause
	whereConditions := []string{"i.chat_id = $1", "i.deleted_at IS NULL", "a.deleted_at IS NULL"}
	args := []interface{}{chatID}
	argIndex := 2

	if status != nil && *status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("i.status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}

	if appID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("i.app_id = $%d", argIndex))
		args = append(args, *appID)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records for pagination
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM integration i 
		JOIN app a ON i.app_id = a.app_id 
		WHERE %s`, whereClause)

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Get the actual integrations with pagination
	integrationsQuery := fmt.Sprintf(`
		SELECT i.integration_id, i.app_id, i.chat_id, i.status, i.config, i.auth_type, i.auth_data, 
		       i.created_at, i.updated_at, i.deleted_at,
		       a.provider, a.name, a.description, a.created_at as app_created_at, a.updated_at as app_updated_at, a.deleted_at as app_deleted_at
		FROM integration i
		JOIN app a ON i.app_id = a.app_id
		WHERE %s 
		ORDER BY i.created_at DESC 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, integrationsQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []domain.IntegrationWithApp
	for rows.Next() {
		var result struct {
			IntegrationID uuid.UUID        `db:"integration_id"`
			AppID         uuid.UUID        `db:"app_id"`
			ChatID        string           `db:"chat_id"`
			Status        string           `db:"status"`
			Config        *json.RawMessage `db:"config"`
			AuthType      string           `db:"auth_type"`
			AuthData      *json.RawMessage `db:"auth_data"`
			CreatedAt     time.Time        `db:"created_at"`
			UpdatedAt     time.Time        `db:"updated_at"`
			DeletedAt     *time.Time       `db:"deleted_at"`
			Provider      string           `db:"provider"`
			Name          string           `db:"name"`
			Description   *string          `db:"description"`
			AppCreatedAt  time.Time        `db:"app_created_at"`
			AppUpdatedAt  time.Time        `db:"app_updated_at"`
			AppDeletedAt  *time.Time       `db:"app_deleted_at"`
		}

		err := rows.Scan(
			&result.IntegrationID, &result.AppID, &result.ChatID, &result.Status,
			&result.Config, &result.AuthType, &result.AuthData,
			&result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
			&result.Provider, &result.Name, &result.Description,
			&result.AppCreatedAt, &result.AppUpdatedAt, &result.AppDeletedAt,
		)
		if err != nil {
			return nil, err
		}

		integration := domain.IntegrationWithApp{
			Integration: domain.Integration{
				IntegrationID: result.IntegrationID,
				AppID:         result.AppID,
				ChatID:        result.ChatID,
				Status:        result.Status,
				Config:        result.Config,
				AuthType:      result.AuthType,
				AuthData:      result.AuthData,
				CreatedAt:     result.CreatedAt,
				UpdatedAt:     result.UpdatedAt,
				DeletedAt:     result.DeletedAt,
			},
			App: domain.App{
				AppID:       result.AppID,
				Provider:    result.Provider,
				Name:        result.Name,
				Description: result.Description,
				CreatedAt:   result.AppCreatedAt,
				UpdatedAt:   result.AppUpdatedAt,
				DeletedAt:   result.AppDeletedAt,
			},
		}

		integrations = append(integrations, integration)
	}

	hasMore := offset+len(integrations) < total

	return &domain.IntegrationsListResult{
		Integrations: integrations,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
		HasMore:      hasMore,
	}, nil
}

// Update updates an existing integration.
func (r *integrationRepository) Update(ctx context.Context, integration *domain.Integration) error {
	query := `
		UPDATE integration 
		SET status = $2, config = $3, auth_type = $4, auth_data = $5, updated_at = $6 
		WHERE integration_id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		integration.IntegrationID, integration.Status, integration.Config,
		integration.AuthType, integration.AuthData, integration.UpdatedAt)

	return err
}

// UpdateForUser updates an existing integration for a specific user (scoped by chat_id).
func (r *integrationRepository) UpdateForUser(ctx context.Context, integrationID uuid.UUID, chatID string, integration *domain.Integration) error {
	query := `
		UPDATE integration 
		SET status = $3, config = $4, auth_type = $5, auth_data = $6, updated_at = $7 
		WHERE integration_id = $1 AND chat_id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		integrationID, chatID, integration.Status, integration.Config,
		integration.AuthType, integration.AuthData, integration.UpdatedAt)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteForUser soft deletes an integration for a specific user (scoped by chat_id).
func (r *integrationRepository) DeleteForUser(ctx context.Context, integrationID uuid.UUID, chatID string) error {
	query := `UPDATE integration SET deleted_at = $3, updated_at = $3 WHERE integration_id = $1 AND chat_id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, integrationID, chatID, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
