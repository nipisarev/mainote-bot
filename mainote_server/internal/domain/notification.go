package domain

import (
	"time"

	"github.com/google/uuid"
)

// ScheduledNotification represents a notification that will be sent at a specific time
type ScheduledNotification struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Message   string    `json:"message" db:"message"`
	SendAt    time.Time `json:"send_at" db:"send_at"`
	Delivered bool      `json:"delivered" db:"delivered"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationPayload represents the payload sent to mainote_bot
type NotificationPayload struct {
	ChatID   string         `json:"chat_id"`
	Type     string         `json:"type"`
	Message  string         `json:"message"`
	ID       string         `json:"id"`
	NotesMap map[int]string `json:"notes_map,omitempty"`
}
