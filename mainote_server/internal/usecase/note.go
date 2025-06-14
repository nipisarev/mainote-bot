package usecase

import (
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
	// TODO: Implement note creation logic
	// This is a stub implementation
	return nil, nil
}

// GetNote retrieves a note by ID
func (u *NoteUsecase) GetNote(id string) (*domain.Note, error) {
	// TODO: Implement note retrieval logic
	// This is a stub implementation
	return u.noteRepo.GetByID(id)
}

// UpdateNote updates an existing note
func (u *NoteUsecase) UpdateNote(id string, req api.UpdateNoteRequest) (*domain.Note, error) {
	// TODO: Implement note update logic
	// This is a stub implementation
	return nil, nil
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
