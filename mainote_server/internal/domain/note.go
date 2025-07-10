package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Note represents a note entity in the domain
type Note struct {
	NoteID        uuid.UUID        `json:"note_id" db:"note_id"`
	ChatID        string           `json:"chat_id" db:"chat_id"`
	UserID        uuid.UUID        `json:"user_id" db:"user_id"`
	Title         *string          `json:"title,omitempty" db:"title"`
	Content       string           `json:"content" db:"content"`
	Category      string           `json:"category" db:"category"`
	Status        string           `json:"status" db:"status"`
	Source        string           `json:"source" db:"source"`
	VoiceFileID   *string          `json:"voice_file_id,omitempty" db:"voice_file_id"`
	Transcription *string          `json:"transcription,omitempty" db:"transcription"`
	Metadata      *json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time       `json:"deleted_at,omitempty" db:"deleted_at"`
}

// NotesListResult represents paginated notes result
type NotesListResult struct {
	Notes   []Note `json:"notes"`
	Total   int    `json:"total"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	HasMore bool   `json:"has_more"`
}
