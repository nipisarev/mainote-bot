package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type UserSettings struct {
	UserSettingsID          int        `json:"user_settings_id" db:"user_settings_id"`
	UserID                  uuid.UUID  `json:"user_id" db:"user_id"`
	ChatID                  string     `json:"chat_id" db:"chat_id"`
	MorningNotificationTime *string    `json:"morning_notification_time,omitempty" db:"morning_notification_time"`
	Timezone                *string    `json:"timezone,omitempty" db:"timezone"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt               *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type UserWithSettings struct {
	User     User         `json:"user"`
	Settings UserSettings `json:"settings"`
}
