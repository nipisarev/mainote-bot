package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"mainote-server/internal/domain"
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByChatID(ctx context.Context, chatID string) (*domain.UserWithSettings, error)
	UpdateChatID(ctx context.Context, userID uuid.UUID, chatID string) error
	CreateUserSettings(ctx context.Context, userID uuid.UUID, chatID string, morningNotificationTime, timezone *string) (*domain.UserSettings, error)
	UpdateUserSettings(ctx context.Context, update *domain.UserSettings) (*domain.UserSettings, error)
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

type userRepository struct {
	db *sqlx.DB
}

// Create creates a new user in the database.
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, email, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, email)
	return &user, err
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, email, password, created_at, updated_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, id)
	return &user, err
}

func (r *userRepository) FindByChatID(ctx context.Context, chatID string) (*domain.UserWithSettings, error) {
	query := `
		SELECT 
			u.id, u.email, u.password, u.created_at, u.updated_at, u.deleted_at,
			us.user_settings_id, us.user_id, us.chat_id, us.morning_notification_time, us.timezone, 
			us.created_at, us.updated_at, us.deleted_at
		FROM users u 
		INNER JOIN user_settings us ON u.id = us.user_id 
		WHERE us.chat_id = $1 AND u.deleted_at IS NULL AND us.deleted_at IS NULL`

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows // Return the expected error
	}

	var user domain.User
	var settings domain.UserSettings

	err = rows.Scan(
		&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		&settings.UserSettingsID, &settings.UserID, &settings.ChatID, &settings.MorningNotificationTime, &settings.Timezone,
		&settings.CreatedAt, &settings.UpdatedAt, &settings.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &domain.UserWithSettings{
		User:     user,
		Settings: settings,
	}, nil
}

func (r *userRepository) UpdateChatID(ctx context.Context, userID uuid.UUID, chatID string) error {
	query := `UPDATE user_settings SET chat_id = $1, updated_at = $2 WHERE user_id = $3`
	_, err := r.db.ExecContext(ctx, query, chatID, time.Now(), userID)
	return err
}

func (r *userRepository) UpdateUserSettings(ctx context.Context, update *domain.UserSettings) (*domain.UserSettings, error) {
	// Build dynamic query based on provided fields
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if update.MorningNotificationTime != nil {
		setParts = append(setParts, fmt.Sprintf("morning_notification_time = $%d", argIndex))
		args = append(args, *update.MorningNotificationTime)
		argIndex++
	}

	if update.Timezone != nil {
		setParts = append(setParts, fmt.Sprintf("timezone = $%d", argIndex))
		args = append(args, *update.Timezone)
		argIndex++
	}

	// Add chat_id to args for WHERE clause
	args = append(args, update.ChatID)

	if len(setParts) == 1 { // Only updated_at was added, meaning no actual fields to update
		return nil, fmt.Errorf("no fields provided for update")
	}

	query := fmt.Sprintf(`
		UPDATE user_settings 
		SET %s 
		WHERE chat_id = $%d AND deleted_at IS NULL
		RETURNING user_settings_id, user_id, chat_id, morning_notification_time, timezone, created_at, updated_at, deleted_at`,
		strings.Join(setParts, ", "), argIndex)

	var settings domain.UserSettings
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&settings.UserSettingsID,
		&settings.UserID,
		&settings.ChatID,
		&settings.MorningNotificationTime,
		&settings.Timezone,
		&settings.CreatedAt,
		&settings.UpdatedAt,
		&settings.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user settings not found for chat_id: %s", update.ChatID)
		}
		return nil, err
	}

	return &settings, nil
}

// CreateUserSettings creates new user settings in the database.
func (r *userRepository) CreateUserSettings(ctx context.Context, userID uuid.UUID, chatID string, morningNotificationTime, timezone *string) (*domain.UserSettings, error) {
	query := `
		INSERT INTO user_settings (user_id, chat_id, morning_notification_time, timezone, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING user_settings_id, user_id, chat_id, morning_notification_time, timezone, created_at, updated_at, deleted_at`
	
	now := time.Now()
	var settings domain.UserSettings
	
	err := r.db.QueryRowContext(ctx, query, userID, chatID, morningNotificationTime, timezone, now, now).Scan(
		&settings.UserSettingsID,
		&settings.UserID,
		&settings.ChatID,
		&settings.MorningNotificationTime,
		&settings.Timezone,
		&settings.CreatedAt,
		&settings.UpdatedAt,
		&settings.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}
