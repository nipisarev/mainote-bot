package handler

import (
	"context"
	"encoding/json"
	"time"

	"mainote-server/internal/domain"
	"mainote-server/internal/usecase"
	api "mainote-server/pkg/generated/api"

	"github.com/google/uuid"
)

// NoteHandler handles HTTP requests for notes and implements NotesAPIServicer.
type NoteHandler struct {
	noteUsecase usecase.NoteUsecase
}

// NewNoteHandler creates a new instance of NoteHandler.
func NewNoteHandler(noteUsecase usecase.NoteUsecase) *NoteHandler {
	return &NoteHandler{noteUsecase: noteUsecase}
}

// CreateNote implements NotesAPIServicer interface.
func (h *NoteHandler) CreateNote(ctx context.Context, req api.CreateNoteRequest) (api.ImplResponse, error) {
	// Convert optional string fields to pointers
	var voiceFileID, transcription *string
	if req.VoiceFileId != "" {
		voiceFileID = &req.VoiceFileId
	}
	if req.Transcription != "" {
		transcription = &req.Transcription
	}

	// Convert metadata if provided and merge extended fields if present
	var metadata interface{}
	if len(req.Metadata) > 0 {
		metadata = req.Metadata
	}
	// Merge due_at/effort_min/priority into metadata to allow usecase to map
	// when updating later; create path sets defaults in usecase.
	if (req.DueAt != time.Time{}) || req.EffortMin != 0 || req.Priority != 0 {
		mm := map[string]interface{}{}
		if m, ok := metadata.(map[string]interface{}); ok {
			for k, v := range m {
				mm[k] = v
			}
		}
		if req.DueAt != (time.Time{}) {
			mm["due_at"] = req.DueAt.Format(time.RFC3339)
		}
		if req.EffortMin != 0 {
			mm["effort_min"] = req.EffortMin
		}
		if req.Priority != 0 {
			mm["priority"] = req.Priority
		}
		if len(mm) > 0 {
			metadata = mm
		}
	}

	note, err := h.noteUsecase.CreateNote(ctx, req.ChatId, req.Title, req.Content,
		req.Category, req.Status, req.Source, voiceFileID, transcription, metadata)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+req.ChatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain note to API response
	noteResponse := domainNoteToAPIResponse(note)
	return api.Response(201, noteResponse), nil
}

