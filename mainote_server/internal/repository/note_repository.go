package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"mainote-server/internal/domain"
)

// NoteRepository defines the interface for note persistence.
type NoteRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, note *domain.Note) error
	GetByID(ctx context.Context, noteID uuid.UUID) (*domain.Note, error)
	Update(ctx context.Context, note *domain.Note) error
	Delete(ctx context.Context, noteID uuid.UUID) error

	// User-scoped operations
	GetByIDForUser(ctx context.Context, noteID uuid.UUID, chatID string) (*domain.Note, error)
	GetNotesForUser(ctx context.Context, chatID string, category, status *string, limit, offset int) (*domain.NotesListResult, error)
	UpdateForUser(ctx context.Context, noteID uuid.UUID, chatID string, note *domain.Note) error
	DeleteForUser(ctx context.Context, noteID uuid.UUID, chatID string) error
}

// NewNoteRepository creates a new instance of NoteRepository.
func NewNoteRepository(db *sqlx.DB) NoteRepository {
	return &noteRepository{db: db}
}

type noteRepository struct {
	db *sqlx.DB
}

// Create creates a new note in the database.
func (r *noteRepository) Create(ctx context.Context, note *domain.Note) error {
	query := `
		INSERT INTO note (note_id, chat_id, user_id, title, content, category, status, source, voice_file_id, transcription, due_at, effort_min, priority, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.db.ExecContext(ctx, query,
		note.NoteID, note.ChatID, note.UserID, note.Title, note.Content,
		note.Category, note.Status, note.Source, note.VoiceFileID,
		note.Transcription, note.DueAt, note.EffortMin, note.Priority, note.Metadata, note.CreatedAt, note.UpdatedAt)

	return err
}

// GetByID retrieves a note by its ID.
func (r *noteRepository) GetByID(ctx context.Context, noteID uuid.UUID) (*domain.Note, error) {
	var note domain.Note
	query := `
		SELECT note_id, chat_id, user_id, title, content, category, status, source, 
		       voice_file_id, transcription, due_at, effort_min, priority, metadata, created_at, updated_at, deleted_at 
		FROM note 
		WHERE note_id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &note, query, noteID)
	if err != nil {
		return nil, err
	}

	return &note, nil
}

// GetByIDForUser retrieves a note by its ID for a specific user (scoped by chat_id).
func (r *noteRepository) GetByIDForUser(ctx context.Context, noteID uuid.UUID, chatID string) (*domain.Note, error) {
	var note domain.Note
	query := `
		SELECT note_id, chat_id, user_id, title, content, category, status, source, 
		       voice_file_id, transcription, due_at, effort_min, priority, metadata, created_at, updated_at, deleted_at 
		FROM note 
		WHERE note_id = $1 AND chat_id = $2 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &note, query, noteID, chatID)
	if err != nil {
		log.Printf("Query failed with error: %v", err)
		return nil, err
	}

	return &note, nil
}

// GetNotesForUser retrieves notes for a specific user with optional filtering and pagination.
func (r *noteRepository) GetNotesForUser(ctx context.Context, chatID string, category, status *string, limit, offset int) (*domain.NotesListResult, error) {
	// Build the WHERE clause
	whereConditions := []string{"chat_id = $1", "deleted_at IS NULL"}
	args := []interface{}{chatID}
	argIndex := 2

	if category != nil && *category != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *category)
		argIndex++
	}

	if status != nil && *status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *status)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total records for pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM note WHERE %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Get the actual notes with pagination
	notesQuery := fmt.Sprintf(`
		SELECT note_id, chat_id, user_id, title, content, category, status, source, 
		       voice_file_id, transcription, due_at, effort_min, priority, metadata, created_at, updated_at, deleted_at 
		FROM note 
		WHERE %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	var notes []domain.Note
	err = r.db.SelectContext(ctx, &notes, notesQuery, args...)
	if err != nil {
		return nil, err
	}

	hasMore := offset+len(notes) < total

	return &domain.NotesListResult{
		Notes:   notes,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

// Update updates an existing note.
func (r *noteRepository) Update(ctx context.Context, note *domain.Note) error {
	query := `
		UPDATE note 
		SET title = $2, content = $3, category = $4, status = $5, voice_file_id = $6, 
		    transcription = $7, due_at = $8, effort_min = $9, priority = $10, metadata = $11, updated_at = $12 
		WHERE note_id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		note.NoteID, note.Title, note.Content, note.Category, note.Status,
		note.VoiceFileID, note.Transcription, note.DueAt, note.EffortMin, note.Priority, note.Metadata, note.UpdatedAt)

	return err
}

// UpdateForUser updates an existing note for a specific user (scoped by chat_id).
func (r *noteRepository) UpdateForUser(ctx context.Context, noteID uuid.UUID, chatID string, note *domain.Note) error {
	query := `
		UPDATE note 
		SET title = $3, content = $4, category = $5, status = $6, voice_file_id = $7, 
		    transcription = $8, due_at = $9, effort_min = $10, priority = $11, metadata = $12, updated_at = $13 
		WHERE note_id = $1 AND chat_id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		noteID, chatID, note.Title, note.Content, note.Category, note.Status,
		note.VoiceFileID, note.Transcription, note.DueAt, note.EffortMin, note.Priority, note.Metadata, note.UpdatedAt)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete soft deletes a note.
func (r *noteRepository) Delete(ctx context.Context, noteID uuid.UUID) error {
	query := `UPDATE note SET deleted_at = $2, updated_at = $2 WHERE note_id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, noteID, time.Now())
	return err
}

// DeleteForUser soft deletes a note for a specific user (scoped by chat_id).
func (r *noteRepository) DeleteForUser(ctx context.Context, noteID uuid.UUID, chatID string) error {
	query := `UPDATE note SET deleted_at = $3, updated_at = $3 WHERE note_id = $1 AND chat_id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, noteID, chatID, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
