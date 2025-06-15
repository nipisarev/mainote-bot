package handler

import (
	"context"
	"net/http"
	"time"

	"mainote-backend/internal/domain"
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

// convertDomainNoteToAPI converts a domain note to API note
func convertDomainNoteToAPI(domainNote *domain.Note) api.Note {
	return api.Note{
		Id:        domainNote.ID,
		ChatId:    domainNote.ChatID,
		Content:   domainNote.Content,
		Title:     domainNote.Title,
		Category:  domainNote.Category,
		Status:    domainNote.Status,
		Metadata:  &domainNote.Metadata,
		CreatedAt: domainNote.CreatedAt,
		UpdatedAt: domainNote.UpdatedAt,
	}
}

// Implement NotesAPIServicer interface

// CreateNote creates a new note
func (h *NoteHandler) CreateNote(ctx context.Context, createNoteRequest api.CreateNoteRequest) (api.ImplResponse, error) {
	// Call usecase to create note
	domainNote, err := h.noteUsecase.CreateNote(createNoteRequest)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
		}), nil
	}

	// Convert domain note to API note
	apiNote := convertDomainNoteToAPI(domainNote)

	return api.Response(201, apiNote), nil
}

// DeleteNote deletes a note by ID
func (h *NoteHandler) DeleteNote(ctx context.Context, noteId string) (api.ImplResponse, error) {
	// Call usecase to delete note
	err := h.noteUsecase.DeleteNote(noteId)
	if err != nil {
		return api.Response(404, api.ErrorResponse{
			Error:   "Not Found",
			Message: err.Error(),
		}), nil
	}

	return api.Response(204, nil), nil
}

// GetNoteById retrieves a note by ID (matching the generated interface)
func (h *NoteHandler) GetNoteById(ctx context.Context, noteId string) (api.ImplResponse, error) {
	// Call usecase to get note
	domainNote, err := h.noteUsecase.GetNote(noteId)
	if err != nil {
		return api.Response(404, api.ErrorResponse{
			Error:   "Not Found",
			Message: err.Error(),
		}), nil
	}

	// Convert domain note to API note
	apiNote := convertDomainNoteToAPI(domainNote)

	return api.Response(200, apiNote), nil
}

// ListNotes retrieves notes with filtering (matching the generated interface)
func (h *NoteHandler) ListNotes(ctx context.Context, chatId string, category api.NoteCategory, status api.NoteStatus, limit int32, offset int32) (api.ImplResponse, error) {
	// Call usecase to get notes
	domainNotes, total, err := h.noteUsecase.ListNotes(chatId, int(limit), int(offset))
	if err != nil {
		return api.Response(500, api.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain notes to API notes
	apiNotes := make([]api.Note, len(domainNotes))
	for i, domainNote := range domainNotes {
		apiNotes[i] = convertDomainNoteToAPI(domainNote)
	}

	response := api.NotesListResponse{
		Notes: apiNotes,
		Pagination: api.PaginationInfo{
			Total:  int32(total),
			Limit:  limit,
			Offset: offset,
		},
	}

	return api.Response(200, response), nil
}

// UpdateNote updates an existing note
func (h *NoteHandler) UpdateNote(ctx context.Context, noteId string, updateNoteRequest api.UpdateNoteRequest) (api.ImplResponse, error) {
	// Call usecase to update note
	domainNote, err := h.noteUsecase.UpdateNote(noteId, updateNoteRequest)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
		}), nil
	}

	// Convert domain note to API note
	apiNote := convertDomainNoteToAPI(domainNote)

	return api.Response(200, apiNote), nil
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
