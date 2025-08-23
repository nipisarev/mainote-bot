package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
func NewNoteUsecase(noteRepo repository.NoteRepository, userRepo repository.UserRepository, syncService domain.SyncService, aiUsecase domain.AIUsecase) NoteUsecase {
	return &noteUsecase{
		noteRepo:    noteRepo,
		userRepo:    userRepo,
		syncService: syncService,
		aiUsecase:   aiUsecase,
	}
}

type noteUsecase struct {
	noteRepo    repository.NoteRepository
	userRepo    repository.UserRepository
	syncService domain.SyncService
	aiUsecase   domain.AIUsecase
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
		DueAt:         nil,
		EffortMin:     30,
		Priority:      0,
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

	// Asynchronously generate AI title if not provided and AI usecase is available
	if uc.aiUsecase != nil {
		log.Printf("Starting AI title generation for note %s (no title provided)", note.NoteID)
		go func() {
			// Create a new context for the AI operation to avoid cancellation issues
			aiCtx := context.Background()

			log.Printf("Generating AI title for note %s with content: %s", note.NoteID, note.Content)
			// Generate title using AI
			generatedTitle, err := uc.aiUsecase.GenerateNoteTitle(aiCtx, note)
			if err != nil {
				if strings.Contains(err.Error(), "quota") {
					log.Printf("OpenAI quota exceeded for note %s - please check billing", note.NoteID)
				} else {
					log.Printf("Failed to generate AI title for note %s: %v", note.NoteID, err)
				}
				return
			}

			if generatedTitle != "" {
				log.Printf("AI generated title for note %s: '%s'", note.NoteID, generatedTitle)
				// Update the note with the generated title
				_, err := uc.UpdateNote(aiCtx, note.NoteID, note.ChatID, &generatedTitle, nil, nil, nil, nil, nil, nil)
				if err != nil {
					log.Printf("Failed to update note %s with AI-generated title: %v", note.NoteID, err)
				} else {
					log.Printf("Successfully generated and updated AI title for note %s: %s", note.NoteID, generatedTitle)
				}
			}
		}()
	} else if uc.aiUsecase == nil {
		log.Printf("AI usecase is nil for note %s, skipping AI title generation", note.NoteID)
	}

	return note, nil
}

// GetNoteByID retrieves a specific note for a user.
func (uc *noteUsecase) GetNoteByID(ctx context.Context, noteID uuid.UUID, chatID string) (*domain.Note, error) {
	log.Printf("GetNoteByID called with noteID: %s, chatID: %s", noteID.String(), chatID)

	// Ensure user exists first
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		log.Printf("User lookup failed for chatID %s: %v", chatID, err)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for chat_id: %s", chatID)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	note, err := uc.noteRepo.GetByIDForUser(ctx, noteID, chatID)
	if err != nil {
		log.Printf("Note lookup failed for noteID %s, chatID %s: %v", noteID.String(), chatID, err)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	log.Printf("Successfully found note: %s", note.NoteID.String())

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

	// Check for optional extended fields in metadata payload and map if present
	if metadata != nil {
		if m, ok := metadata.(map[string]interface{}); ok {
			if v, ok := m["due_at"].(string); ok && v != "" {
				if ts, err := time.Parse(time.RFC3339, v); err == nil {
					updatedNote.DueAt = &ts
				}
			}
			if v, ok := m["effort_min"].(float64); ok {
				updatedNote.EffortMin = int(v)
			}
			if v, ok := m["priority"].(float64); ok {
				updatedNote.Priority = int(v)
			}
		}
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
