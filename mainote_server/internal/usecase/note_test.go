package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"mainote-server/internal/domain"
)

// Mock repositories
type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) Create(ctx context.Context, note *domain.Note) error {
	args := m.Called(ctx, note)
	return args.Error(0)
}

func (m *MockNoteRepository) GetByID(ctx context.Context, noteID uuid.UUID) (*domain.Note, error) {
	args := m.Called(ctx, noteID)
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) Update(ctx context.Context, note *domain.Note) error {
	args := m.Called(ctx, note)
	return args.Error(0)
}

func (m *MockNoteRepository) Delete(ctx context.Context, noteID uuid.UUID) error {
	args := m.Called(ctx, noteID)
	return args.Error(0)
}

func (m *MockNoteRepository) GetByIDForUser(ctx context.Context, noteID uuid.UUID, chatID string) (*domain.Note, error) {
	args := m.Called(ctx, noteID, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetNotesForUser(ctx context.Context, chatID string, category, status *string, limit, offset int) (*domain.NotesListResult, error) {
	args := m.Called(ctx, chatID, category, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NotesListResult), args.Error(1)
}

func (m *MockNoteRepository) UpdateForUser(ctx context.Context, noteID uuid.UUID, chatID string, note *domain.Note) error {
	args := m.Called(ctx, noteID, chatID, note)
	return args.Error(0)
}

func (m *MockNoteRepository) DeleteForUser(ctx context.Context, noteID uuid.UUID, chatID string) error {
	args := m.Called(ctx, noteID, chatID)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByChatID(ctx context.Context, chatID string) (*domain.UserWithSettings, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithSettings), args.Error(1)
}

func (m *MockUserRepository) UpdateChatID(ctx context.Context, userID uuid.UUID, chatID string) error {
	args := m.Called(ctx, userID, chatID)
	return args.Error(0)
}

func (m *MockUserRepository) CreateUserSettings(ctx context.Context, userID uuid.UUID, chatID string, morningNotificationTime, timezone *string) (*domain.UserSettings, error) {
	args := m.Called(ctx, userID, chatID, morningNotificationTime, timezone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSettings), args.Error(1)
}

func (m *MockUserRepository) UpdateUserSettings(ctx context.Context, update *domain.UserSettings) (*domain.UserSettings, error) {
	args := m.Called(ctx, update)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSettings), args.Error(1)
}

func TestNoteUsecase_CreateNote(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	userWithSettings := &domain.UserWithSettings{
		User: domain.User{
			ID:        userID,
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Settings: domain.UserSettings{
			UserID: userID,
			ChatID: chatID,
		},
	}

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(userWithSettings, nil)
	mockNoteRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Note")).Return(nil)

	// Execute
	result, err := usecase.CreateNote(context.Background(), chatID, "Test Note", "Test content", "", "", "", nil, nil, nil)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, chatID, result.ChatID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "Test Note", *result.Title)
	assert.Equal(t, "Test content", result.Content)
	assert.Equal(t, "idea", result.Category)   // default value
	assert.Equal(t, "active", result.Status)   // default value
	assert.Equal(t, "telegram", result.Source) // default value

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteUsecase_CreateNote_UserNotFound(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	chatID := "123456789"

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(nil, sql.ErrNoRows)

	// Execute
	result, err := usecase.CreateNote(context.Background(), chatID, "Test Note", "Test content", "", "", "", nil, nil, nil)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found for chat_id")

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertNotCalled(t, "Create")
}

func TestNoteUsecase_GetNoteByID(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	userWithSettings := &domain.UserWithSettings{
		User: domain.User{
			ID:        userID,
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Settings: domain.UserSettings{
			UserID: userID,
			ChatID: chatID,
		},
	}

	note := &domain.Note{
		NoteID:    noteID,
		ChatID:    chatID,
		UserID:    userID,
		Title:     stringPtr("Test Note"),
		Content:   "Test content",
		Category:  "idea",
		Status:    "active",
		Source:    "telegram",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(userWithSettings, nil)
	mockNoteRepo.On("GetByIDForUser", mock.Anything, noteID, chatID).Return(note, nil)

	// Execute
	result, err := usecase.GetNoteByID(context.Background(), noteID, chatID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, chatID, result.ChatID)
	assert.Equal(t, userID, result.UserID)

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteUsecase_GetNoteByID_NoteNotFound(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	userWithSettings := &domain.UserWithSettings{
		User: domain.User{
			ID:        userID,
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Settings: domain.UserSettings{
			UserID: userID,
			ChatID: chatID,
		},
	}

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(userWithSettings, nil)
	mockNoteRepo.On("GetByIDForUser", mock.Anything, noteID, chatID).Return(nil, sql.ErrNoRows)

	// Execute
	result, err := usecase.GetNoteByID(context.Background(), noteID, chatID)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "note not found")

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteUsecase_GetNotesForUser(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	chatID := "123456789"
	userID := uuid.New()
	category := "idea"
	status := "active"
	limit := 10
	offset := 0
	now := time.Now()

	userWithSettings := &domain.UserWithSettings{
		User: domain.User{
			ID:        userID,
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Settings: domain.UserSettings{
			UserID: userID,
			ChatID: chatID,
		},
	}

	notesResult := &domain.NotesListResult{
		Notes: []domain.Note{
			{
				NoteID:    uuid.New(),
				ChatID:    chatID,
				UserID:    userID,
				Title:     stringPtr("Note 1"),
				Content:   "Content 1",
				Category:  category,
				Status:    status,
				Source:    "telegram",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Total:   1,
		Limit:   limit,
		Offset:  offset,
		HasMore: false,
	}

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(userWithSettings, nil)
	mockNoteRepo.On("GetNotesForUser", mock.Anything, chatID, &category, &status, limit, offset).Return(notesResult, nil)

	// Execute
	result, err := usecase.GetNotesForUser(context.Background(), chatID, &category, &status, limit, offset)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Notes, 1)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, limit, result.Limit)
	assert.Equal(t, offset, result.Offset)
	assert.False(t, result.HasMore)

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteUsecase_UpdateNote(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	userWithSettings := &domain.UserWithSettings{
		User: domain.User{
			ID:        userID,
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Settings: domain.UserSettings{
			UserID: userID,
			ChatID: chatID,
		},
	}

	existingNote := &domain.Note{
		NoteID:    noteID,
		ChatID:    chatID,
		UserID:    userID,
		Title:     stringPtr("Original Title"),
		Content:   "Original content",
		Category:  "idea",
		Status:    "active",
		Source:    "telegram",
		CreatedAt: now,
		UpdatedAt: now,
	}

	updatedTitle := "Updated Title"
	updatedContent := "Updated content"

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(userWithSettings, nil)
	mockNoteRepo.On("GetByIDForUser", mock.Anything, noteID, chatID).Return(existingNote, nil)
	mockNoteRepo.On("UpdateForUser", mock.Anything, noteID, chatID, mock.AnythingOfType("*domain.Note")).Return(nil)

	// Execute
	result, err := usecase.UpdateNote(context.Background(), noteID, chatID, &updatedTitle, &updatedContent, nil, nil, nil, nil, nil)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, updatedTitle, *result.Title)
	assert.Equal(t, updatedContent, result.Content)

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteUsecase_DeleteNote(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	userWithSettings := &domain.UserWithSettings{
		User: domain.User{
			ID:        userID,
			Email:     "test@example.com",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Settings: domain.UserSettings{
			UserID: userID,
			ChatID: chatID,
		},
	}

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(userWithSettings, nil)
	mockNoteRepo.On("DeleteForUser", mock.Anything, noteID, chatID).Return(nil)

	// Execute
	err := usecase.DeleteNote(context.Background(), noteID, chatID)

	// Assert
	require.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteUsecase_DeleteNote_UserNotFound(t *testing.T) {
	// Setup
	mockNoteRepo := new(MockNoteRepository)
	mockUserRepo := new(MockUserRepository)
	usecase := NewNoteUsecase(mockNoteRepo, mockUserRepo)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"

	// Mock expectations
	mockUserRepo.On("FindByChatID", mock.Anything, chatID).Return(nil, sql.ErrNoRows)

	// Execute
	err := usecase.DeleteNote(context.Background(), noteID, chatID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found for chat_id")

	mockUserRepo.AssertExpectations(t)
	mockNoteRepo.AssertNotCalled(t, "DeleteForUser")
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
