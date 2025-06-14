/*
 * Mainote Server API
 *
 * REST API for Mainote Server - a note management service.  This API provides CRUD operations for managing notes with support for: - Creating, reading, updating, and deleting notes - Filtering by category, status, and chat ID - Pagination for large result sets - Rich metadata support using JSON  ## Authentication Currently no authentication is required (development mode).  ## Rate Limiting No rate limiting is currently implemented.
 *
 * API version: 1.0.0
 * Contact: support@mainote.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

type CreateNoteRequest struct {

	// Telegram chat ID associated with the note
	ChatId string `json:"chat_id"`

	// The main content/text of the note
	Content string `json:"content"`

	// Optional title for the note
	Title *string `json:"title,omitempty"`

	Category NoteCategory `json:"category"`

	// Source of the note
	Source string `json:"source,omitempty"`

	// Additional metadata in JSON format
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
}

// AssertCreateNoteRequestRequired checks if the required fields are not zero-ed
func AssertCreateNoteRequestRequired(obj *CreateNoteRequest) (err error) {
	elements := map[string]any{
		"chat_id":  obj.ChatId,
		"content":  obj.Content,
		"category": obj.Category,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}
	return nil
}

// AssertCreateNoteRequestConstraints checks if the values respects the defined constraints
func AssertCreateNoteRequestConstraints(obj *CreateNoteRequest) error {
	return nil
}
