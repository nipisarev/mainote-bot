package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mainote-backend/internal/domain"
	api "mainote-backend/pkg/generated/api"
)

func TestNoteRepository_Create(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)

	t.Run("successful creation", func(t *testing.T) {
		// Prepare test data
		note := &domain.Note{
			ChatID:   "123456789",
			Content:  "Test note content",
			Title:    stringPtr("Test Title"),
			Category: api.NOTECATEGORY_IDEA,
			Status:   api.NOTESTATUS_ACTIVE,
			Metadata: map[string]interface{}{
				"source":  "telegram_bot",
				"user_id": "user123",
			},
		}

		// Set up mock expectations
		mock.ExpectQuery(`INSERT INTO notes`).
			WithArgs(
				sqlmock.AnyArg(), // uuid_id (generated)
				"123456789",      // chat_id
				"Test note content", // content
				"Test Title",     // title
				"idea",          // category
				"active",        // status
				"telegram_bot",  // source
				sqlmock.AnyArg(), // metadata (JSON)
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnRows(
				sqlmock.NewRows([]string{"uuid_id", "created_at", "updated_at"}).
					AddRow("test-uuid", time.Now(), time.Now()),
			)

		// Execute
		err := repo.Create(note)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, note.ID)
		assert.NotZero(t, note.CreatedAt)
		assert.NotZero(t, note.UpdatedAt)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		note := &domain.Note{
			ChatID:   "123456789",
			Content:  "Test note content",
			Category: api.NOTECATEGORY_IDEA,
			Status:   api.NOTESTATUS_ACTIVE,
			Metadata: make(map[string]interface{}),
		}

		// Set up mock to return error
		mock.ExpectQuery(`INSERT INTO notes`).
			WillReturnError(sql.ErrConnDone)

		// Execute
		err := repo.Create(note)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create note")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid metadata JSON", func(t *testing.T) {
		note := &domain.Note{
			ChatID:   "123456789",
			Content:  "Test note content",
			Category: api.NOTECATEGORY_IDEA,
			Status:   api.NOTESTATUS_ACTIVE,
			Metadata: map[string]interface{}{
				"invalid": make(chan int), // Channels can't be marshaled to JSON
			},
		}

		// Execute
		err := repo.Create(note)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal metadata")
	})
}

func TestMarshalUnmarshalMetadata(t *testing.T) {
	t.Run("marshal and unmarshal metadata", func(t *testing.T) {
		original := map[string]interface{}{
			"source":     "telegram_bot",
			"user_id":    "user123",
			"message_id": float64(42), // JSON numbers are float64
			"is_voice":   true,
		}

		// Marshal
		data, err := marshalMetadata(original)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Unmarshal
		result, err := unmarshalMetadata(data)
		assert.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("nil metadata", func(t *testing.T) {
		// Marshal nil
		data, err := marshalMetadata(nil)
		assert.NoError(t, err)
		assert.Nil(t, data)

		// Unmarshal nil
		result, err := unmarshalMetadata(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("empty metadata", func(t *testing.T) {
		original := make(map[string]interface{})

		// Marshal empty map
		data, err := marshalMetadata(original)
		assert.NoError(t, err)
		assert.Equal(t, []byte("{}"), data)

		// Unmarshal
		result, err := unmarshalMetadata(data)
		assert.NoError(t, err)
		assert.Equal(t, original, result)
	})

	t.Run("invalid JSON unmarshal", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": json}`)

		result, err := unmarshalMetadata(invalidJSON)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to unmarshal metadata")
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
