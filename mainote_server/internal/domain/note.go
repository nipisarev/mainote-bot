package domain

import (
	api "mainote-backend/pkg/generated"
)

// Note represents the domain model for a note
type Note struct {
	ID        string                 `json:"id"`
	ChatID    string                 `json:"chat_id"`
	Content   string                 `json:"content"`
	Title     *string                `json:"title,omitempty"`
	Category  api.NoteCategory       `json:"category"`
	Status    api.NoteStatus         `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NoteRepository defines the interface for note data access
type NoteRepository interface {
	Create(note *Note) error
	GetByID(id string) (*Note, error)
	GetByChatID(chatID string, limit, offset int) ([]*Note, int, error)
	Update(note *Note) error
	Delete(id string) error
}
