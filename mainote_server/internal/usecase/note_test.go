package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mainote-backend/internal/domain"
	api "mainote-backend/pkg/generated/api"
)

// MockNoteRepository is a mock implementation of domain.NoteRepository
type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) Create(note *domain.Note) error {
	args := m.Called(note)
	return args.Error(0)
}

func (m *MockNoteRepository) GetByID(id string) (*domain.Note, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetByChatID(chatID string, limit, offset int) ([]*domain.Note, int, error) {
	args := m.Called(chatID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Note), args.Int(1), args.Error(2)
}

func (m *MockNoteRepository) Update(note *domain.Note) error {
	args := m.Called(note)
	return args.Error(0)
}

func (m *MockNoteRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestNoteUsecase_CreateNote(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		req := api.CreateNoteRequest{
			ChatId:   "123456789",
			Content:  "Test note content",
			Title:    stringPtr("Test Title"),
			Category: api.NOTECATEGORY_IDEA,
			Source:   "telegram_bot",
			Metadata: &map[string]interface{}{
				"user_id":    "user123",
				"message_id": 42,
			},
		}

		// Set up mock expectations
		mockRepo.On("Create", mock.AnythingOfType("*domain.Note")).Return(nil).Run(func(args mock.Arguments) {
			note := args.Get(0).(*domain.Note)
			// Verify the note was created correctly
			assert.Equal(t, req.ChatId, note.ChatID)
			assert.Equal(t, req.Content, note.Content)
			assert.Equal(t, req.Title, note.Title)
			assert.Equal(t, req.Category, note.Category)
			assert.Equal(t, api.NOTESTATUS_ACTIVE, note.Status)
			assert.NotNil(t, note.Metadata)
			assert.Equal(t, "telegram_bot", note.Metadata["source"])
			assert.Equal(t, "user123", note.Metadata["user_id"])
			assert.Equal(t, 42, note.Metadata["message_id"])
			assert.Equal(t, "api", note.Metadata["created_via"])
			assert.Equal(t, "v1", note.Metadata["api_version"])
		})

		// Execute
		result, err := usecase.CreateNote(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.ChatId, result.ChatID)
		assert.Equal(t, req.Content, result.Content)
		assert.Equal(t, req.Category, result.Category)
		assert.Equal(t, api.NOTESTATUS_ACTIVE, result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty chat_id validation", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		req := api.CreateNoteRequest{
			ChatId:   "", // Empty chat ID
			Content:  "Test content",
			Category: api.NOTECATEGORY_IDEA,
		}

		// Execute
		result, err := usecase.CreateNote(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "chat ID cannot be empty")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("empty content validation", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		req := api.CreateNoteRequest{
			ChatId:   "123456789",
			Content:  "", // Empty content
			Category: api.NOTECATEGORY_IDEA,
		}

		// Execute
		result, err := usecase.CreateNote(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "note content cannot be empty")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("default values", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		req := api.CreateNoteRequest{
			ChatId:   "123456789",
			Content:  "Test content",
			Category: api.NOTECATEGORY_TASK,
			// No title, source, or metadata provided
		}

		// Set up mock expectations
		mockRepo.On("Create", mock.AnythingOfType("*domain.Note")).Return(nil).Run(func(args mock.Arguments) {
			note := args.Get(0).(*domain.Note)
			// Verify default values
			assert.Nil(t, note.Title)
			assert.Equal(t, api.NOTESTATUS_ACTIVE, note.Status)
			assert.Equal(t, "telegram_bot", note.Metadata["source"])
			assert.Equal(t, "api", note.Metadata["created_via"])
		})

		// Execute
		result, err := usecase.CreateNote(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		req := api.CreateNoteRequest{
			ChatId:   "123456789",
			Content:  "Test content",
			Category: api.NOTECATEGORY_IDEA,
		}

		// Set up mock to return error
		mockRepo.On("Create", mock.AnythingOfType("*domain.Note")).Return(errors.New("database error"))

		// Execute
		result, err := usecase.CreateNote(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create note")
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("metadata merging", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		req := api.CreateNoteRequest{
			ChatId:   "123456789",
			Content:  "Test content",
			Category: api.NOTECATEGORY_PERSONAL,
			Source:   "voice_message",
			Metadata: &map[string]interface{}{
				"duration":     30.5,
				"transcribed":  true,
				"language":     "en",
				"custom_field": "custom_value",
			},
		}

		// Set up mock expectations
		mockRepo.On("Create", mock.AnythingOfType("*domain.Note")).Return(nil).Run(func(args mock.Arguments) {
			note := args.Get(0).(*domain.Note)
			// Verify metadata merging
			assert.Equal(t, "voice_message", note.Metadata["source"])
			assert.Equal(t, 30.5, note.Metadata["duration"])
			assert.Equal(t, true, note.Metadata["transcribed"])
			assert.Equal(t, "en", note.Metadata["language"])
			assert.Equal(t, "custom_value", note.Metadata["custom_field"])
			assert.Equal(t, "api", note.Metadata["created_via"])
			assert.Equal(t, "v1", note.Metadata["api_version"])
		})

		// Execute
		result, err := usecase.CreateNote(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestNoteUsecase_GetNote(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		expectedNote := &domain.Note{
			ID:       "test-id",
			ChatID:   "123456789",
			Content:  "Test content",
			Category: api.NOTECATEGORY_IDEA,
			Status:   api.NOTESTATUS_ACTIVE,
		}

		// Set up mock expectations
		mockRepo.On("GetByID", "test-id").Return(expectedNote, nil)

		// Execute
		result, err := usecase.GetNote("test-id")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedNote, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		// Setup
		mockRepo := new(MockNoteRepository)
		usecase := NewNoteUsecase(mockRepo)

		// Set up mock to return error
		mockRepo.On("GetByID", "non-existent-id").Return(nil, errors.New("note not found"))

		// Execute
		result, err := usecase.GetNote("non-existent-id")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "note not found")
		mockRepo.AssertExpectations(t)
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