// GetNotes implements NotesAPIServicer interface.
func (h *NoteHandler) GetNotes(ctx context.Context, chatId string, category string, status string, limit int32, offset int32) (api.ImplResponse, error) {
	// Convert optional parameters
	var categoryPtr, statusPtr *string
	if category != "" {
		categoryPtr = &category
	}
	if status != "" {
		statusPtr = &status
	}

	result, err := h.noteUsecase.GetNotesForUser(ctx, chatId, categoryPtr, statusPtr, int(limit), int(offset))
	if err != nil {
		if err.Error() == "user not found for chat_id: "+chatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain result to API response
	notes := make([]api.NoteResponse, len(result.Notes))
	for i, note := range result.Notes {
		notes[i] = domainNoteToAPIResponse(&note)
	}

	listResponse := api.NotesListResponse{
		Notes:   notes,
		Total:   int32(result.Total),
		Limit:   int32(result.Limit),
		Offset:  int32(result.Offset),
		HasMore: result.HasMore,
	}

	return api.Response(200, listResponse), nil
}

// GetNoteById implements NotesAPIServicer interface.
func (h *NoteHandler) GetNoteById(ctx context.Context, noteId string, chatId string) (api.ImplResponse, error) {
	// Parse noteId
	noteUUID, err := uuid.Parse(noteId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid note_id format",
		}), nil
	}

	note, err := h.noteUsecase.GetNoteByID(ctx, noteUUID, chatId)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+chatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}
		if err.Error() == "note not found" {
			return api.Response(404, api.ErrorResponse{
				Error:   "Note not found",
				Message: "No note found with the provided note_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain note to API response
	noteResponse := domainNoteToAPIResponse(note)
	return api.Response(200, noteResponse), nil
}

// UpdateNote implements NotesAPIServicer interface.
func (h *NoteHandler) UpdateNote(ctx context.Context, noteId string, chatId string, req api.UpdateNoteRequest) (api.ImplResponse, error) {
	// Parse noteId
	noteUUID, err := uuid.Parse(noteId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid note_id format",
		}), nil
	}

	// Convert optional fields to pointers
	var title, content, category, status, voiceFileID, transcription *string
	var metadata interface{}

	if req.Title != "" {
		title = &req.Title
	}
	if req.Content != "" {
		content = &req.Content
	}
	if req.Category != "" {
		category = &req.Category
	}
	if req.Status != "" {
		status = &req.Status
	}
	if req.VoiceFileId != "" {
		voiceFileID = &req.VoiceFileId
	}
	if req.Transcription != "" {
		transcription = &req.Transcription
	}
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	// Include due_at/effort_min/priority in metadata if provided
	// The usecase maps these to concrete fields on the note.
	if req.DueAt != (time.Time{}) {
		if metadata == nil {
			metadata = map[string]interface{}{}
		}
		mm := metadata.(map[string]interface{})
		mm["due_at"] = req.DueAt.Format(time.RFC3339)
	}
	if req.EffortMin != 0 {
		if metadata == nil {
			metadata = map[string]interface{}{}
		}
		mm := metadata.(map[string]interface{})
		mm["effort_min"] = req.EffortMin
	}
	if req.Priority != 0 {
		if metadata == nil {
			metadata = map[string]interface{}{}
		}
		mm := metadata.(map[string]interface{})
		mm["priority"] = req.Priority
	}

	// Check if at least one field is provided
	if title == nil && content == nil && category == nil && status == nil &&
		voiceFileID == nil && transcription == nil && metadata == nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "At least one field must be provided for update",
		}), nil
	}

	note, err := h.noteUsecase.UpdateNote(ctx, noteUUID, chatId, title, content, category, status, voiceFileID, transcription, metadata)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+chatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}
		if err.Error() == "note not found" {
			return api.Response(404, api.ErrorResponse{
				Error:   "Note not found",
				Message: "No note found with the provided note_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain note to API response
	noteResponse := domainNoteToAPIResponse(note)
	return api.Response(200, noteResponse), nil
}

// DeleteNote implements NotesAPIServicer interface.
func (h *NoteHandler) DeleteNote(ctx context.Context, noteId string, chatId string) (api.ImplResponse, error) {
	// Parse noteId
	noteUUID, err := uuid.Parse(noteId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid note_id format",
		}), nil
	}

	err = h.noteUsecase.DeleteNote(ctx, noteUUID, chatId)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+chatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}
		if err.Error() == "note not found" {
			return api.Response(404, api.ErrorResponse{
				Error:   "Note not found",
				Message: "No note found with the provided note_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Return 204 No Content for successful deletion
	return api.Response(204, nil), nil
}

// Helper function to convert domain Note to API NoteResponse
func domainNoteToAPIResponse(note *domain.Note) api.NoteResponse {
	var title, voiceFileID, transcription *string
	var deletedAt *time.Time
	var dueAt *time.Time

	// Handle nullable fields
	if note.Title != nil {
		title = note.Title
	}
	if note.VoiceFileID != nil {
		voiceFileID = note.VoiceFileID
	}
	if note.Transcription != nil {
		transcription = note.Transcription
	}
	if note.DeletedAt != nil {
		deletedAt = note.DeletedAt
	}
	if note.DueAt != nil {
		dueAt = note.DueAt
	}

	// Convert metadata to map[string]interface{} if it exists
	var metadataMap *map[string]interface{}
	if note.Metadata != nil {
		// Try to convert to map[string]interface{}
		if metadataBytes, err := json.Marshal(note.Metadata); err == nil {
			var temp map[string]interface{}
			if json.Unmarshal(metadataBytes, &temp) == nil {
				metadataMap = &temp
			}
		}
	}

	return api.NoteResponse{
		NoteId:        note.NoteID.String(),
		ChatId:        note.ChatID,
		UserId:        note.UserID.String(),
		Title:         title,
		Content:       note.Content,
		Category:      note.Category,
		Status:        note.Status,
		Source:        note.Source,
		VoiceFileId:   voiceFileID,
		Transcription: transcription,
		DueAt:         dueAt,
		EffortMin:     int32(note.EffortMin),
		Priority:      int32(note.Priority),
		Metadata:      metadataMap,
		CreatedAt:     note.CreatedAt,
		UpdatedAt:     note.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}
