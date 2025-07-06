package handler

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"mainote-server/internal/domain"
	api "mainote-server/pkg/generated/api"
)

// Mock usecase
type MockNoteUsecase struct {
	mock.Mock
}

func (m *MockNoteUsecase) CreateNote(ctx context.Context, chatID, title, content, category, status, source string, voiceFileID, transcription *string, metadata interface{}) (*domain.Note, error) {
	args := m.Called(ctx, chatID, title, content, category, status, source, voiceFileID, transcription, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteUsecase) GetNoteByID(ctx context.Context, noteID uuid.UUID, chatID string) (*domain.Note, error) {
	args := m.Called(ctx, noteID, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteUsecase) GetNotesForUser(ctx context.Context, chatID string, category, status *string, limit, offset int) (*domain.NotesListResult, error) {
	args := m.Called(ctx, chatID, category, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NotesListResult), args.Error(1)
}

func (m *MockNoteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, chatID string, title, content, category, status *string, voiceFileID, transcription *string, metadata interface{}) (*domain.Note, error) {
	args := m.Called(ctx, noteID, chatID, title, content, category, status, voiceFileID, transcription, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, chatID string) error {
	args := m.Called(ctx, noteID, chatID)
	return args.Error(0)
}

func TestNoteHandler_CreateNote(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	chatID := "123456789"
	userID := uuid.New()
	noteID := uuid.New()
	now := time.Now()

	req := api.CreateNoteRequest{
		ChatId:   chatID,
		Title:    "Test Note",
		Content:  "Test content",
		Category: "idea",
		Status:   "active",
		Source:   "telegram",
	}

	expectedNote := &domain.Note{
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
	mockUsecase.On("CreateNote", mock.Anything, chatID, "Test Note", "Test content", "idea", "active", "telegram", (*string)(nil), (*string)(nil), nil).Return(expectedNote, nil)

	// Execute
	result, err := handler.CreateNote(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 201, result.Code)

	noteResponse := result.Body.(api.NoteResponse)
	assert.Equal(t, noteID.String(), noteResponse.NoteId)
	assert.Equal(t, chatID, noteResponse.ChatId)
	assert.Equal(t, userID.String(), noteResponse.UserId)
	assert.Equal(t, "Test Note", *noteResponse.Title)
	assert.Equal(t, "Test content", noteResponse.Content)

	mockUsecase.AssertExpectations(t)
}

func TestNoteHandler_CreateNote_UserNotFound(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	chatID := "123456789"

	req := api.CreateNoteRequest{
		ChatId:  chatID,
		Content: "Test content",
	}

	// Mock expectations
	mockUsecase.On("CreateNote", mock.Anything, chatID, "", "Test content", "", "", "", (*string)(nil), (*string)(nil), nil).Return(nil, assert.AnError)

	// Simulate user not found error
	mockUsecase.ExpectedCalls[0].ReturnArguments[1] = assert.AnError
	mockUsecase.ExpectedCalls[0].ReturnArguments[1] = &MockError{message: "user not found for chat_id: " + chatID}

	// Execute
	result, err := handler.CreateNote(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 404, result.Code)

	errorResponse := result.Body.(api.ErrorResponse)
	assert.Equal(t, "User not found", errorResponse.Error)
	assert.Equal(t, "No user found with the provided chat_id", errorResponse.Message)
}

func TestNoteHandler_GetNotes(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	chatID := "123456789"
	category := "idea"
	status := "active"
	limit := int32(10)
	offset := int32(0)

	noteID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	notesResult := &domain.NotesListResult{
		Notes: []domain.Note{
			{
				NoteID:    noteID,
				ChatID:    chatID,
				UserID:    userID,
				Title:     stringPtr("Test Note"),
				Content:   "Test content",
				Category:  category,
				Status:    status,
				Source:    "telegram",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Total:   1,
		Limit:   10,
		Offset:  0,
		HasMore: false,
	}

	// Mock expectations
	mockUsecase.On("GetNotesForUser", mock.Anything, chatID, &category, &status, 10, 0).Return(notesResult, nil)

	// Execute
	result, err := handler.GetNotes(context.Background(), chatID, category, status, limit, offset)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 200, result.Code)

	listResponse := result.Body.(api.NotesListResponse)
	assert.Len(t, listResponse.Notes, 1)
	assert.Equal(t, int32(1), listResponse.Total)
	assert.Equal(t, limit, listResponse.Limit)
	assert.Equal(t, offset, listResponse.Offset)
	assert.False(t, listResponse.HasMore)

	mockUsecase.AssertExpectations(t)
}

func TestNoteHandler_GetNoteById(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	expectedNote := &domain.Note{
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
	mockUsecase.On("GetNoteByID", mock.Anything, noteID, chatID).Return(expectedNote, nil)

	// Execute
	result, err := handler.GetNoteById(context.Background(), noteID.String(), chatID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 200, result.Code)

	noteResponse := result.Body.(api.NoteResponse)
	assert.Equal(t, noteID.String(), noteResponse.NoteId)
	assert.Equal(t, chatID, noteResponse.ChatId)
	assert.Equal(t, userID.String(), noteResponse.UserId)
	assert.Equal(t, "Test Note", *noteResponse.Title)

	mockUsecase.AssertExpectations(t)
}

func TestNoteHandler_GetNoteById_InvalidUUID(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	invalidNoteID := "invalid-uuid"
	chatID := "123456789"

	// Execute
	result, err := handler.GetNoteById(context.Background(), invalidNoteID, chatID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 400, result.Code)

	errorResponse := result.Body.(api.ErrorResponse)
	assert.Equal(t, "Bad request", errorResponse.Error)
	assert.Equal(t, "Invalid note_id format", errorResponse.Message)

	// Should not call usecase with invalid UUID
	mockUsecase.AssertNotCalled(t, "GetNoteByID")
}

func TestNoteHandler_UpdateNote(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"
	userID := uuid.New()
	now := time.Now()

	req := api.UpdateNoteRequest{
		Title:    "Updated Title",
		Content:  "Updated content",
		Category: "work",
	}

	expectedNote := &domain.Note{
		NoteID:    noteID,
		ChatID:    chatID,
		UserID:    userID,
		Title:     stringPtr("Updated Title"),
		Content:   "Updated content",
		Category:  "work",
		Status:    "active",
		Source:    "telegram",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Mock expectations
	mockUsecase.On("UpdateNote", mock.Anything, noteID, chatID,
		stringPtr("Updated Title"), stringPtr("Updated content"), stringPtr("work"),
		(*string)(nil), (*string)(nil), (*string)(nil), nil).Return(expectedNote, nil)

	// Execute
	result, err := handler.UpdateNote(context.Background(), noteID.String(), chatID, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 200, result.Code)

	noteResponse := result.Body.(api.NoteResponse)
	assert.Equal(t, "Updated Title", *noteResponse.Title)
	assert.Equal(t, "Updated content", noteResponse.Content)
	assert.Equal(t, "work", noteResponse.Category)

	mockUsecase.AssertExpectations(t)
}

func TestNoteHandler_UpdateNote_NoFieldsProvided(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"

	req := api.UpdateNoteRequest{
		// All fields empty
	}

	// Execute
	result, err := handler.UpdateNote(context.Background(), noteID.String(), chatID, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 400, result.Code)

	errorResponse := result.Body.(api.ErrorResponse)
	assert.Equal(t, "Bad request", errorResponse.Error)
	assert.Equal(t, "At least one field must be provided for update", errorResponse.Message)

	// Should not call usecase
	mockUsecase.AssertNotCalled(t, "UpdateNote")
}

func TestNoteHandler_DeleteNote(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"

	// Mock expectations
	mockUsecase.On("DeleteNote", mock.Anything, noteID, chatID).Return(nil)

	// Execute
	result, err := handler.DeleteNote(context.Background(), noteID.String(), chatID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 204, result.Code)
	assert.Nil(t, result.Body)

	mockUsecase.AssertExpectations(t)
}

func TestNoteHandler_DeleteNote_NotFound(t *testing.T) {
	// Setup
	mockUsecase := new(MockNoteUsecase)
	handler := NewNoteHandler(mockUsecase)

	// Test data
	noteID := uuid.New()
	chatID := "123456789"

	// Mock expectations
	mockUsecase.On("DeleteNote", mock.Anything, noteID, chatID).Return(&MockError{message: "note not found"})

	// Execute
	result, err := handler.DeleteNote(context.Background(), noteID.String(), chatID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 404, result.Code)

	errorResponse := result.Body.(api.ErrorResponse)
	assert.Equal(t, "Note not found", errorResponse.Error)
	assert.Equal(t, "No note found with the provided note_id", errorResponse.Message)

	mockUsecase.AssertExpectations(t)
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// Mock error for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}
