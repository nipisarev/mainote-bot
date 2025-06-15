package usecase

import (
	"fmt"

	"mainote-backend/internal/domain"
	api "mainote-backend/pkg/generated/api"
)

// NoteUsecase handles business logic for notes
type NoteUsecase struct {
	noteRepo domain.NoteRepository
}

// NewNoteUsecase creates a new note usecase
func NewNoteUsecase(noteRepo domain.NoteRepository) *NoteUsecase {
	return &NoteUsecase{
		noteRepo: noteRepo,
	}
}

// CreateNote creates a new note
func (u *NoteUsecase) CreateNote(req api.CreateNoteRequest) (*domain.Note, error) {
	// Create domain note from API request
	note := &domain.Note{
		ChatID:   req.ChatId,
		Content:  req.Content,
		Title:    req.Title,
		Category: req.Category,
		Status:   api.NOTESTATUS_ACTIVE, // Default status for new notes
		Metadata: make(map[string]interface{}),
	}

	// Add source information if provided
	if req.Source != "" {
		note.Metadata["source"] = req.Source
	} else {
		note.Metadata["source"] = "telegram_bot"
	}

	// Add any additional metadata from request
	if req.Metadata != nil {
		for key, value := range *req.Metadata {
			note.Metadata[key] = value
		}
	}

	// Add creation context metadata
	note.Metadata["created_via"] = "api"
	note.Metadata["api_version"] = "v1"

	// Validate business rules
	if note.Content == "" {
		return nil, fmt.Errorf("note content cannot be empty")
	}

	if note.ChatID == "" {
		return nil, fmt.Errorf("chat ID cannot be empty")
	}

	// Create note in repository
	err := u.noteRepo.Create(note)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

// GetNote retrieves a note by ID
func (u *NoteUsecase) GetNote(id string) (*domain.Note, error) {
	// TODO: Implement note retrieval logic
	// This is a stub implementation
	return u.noteRepo.GetByID(id)
}

// UpdateNote updates an existing note
func (u *NoteUsecase) UpdateNote(id string, req api.UpdateNoteRequest) (*domain.Note, error) {
	// First get the existing note
	existingNote, err := u.noteRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	// Update only the fields that are provided in the request
	if req.Content != "" {
		existingNote.Content = req.Content
	}

	if req.Title != nil {
		existingNote.Title = req.Title
	}

	if req.Category != "" {
		existingNote.Category = req.Category
	}

	if req.Status != "" {
		existingNote.Status = req.Status
	}

	if req.Metadata != nil {
		// Merge metadata instead of replacing completely
		if existingNote.Metadata == nil {
			existingNote.Metadata = make(map[string]interface{})
		}
		for key, value := range *req.Metadata {
			existingNote.Metadata[key] = value
		}
	}

	// Add update context metadata
	existingNote.Metadata["updated_via"] = "api"
	existingNote.Metadata["api_version"] = "v1"

	// Update in repository
	err = u.noteRepo.Update(existingNote)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return existingNote, nil
}

// DeleteNote deletes a note by ID
func (u *NoteUsecase) DeleteNote(id string) error {
	// TODO: Implement note deletion logic
	// This is a stub implementation
	return u.noteRepo.Delete(id)
}

// ListNotes retrieves notes with pagination
func (u *NoteUsecase) ListNotes(chatID string, limit, offset int) ([]*domain.Note, int, error) {
	// TODO: Implement note listing logic
	// This is a stub implementation
	return u.noteRepo.GetByChatID(chatID, limit, offset)
}
