package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"mainote-backend/internal/domain"
	api "mainote-backend/pkg/generated/api"
)

// noteRepository implements domain.NoteRepository interface
type noteRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

// NewNoteRepository creates a new note repository instance
func NewNoteRepository(db *sql.DB) domain.NoteRepository {
	return &noteRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Create inserts a new note into the database
func (r *noteRepository) Create(note *domain.Note) error {
	// Generate UUID for the note if not provided
	if note.ID == "" {
		note.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now().UTC()
	note.CreatedAt = now
	note.UpdatedAt = now

	// Marshal metadata to JSON
	metadataJSON, err := marshalMetadata(note.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Get source from metadata or use default
	source := "telegram_bot" // Default source
	if sourceValue, exists := note.Metadata["source"]; exists {
		if sourceStr, ok := sourceValue.(string); ok {
			source = sourceStr
		}
	}

	// Build SQL query using Squirrel
	query := r.sb.Insert("notes").
		Columns(
			"uuid_id",
			"chat_id",
			"content",
			"title",
			"category",
			"status",
			"source",
			"metadata",
			"created_at",
			"updated_at",
		).
		Values(
			note.ID,
			note.ChatID,
			note.Content,
			note.Title,
			string(note.Category),
			string(note.Status),
			source,
			metadataJSON,
			note.CreatedAt,
			note.UpdatedAt,
		).
		Suffix("RETURNING uuid_id, created_at, updated_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	// Execute query
	row := r.db.QueryRow(sqlStr, args...)
	err = row.Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

// GetByID retrieves a note by its ID
func (r *noteRepository) GetByID(id string) (*domain.Note, error) {
	query := r.sb.Select(
		"uuid_id",
		"chat_id",
		"content",
		"title",
		"category",
		"status",
		"source",
		"metadata",
		"created_at",
		"updated_at",
	).
		From("notes").
		Where(squirrel.Eq{"uuid_id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	note := &domain.Note{}
	var category, status, source string
	var metadataJSON []byte

	row := r.db.QueryRow(sqlStr, args...)
	err = row.Scan(
		&note.ID,
		&note.ChatID,
		&note.Content,
		&note.Title,
		&category,
		&status,
		&source,
		&metadataJSON,
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("note with id %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Unmarshal metadata from JSON
	note.Metadata, err = unmarshalMetadata(metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Convert string fields to API types
	note.Category = api.NoteCategory(category)
	note.Status = api.NoteStatus(status)

	return note, nil
}

// GetByChatID retrieves notes by chat ID with pagination
func (r *noteRepository) GetByChatID(chatID string, limit, offset int) ([]*domain.Note, int, error) {
	// First get total count
	countQuery := r.sb.Select("COUNT(*)").
		From("notes").
		Where(squirrel.Eq{"chat_id": chatID})

	countSqlStr, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int
	err = r.db.QueryRow(countSqlStr, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get notes count: %w", err)
	}

	// Then get the notes
	query := r.sb.Select(
		"uuid_id",
		"chat_id",
		"content",
		"title",
		"category",
		"status",
		"source",
		"metadata",
		"created_at",
		"updated_at",
	).
		From("notes").
		Where(squirrel.Eq{"chat_id": chatID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build select query: %w", err)
	}

	rows, err := r.db.Query(sqlStr, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		note := &domain.Note{}
		var category, status, source string
		var metadataJSON []byte

		err := rows.Scan(
			&note.ID,
			&note.ChatID,
			&note.Content,
			&note.Title,
			&category,
			&status,
			&source,
			&metadataJSON,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan note: %w", err)
		}

		// Unmarshal metadata from JSON
		note.Metadata, err = unmarshalMetadata(metadataJSON)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		// Convert string fields to API types
		note.Category = api.NoteCategory(category)
		note.Status = api.NoteStatus(status)

		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return notes, total, nil
}

// Update updates an existing note
func (r *noteRepository) Update(note *domain.Note) error {
	note.UpdatedAt = time.Now().UTC()

	// Marshal metadata to JSON
	metadataJSON, err := marshalMetadata(note.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := r.sb.Update("notes").
		Set("content", note.Content).
		Set("title", note.Title).
		Set("category", string(note.Category)).
		Set("status", string(note.Status)).
		Set("metadata", metadataJSON).
		Set("updated_at", note.UpdatedAt).
		Where(squirrel.Eq{"uuid_id": note.ID}).
		Suffix("RETURNING updated_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	row := r.db.QueryRow(sqlStr, args...)
	err = row.Scan(&note.UpdatedAt)
	if err == sql.ErrNoRows {
		return fmt.Errorf("note with id %s not found", note.ID)
	}
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	return nil
}

// Delete removes a note from the database
func (r *noteRepository) Delete(id string) error {
	query := r.sb.Delete("notes").
		Where(squirrel.Eq{"uuid_id": id})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.Exec(sqlStr, args...)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("note with id %s not found", id)
	}

	return nil
}

// Helper functions for JSON handling

// marshalMetadata converts metadata map to JSON bytes for database storage
func marshalMetadata(metadata map[string]interface{}) ([]byte, error) {
	if metadata == nil {
		return nil, nil
	}
	return json.Marshal(metadata)
}

// unmarshalMetadata converts JSON bytes from database to metadata map
func unmarshalMetadata(data []byte) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}

	var metadata map[string]interface{}
	err := json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}
