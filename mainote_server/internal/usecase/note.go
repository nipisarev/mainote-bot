package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"mainote-server/internal/domain"
	"mainote-server/internal/repository"
)

// NoteUsecase defines the interface for note business logic.
type NoteUsecase interface {
	CreateNote(ctx context.Context, chatID, title, content, category, status, source string, voiceFileID, transcription *string, metadata interface{}) (*domain.Note, error)
	GetNoteByID(ctx context.Context, noteID uuid.UUID, chatID string) (*domain.Note, error)
	GetNotesForUser(ctx context.Context, chatID string, category, status *string, limit, offset int) (*domain.NotesListResult, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, chatID string, title, content, category, status *string, voiceFileID, transcription *string, metadata interface{}) (*domain.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID, chatID string) error
}

// NewNoteUsecase creates a new instance of NoteUsecase.
func NewNoteUsecase(noteRepo repository.NoteRepository, userRepo repository.UserRepository, syncService domain.SyncService) NoteUsecase {
	return &noteUsecase{
		noteRepo:    noteRepo,
		userRepo:    userRepo,
		syncService: syncService,
	}
}

type noteUsecase struct {
	noteRepo    repository.NoteRepository
	userRepo    repository.UserRepository
	syncService domain.SyncService
}

// CreateNote creates a new note for a user.
func (uc *noteUsecase) CreateNote(ctx context.Context, chatID, title, content, category, status, source string, voiceFileID, transcription *string, metadata interface{}) (*domain.Note, error) {
	// First, get the user by chat_id to ensure they exist and get user_id
	userWithSettings, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for chat_id: %s", chatID)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Set defaults if not provided
	if category == "" {
		category = "idea"
	}
	if status == "" {
		status = "active"
	}
	if source == "" {
		source = "telegram"
	}

	// Convert metadata to json.RawMessage
	var metadataJSON *json.RawMessage
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata format: %w", err)
		}
		metadataJSON = (*json.RawMessage)(&metadataBytes)
	}

	// Create the note
	note := &domain.Note{
		NoteID:        uuid.New(),
		ChatID:        chatID,
		UserID:        userWithSettings.User.ID,
		Title:         nil,
		Content:       content,
		Category:      category,
		Status:        status,
		Source:        source,
		VoiceFileID:   voiceFileID,
		Transcription: transcription,
		Metadata:      metadataJSON,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Set title if provided
	if title != "" {
		note.Title = &title
	}

	if err := uc.noteRepo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Asynchronously sync the note to active integrations
	go func() {
		// Create a new context for the sync operation to avoid cancellation issues
		syncCtx := context.Background()
		if err := uc.syncService.SyncNoteToIntegrations(syncCtx, note); err != nil {
			log.Printf("Failed to sync note %s to integrations: %v", note.NoteID, err)
		}
	}()

	return note, nil
}

// GetNoteByID retrieves a specific note for a user.
func (uc *noteUsecase) GetNoteByID(ctx context.Context, noteID uuid.UUID, chatID string) (*domain.Note, error) {
	// Ensure user exists first
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for chat_id: %s", chatID)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	note, err := uc.noteRepo.GetByIDForUser(ctx, noteID, chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

// GetNotesForUser retrieves notes for a user with optional filtering and pagination.
func (uc *noteUsecase) GetNotesForUser(ctx context.Context, chatID string, category, status *string, limit, offset int) (*domain.NotesListResult, error) {
	// Ensure user exists first
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for chat_id: %s", chatID)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Validate and set default values for pagination
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	result, err := uc.noteRepo.GetNotesForUser(ctx, chatID, category, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	return result, nil
}

// UpdateNote updates an existing note for a user.
func (uc *noteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, chatID string, title, content, category, status *string, voiceFileID, transcription *string, metadata interface{}) (*domain.Note, error) {
	// Ensure user exists first
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for chat_id: %s", chatID)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Get the existing note to ensure it exists and belongs to the user
	existingNote, err := uc.noteRepo.GetByIDForUser(ctx, noteID, chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to get existing note: %w", err)
	}

	// Update only provided fields
	updatedNote := *existingNote
	updatedNote.UpdatedAt = time.Now()

	if title != nil {
		if *title == "" {
			updatedNote.Title = nil
		} else {
			updatedNote.Title = title
		}
	}
	if content != nil {
		updatedNote.Content = *content
	}
	if category != nil {
		updatedNote.Category = *category
	}
	if status != nil {
		updatedNote.Status = *status
	}
	if voiceFileID != nil {
		if *voiceFileID == "" {
			updatedNote.VoiceFileID = nil
		} else {
			updatedNote.VoiceFileID = voiceFileID
		}
	}
	if transcription != nil {
		if *transcription == "" {
			updatedNote.Transcription = nil
		} else {
			updatedNote.Transcription = transcription
		}
	}
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata format: %w", err)
		}
		updatedNote.Metadata = (*json.RawMessage)(&metadataBytes)
	}

	if err := uc.noteRepo.UpdateForUser(ctx, noteID, chatID, &updatedNote); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return &updatedNote, nil
}

// DeleteNote soft deletes a note for a user.
func (uc *noteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, chatID string) error {
	// Ensure user exists first
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found for chat_id: %s", chatID)
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if err := uc.noteRepo.DeleteForUser(ctx, noteID, chatID); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("note not found")
		}
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}
