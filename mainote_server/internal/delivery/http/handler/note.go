package handler

import (
	"context"
	"net/http"
	"time"

	"mainote-backend/internal/usecase"
	api "mainote-backend/pkg/generated/api"
)

// NoteHandler implements the generated API interfaces
type NoteHandler struct {
	noteUsecase *usecase.NoteUsecase
}

// NewNoteHandler creates a new note handler
func NewNoteHandler(noteUsecase *usecase.NoteUsecase) *NoteHandler {
	return &NoteHandler{
		noteUsecase: noteUsecase,
	}
}

// Implement NotesAPIServicer interface

// CreateNote creates a new note
func (h *NoteHandler) CreateNote(ctx context.Context, createNoteRequest api.CreateNoteRequest) (api.ImplResponse, error) {
	// TODO: Implement create note logic
	// This is a stub implementation
	return api.Response(201, api.Note{}), nil
}

// DeleteNote deletes a note by ID
func (h *NoteHandler) DeleteNote(ctx context.Context, noteId string) (api.ImplResponse, error) {
	// TODO: Implement delete note logic
	// This is a stub implementation
	return api.Response(204, nil), nil
}

// GetNoteById retrieves a note by ID (matching the generated interface)
func (h *NoteHandler) GetNoteById(ctx context.Context, noteId string) (api.ImplResponse, error) {
	// TODO: Implement get note logic
	// This is a stub implementation
	return api.Response(200, api.Note{}), nil
}

// ListNotes retrieves notes with filtering (matching the generated interface)
func (h *NoteHandler) ListNotes(ctx context.Context, chatId string, category api.NoteCategory, status api.NoteStatus, limit int32, offset int32) (api.ImplResponse, error) {
	// TODO: Implement list notes logic
	// This is a stub implementation
	response := api.NotesListResponse{
		Notes: []api.Note{},
		Pagination: api.PaginationInfo{
			Total:  0,
			Limit:  limit,
			Offset: offset,
		},
	}
	return api.Response(200, response), nil
}

// UpdateNote updates an existing note
func (h *NoteHandler) UpdateNote(ctx context.Context, noteId string, updateNoteRequest api.UpdateNoteRequest) (api.ImplResponse, error) {
	// TODO: Implement update note logic
	// This is a stub implementation
	return api.Response(200, api.Note{}), nil
}

// Implement HealthAPIServicer interface

// CheckHealth performs a health check
func (h *NoteHandler) CheckHealth(ctx context.Context) (api.ImplResponse, error) {
	response := api.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
	return api.Response(200, response), nil
}

// Legacy HTTP handler for health check (for backward compatibility)
func (h *NoteHandler) CheckHealthLegacy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","version":"1.0.0"}`))
}
