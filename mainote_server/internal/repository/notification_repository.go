package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"mainote-server/internal/domain"
)

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.ScheduledNotification) error
	GetPendingNotifications(ctx context.Context, cutoffTime time.Time) ([]domain.ScheduledNotification, error)
	MarkAsDelivered(ctx context.Context, id uuid.UUID) error
}

// NewNotificationRepository creates a new instance of NotificationRepository
func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

type notificationRepository struct {
	db *sqlx.DB
}

// Create creates a new scheduled notification in the database
func (r *notificationRepository) Create(ctx context.Context, notification *domain.ScheduledNotification) error {
	query := `
		INSERT INTO scheduled_notifications (id, user_id, message, send_at, delivered, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Message,
		notification.SendAt,
		notification.Delivered,
		notification.CreatedAt,
		notification.UpdatedAt,
	)
	return err
}

// GetPendingNotifications retrieves all notifications that should be sent before cutoffTime
func (r *notificationRepository) GetPendingNotifications(ctx context.Context, cutoffTime time.Time) ([]domain.ScheduledNotification, error) {
	query := `
		SELECT id, user_id, message, send_at, delivered, created_at, updated_at
		FROM scheduled_notifications
		WHERE send_at <= $1 AND delivered = false
		ORDER BY send_at ASC`

	var notifications []domain.ScheduledNotification
	err := r.db.SelectContext(ctx, &notifications, query, cutoffTime)
	return notifications, err
}

// MarkAsDelivered marks a notification as delivered
func (r *notificationRepository) MarkAsDelivered(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE scheduled_notifications
		SET delivered = true, updated_at = $1
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
