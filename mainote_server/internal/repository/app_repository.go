package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/google/uuid"

	"mainote-server/internal/domain"
)

// AppRepository defines the interface for app persistence.
type AppRepository interface {
	GetAll(ctx context.Context) ([]domain.App, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.App, error)
	GetByProvider(ctx context.Context, provider string) (*domain.App, error)
}

// NewAppRepository creates a new instance of AppRepository.
func NewAppRepository(db *sqlx.DB) AppRepository {
	return &appRepository{db: db}
}

type appRepository struct {
	db *sqlx.DB
}

// GetAll retrieves all available apps for installation.
func (r *appRepository) GetAll(ctx context.Context) ([]domain.App, error) {
	var apps []domain.App
	query := `SELECT app_id, provider, name, description, created_at, updated_at, deleted_at 
	          FROM app 
	          WHERE deleted_at IS NULL 
	          ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &apps, query)
	return apps, err
}

// GetByID finds an app by its ID.
func (r *appRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.App, error) {
	var app domain.App
	query := `SELECT app_id, provider, name, description, created_at, updated_at, deleted_at 
	          FROM app 
	          WHERE app_id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &app, query, id)
	return &app, err
}

// GetByProvider finds an app by its provider name.
func (r *appRepository) GetByProvider(ctx context.Context, provider string) (*domain.App, error) {
	var app domain.App
	query := `SELECT app_id, provider, name, description, created_at, updated_at, deleted_at 
	          FROM app 
	          WHERE provider = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &app, query, provider)
	return &app, err
}
