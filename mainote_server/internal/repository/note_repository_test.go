package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mainote-server/internal/domain"
)

func TestNoteRepository_Create(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	note := &domain.Note{
		NoteID:    noteID,
		ChatID:    "123456789",
		UserID:    userID,
		Title:     stringPtr("Test Note"),
		Content:   "Test content",
		Category:  "idea",
		Status:    "active",
		Source:    "telegram",
		Metadata:  nil, // Use nil for empty metadata
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set up mock expectation
	mock.ExpectExec(`INSERT INTO note`).
		WithArgs(noteID, "123456789", userID, "Test Note", "Test content", "idea", "active", "telegram", nil, nil, nil, now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	err = repo.Create(context.Background(), note)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_GetByID(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	// Set up mock expectation
	rows := sqlmock.NewRows([]string{
		"note_id", "chat_id", "user_id", "title", "content", "category",
		"status", "source", "voice_file_id", "transcription", "metadata",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(
		noteID, "123456789", userID, "Test Note", "Test content", "idea",
		"active", "telegram", nil, nil, nil,
		now, now, nil,
	)

	mock.ExpectQuery(`SELECT (.+) FROM note WHERE note_id = \$1 AND deleted_at IS NULL`).
		WithArgs(noteID).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetByID(context.Background(), noteID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, "123456789", result.ChatID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "Test Note", *result.Title)
	assert.Equal(t, "Test content", result.Content)
	assert.Equal(t, "idea", result.Category)
	assert.Equal(t, "active", result.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_GetByID_NotFound(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()

	// Set up mock expectation for no rows
	mock.ExpectQuery(`SELECT (.+) FROM note WHERE note_id = \$1 AND deleted_at IS NULL`).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	// Execute
	result, err := repo.GetByID(context.Background(), noteID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_GetByIDForUser(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()
	userID := uuid.New()
	chatID := "123456789"
	now := time.Now()

	// Set up mock expectation
	rows := sqlmock.NewRows([]string{
		"note_id", "chat_id", "user_id", "title", "content", "category",
		"status", "source", "voice_file_id", "transcription", "metadata",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(
		noteID, chatID, userID, "Test Note", "Test content", "idea",
		"active", "telegram", nil, nil, nil,
		now, now, nil,
	)

	mock.ExpectQuery(`SELECT (.+) FROM note WHERE note_id = \$1 AND chat_id = \$2 AND deleted_at IS NULL`).
		WithArgs(noteID, chatID).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetByIDForUser(context.Background(), noteID, chatID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, chatID, result.ChatID)
	assert.Equal(t, userID, result.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_GetNotesForUser(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	chatID := "123456789"
	category := "idea"
	status := "active"
	limit := 10
	offset := 0

	noteID1 := uuid.New()
	noteID2 := uuid.New()
	userID := uuid.New()
	now := time.Now()

	// Mock count query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM note WHERE (.+)`).
		WithArgs(chatID, category, status).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock select query
	rows := sqlmock.NewRows([]string{
		"note_id", "chat_id", "user_id", "title", "content", "category",
		"status", "source", "voice_file_id", "transcription", "metadata",
		"created_at", "updated_at", "deleted_at",
	}).AddRow(
		noteID1, chatID, userID, "Note 1", "Content 1", category,
		status, "telegram", nil, nil, nil,
		now, now, nil,
	).AddRow(
		noteID2, chatID, userID, "Note 2", "Content 2", category,
		status, "telegram", nil, nil, nil,
		now, now, nil,
	)

	mock.ExpectQuery(`SELECT (.+) FROM note WHERE (.+) ORDER BY created_at DESC LIMIT \$4 OFFSET \$5`).
		WithArgs(chatID, category, status, limit, offset).
		WillReturnRows(rows)

	// Execute
	result, err := repo.GetNotesForUser(context.Background(), chatID, &category, &status, limit, offset)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Notes, 2)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, limit, result.Limit)
	assert.Equal(t, offset, result.Offset)
	assert.False(t, result.HasMore) // 0 + 2 >= 2, so no more
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_UpdateForUser(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	note := &domain.Note{
		NoteID:    noteID,
		ChatID:    chatID,
		UserID:    userID,
		Title:     stringPtr("Updated Note"),
		Content:   "Updated content",
		Category:  "work",
		Status:    "active",
		Metadata:  nil, // Use nil for empty metadata
		UpdatedAt: now,
	}

	// Set up mock expectation
	mock.ExpectExec(`UPDATE note SET (.+) WHERE note_id = \$1 AND chat_id = \$2 AND deleted_at IS NULL`).
		WithArgs(noteID, chatID, "Updated Note", "Updated content", "work", "active", nil, nil, nil, now).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	// Execute
	err = repo.UpdateForUser(context.Background(), noteID, chatID, note)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_UpdateForUser_NotFound(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	note := &domain.Note{
		NoteID:    noteID,
		ChatID:    chatID,
		UserID:    userID,
		Title:     stringPtr("Updated Note"),
		Content:   "Updated content",
		Category:  "work",
		Status:    "active",
		Metadata:  nil, // Use nil for empty metadata
		UpdatedAt: now,
	}

	// Set up mock expectation - no rows affected
	mock.ExpectExec(`UPDATE note SET (.+) WHERE note_id = \$1 AND chat_id = \$2 AND deleted_at IS NULL`).
		WithArgs(noteID, chatID, "Updated Note", "Updated content", "work", "active", nil, nil, nil, now).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	// Execute
	err = repo.UpdateForUser(context.Background(), noteID, chatID, note)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_DeleteForUser(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"

	// Set up mock expectation
	mock.ExpectExec(`UPDATE note SET deleted_at = \$3, updated_at = \$3 WHERE note_id = \$1 AND chat_id = \$2 AND deleted_at IS NULL`).
		WithArgs(noteID, chatID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	// Execute
	err = repo.DeleteForUser(context.Background(), noteID, chatID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_DeleteForUser_NotFound(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewNoteRepository(sqlxDB)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"

	// Set up mock expectation - no rows affected
	mock.ExpectExec(`UPDATE note SET deleted_at = \$3, updated_at = \$3 WHERE note_id = \$1 AND chat_id = \$2 AND deleted_at IS NULL`).
		WithArgs(noteID, chatID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	// Execute
	err = repo.DeleteForUser(context.Background(), noteID, chatID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
