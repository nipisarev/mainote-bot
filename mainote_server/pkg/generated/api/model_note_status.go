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

// NoteStatus : Status of the note: * `active` - Active note (default) * `done` - Completed/finished note * `archived` - Archived/inactive note
type NoteStatus string

// List of NoteStatus
const (
	NOTESTATUS_ACTIVE   NoteStatus = "active"
	NOTESTATUS_DONE     NoteStatus = "done"
	NOTESTATUS_ARCHIVED NoteStatus = "archived"
)

// AssertNoteStatusRequired checks if the required fields are not zero-ed
func AssertNoteStatusRequired(obj *NoteStatus) (err error) {
	return nil
}

// AssertNoteStatusConstraints checks if the values respects the defined constraints
func AssertNoteStatusConstraints(obj *NoteStatus) error {
	return nil
}
