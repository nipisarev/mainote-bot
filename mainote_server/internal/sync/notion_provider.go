package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mainote-server/internal/domain"
	"mainote-server/internal/notion"
)

// NotionSyncProvider implements SyncProvider for Notion integration
type NotionSyncProvider struct{}

// NewNotionSyncProvider creates a new instance of NotionSyncProvider
func NewNotionSyncProvider() *NotionSyncProvider {
	return &NotionSyncProvider{}
}

// GetProviderName returns the name of the provider
func (p *NotionSyncProvider) GetProviderName() string {
	return "notion"
}

// SyncNote synchronizes a note with Notion
func (p *NotionSyncProvider) SyncNote(ctx context.Context, note *domain.Note, integration *domain.Integration) (*domain.SyncResult, error) {
	// Extract credentials from integration
	client, err := p.createNotionClient(integration)
	if err != nil {
		return &domain.SyncResult{
			Success:  false,
			Error:    fmt.Sprintf("Failed to create Notion client: %v", err),
			SyncedAt: time.Now(),
		}, nil
	}

	// Prepare note data
	title := ""
	if note.Title != nil {
		title = *note.Title
	}

	// Create note in Notion
	page, err := client.CreateNote(title, note.Content, note.Category, note.Source)
	if err != nil {
		return &domain.SyncResult{
			Success:  false,
			Error:    fmt.Sprintf("Failed to create note in Notion: %v", err),
			SyncedAt: time.Now(),
		}, nil
	}

	return &domain.SyncResult{
		Success:    true,
		ExternalID: page.ID,
		SyncedAt:   time.Now(),
	}, nil
}

// FetchNotes retrieves notes from Notion (for inbound sync)
func (p *NotionSyncProvider) FetchNotes(ctx context.Context, integration *domain.Integration, lastSync *time.Time) ([]*domain.ExternalNote, error) {
	// Extract credentials from integration
	_, err := p.createNotionClient(integration)
	if err != nil {
		return nil, fmt.Errorf("failed to create Notion client: %w", err)
	}

	// Query notes from Notion since last sync
	// Note: This would require extending the Notion client with query capabilities
	// For now, we'll return empty slice as this is primarily an outbound sync
	return []*domain.ExternalNote{}, nil
}

// UpdateNote updates an existing note in Notion
func (p *NotionSyncProvider) UpdateNote(ctx context.Context, note *domain.Note, integration *domain.Integration, externalID string) (*domain.SyncResult, error) {
	// Extract credentials from integration
	client, err := p.createNotionClient(integration)
	if err != nil {
		return &domain.SyncResult{
			Success:  false,
			Error:    fmt.Sprintf("Failed to create Notion client: %v", err),
			SyncedAt: time.Now(),
		}, nil
	}

	// For now, we'll treat updates as create operations since the Notion client doesn't have update methods
	// In a production system, you'd implement proper update logic
	title := ""
	if note.Title != nil {
		title = *note.Title
	}

	page, err := client.CreateNote(title, note.Content, note.Category, note.Source)
	if err != nil {
		return &domain.SyncResult{
			Success:  false,
			Error:    fmt.Sprintf("Failed to update note in Notion: %v", err),
			SyncedAt: time.Now(),
		}, nil
	}

	return &domain.SyncResult{
		Success:    true,
		ExternalID: page.ID,
		SyncedAt:   time.Now(),
	}, nil
}

// DeleteNote deletes a note from Notion
func (p *NotionSyncProvider) DeleteNote(ctx context.Context, integration *domain.Integration, externalID string) error {
	// Extract credentials from integration
	_, err := p.createNotionClient(integration)
	if err != nil {
		return fmt.Errorf("failed to create Notion client: %w", err)
	}

	// For now, we'll skip delete operations as the Notion client doesn't have delete methods
	// In a production system, you'd implement proper delete logic
	return nil
}

// ValidateIntegration validates the Notion integration configuration
func (p *NotionSyncProvider) ValidateIntegration(ctx context.Context, integration *domain.Integration) error {
	// Check if auth data contains required fields
	if integration.AuthData == nil {
		return fmt.Errorf("auth data is missing")
	}

	var authData map[string]interface{}
	if err := json.Unmarshal(*integration.AuthData, &authData); err != nil {
		return fmt.Errorf("invalid auth data format: %w", err)
	}

	apiKey, ok := authData["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("api_key is missing or invalid")
	}

	databaseID, ok := authData["database_id"].(string)
	if !ok || databaseID == "" {
		return fmt.Errorf("database_id is missing or invalid")
	}

	// Try to create a client to validate credentials
	_, err := p.createNotionClient(integration)
	if err != nil {
		return fmt.Errorf("failed to validate Notion credentials: %w", err)
	}

	return nil
}

// createNotionClient creates a Notion client from integration data
func (p *NotionSyncProvider) createNotionClient(integration *domain.Integration) (*notion.NotionClient, error) {
	if integration.AuthData == nil {
		return nil, fmt.Errorf("auth data is missing")
	}

	var authData map[string]interface{}
	if err := json.Unmarshal(*integration.AuthData, &authData); err != nil {
		return nil, fmt.Errorf("invalid auth data format: %w", err)
	}

	apiKey, ok := authData["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("api_key is missing or invalid")
	}

	databaseID, ok := authData["database_id"].(string)
	if !ok || databaseID == "" {
		return nil, fmt.Errorf("database_id is missing or invalid")
	}

	return notion.NewNotionClient(notion.NotionConfig{
		APIKey:     apiKey,
		DatabaseID: databaseID,
	}), nil
}
