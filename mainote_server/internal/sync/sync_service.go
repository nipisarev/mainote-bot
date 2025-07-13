package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"mainote-server/internal/domain"
	"mainote-server/internal/repository"
)

// SyncServiceImpl implements the SyncService interface
type SyncServiceImpl struct {
	providers           map[string]domain.SyncProvider
	providerMutex       sync.RWMutex
	noteIntegrationRepo repository.NoteIntegrationRepository
	integrationRepo     repository.IntegrationRepository
	noteRepo            repository.NoteRepository
	maxRetries          int
	syncTimeout         time.Duration
}

// NewSyncService creates a new instance of SyncService
func NewSyncService(
	noteIntegrationRepo repository.NoteIntegrationRepository,
	integrationRepo repository.IntegrationRepository,
	noteRepo repository.NoteRepository,
) domain.SyncService {
	service := &SyncServiceImpl{
		providers:           make(map[string]domain.SyncProvider),
		noteIntegrationRepo: noteIntegrationRepo,
		integrationRepo:     integrationRepo,
		noteRepo:            noteRepo,
		maxRetries:          3,
		syncTimeout:         30 * time.Second,
	}

	// Register built-in providers
	service.RegisterProvider(NewNotionSyncProvider())

	return service
}

// RegisterProvider registers a new sync provider
func (s *SyncServiceImpl) RegisterProvider(provider domain.SyncProvider) {
	s.providerMutex.Lock()
	defer s.providerMutex.Unlock()
	s.providers[provider.GetProviderName()] = provider
}

// GetProvider returns a sync provider by name
func (s *SyncServiceImpl) GetProvider(providerName string) (domain.SyncProvider, error) {
	s.providerMutex.RLock()
	defer s.providerMutex.RUnlock()

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("sync provider '%s' not found", providerName)
	}
	return provider, nil
}

// SyncNoteToIntegrations synchronizes a note to all active integrations for a user
func (s *SyncServiceImpl) SyncNoteToIntegrations(ctx context.Context, note *domain.Note) error {
	// Get active integrations for the user
	integrations, err := s.integrationRepo.GetIntegrationsForUser(ctx, note.ChatID, stringPtr("active"), nil, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get integrations for user: %w", err)
	}

	if len(integrations.Integrations) == 0 {
		// No integrations to sync to
		return nil
	}

	// Sync to each integration asynchronously
	var wg sync.WaitGroup
	errorChan := make(chan error, len(integrations.Integrations))

	for _, integration := range integrations.Integrations {
		wg.Add(1)
		go func(integration domain.IntegrationWithApp) {
			defer wg.Done()
			if err := s.syncNoteToIntegration(ctx, note, &integration.Integration); err != nil {
				errorChan <- err
			}
		}(integration)
	}

	wg.Wait()
	close(errorChan)

	// Collect errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync failed for %d integrations: %v", len(errors), errors)
	}

	return nil
}

// syncNoteToIntegration synchronizes a note to a specific integration
func (s *SyncServiceImpl) syncNoteToIntegration(ctx context.Context, note *domain.Note, integration *domain.Integration) error {
	// Check if this note is already synced with this integration
	existingSync, err := s.noteIntegrationRepo.GetByNoteAndIntegration(ctx, note.NoteID, integration.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to check existing sync: %w", err)
	}

	if existingSync != nil {
		// Note already synced, skip
		return nil
	}

	// Get the app information for this integration
	integrationWithApp, err := s.integrationRepo.GetByIDForUser(ctx, integration.IntegrationID, integration.ChatID)
	if err != nil {
		return fmt.Errorf("failed to get integration with app: %w", err)
	}

	// Get the appropriate sync provider
	provider, err := s.GetProvider(integrationWithApp.App.Provider)
	if err != nil {
		return fmt.Errorf("failed to get sync provider for %s: %w", integrationWithApp.App.Provider, err)
	}

	// Create a pending sync record
	noteIntegration := &domain.NoteIntegration{
		NoteID:        note.NoteID,
		IntegrationID: integration.IntegrationID,
		ExternalID:    "",
		SyncStatus:    "pending",
		SyncedAt:      time.Now(),
		RetryCount:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.noteIntegrationRepo.Create(ctx, noteIntegration); err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}

	// Perform the sync
	syncCtx, cancel := context.WithTimeout(ctx, s.syncTimeout)
	defer cancel()

	result, err := provider.SyncNote(syncCtx, note, integration)
	if err != nil {
		// Update sync status to error
		errorMsg := err.Error()
		if updateErr := s.noteIntegrationRepo.UpdateSyncStatus(ctx, noteIntegration.NoteIntegrationID, "error", nil, &errorMsg); updateErr != nil {
			log.Printf("Failed to update sync status to error: %v", updateErr)
		}
		return fmt.Errorf("sync failed: %w", err)
	}

	// Update sync status based on result
	if result.Success {
		if updateErr := s.noteIntegrationRepo.UpdateSyncStatus(ctx, noteIntegration.NoteIntegrationID, "synced", &result.ExternalID, nil); updateErr != nil {
			log.Printf("Failed to update sync status to synced: %v", updateErr)
		}
	} else {
		if updateErr := s.noteIntegrationRepo.UpdateSyncStatus(ctx, noteIntegration.NoteIntegrationID, "error", nil, &result.Error); updateErr != nil {
			log.Printf("Failed to update sync status to error: %v", updateErr)
		}
		return fmt.Errorf("sync failed: %s", result.Error)
	}

	return nil
}

// SyncFromIntegrations synchronizes notes from external systems to internal
func (s *SyncServiceImpl) SyncFromIntegrations(ctx context.Context, chatID string) error {
	// Get active integrations for the user
	integrations, err := s.integrationRepo.GetIntegrationsForUser(ctx, chatID, stringPtr("active"), nil, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get integrations for user: %w", err)
	}

	if len(integrations.Integrations) == 0 {
		// No integrations to sync from
		return nil
	}

	// For each integration, fetch notes and sync them
	for _, integration := range integrations.Integrations {
		if err := s.syncFromIntegration(ctx, &integration.Integration); err != nil {
			log.Printf("Failed to sync from integration %s: %v", integration.Integration.IntegrationID, err)
		}
	}

	return nil
}

// syncFromIntegration synchronizes notes from a specific integration
func (s *SyncServiceImpl) syncFromIntegration(ctx context.Context, integration *domain.Integration) error {
	// Get the app information for this integration
	integrationWithApp, err := s.integrationRepo.GetByIDForUser(ctx, integration.IntegrationID, integration.ChatID)
	if err != nil {
		return fmt.Errorf("failed to get integration with app: %w", err)
	}

	// Get the appropriate sync provider
	provider, err := s.GetProvider(integrationWithApp.App.Provider)
	if err != nil {
		return fmt.Errorf("failed to get sync provider for %s: %w", integrationWithApp.App.Provider, err)
	}

	// Get last sync time (for now, we'll sync all notes)
	var lastSync *time.Time

	// Fetch notes from external system
	syncCtx, cancel := context.WithTimeout(ctx, s.syncTimeout)
	defer cancel()

	externalNotes, err := provider.FetchNotes(syncCtx, integration, lastSync)
	if err != nil {
		return fmt.Errorf("failed to fetch notes from external system: %w", err)
	}

	// Process each external note
	for _, externalNote := range externalNotes {
		if err := s.processExternalNote(ctx, externalNote, integration); err != nil {
			log.Printf("Failed to process external note %s: %v", externalNote.ExternalID, err)
		}
	}

	return nil
}

// processExternalNote processes a note from an external system
func (s *SyncServiceImpl) processExternalNote(ctx context.Context, externalNote *domain.ExternalNote, integration *domain.Integration) error {
	// Check if we already have this note synced
	existingSync, err := s.noteIntegrationRepo.GetByIntegrationID(ctx, integration.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to check existing syncs: %w", err)
	}

	// Look for existing sync with this external ID
	for _, sync := range existingSync {
		if sync.ExternalID == externalNote.ExternalID {
			// Note already exists, skip
			return nil
		}
	}

	// Note: We'd need to get the user_id from chat_id here and create the note
	// For now, we'll skip creating the note as it requires user lookup
	// This would be implemented when we add inbound sync functionality

	return nil
}

// ResyncFailedNotes retries failed sync operations
func (s *SyncServiceImpl) ResyncFailedNotes(ctx context.Context) error {
	// Get failed syncs that haven't exceeded max retries
	failedSyncs, err := s.noteIntegrationRepo.GetFailedSyncs(ctx, s.maxRetries)
	if err != nil {
		return fmt.Errorf("failed to get failed syncs: %w", err)
	}

	for _, noteIntegration := range failedSyncs {
		if err := s.retrySync(ctx, noteIntegration); err != nil {
			log.Printf("Failed to retry sync for note %s: %v", noteIntegration.NoteID, err)
		}
	}

	return nil
}

// retrySync retries a failed sync operation
func (s *SyncServiceImpl) retrySync(ctx context.Context, noteIntegration *domain.NoteIntegration) error {
	// Get the note and integration
	note, err := s.noteRepo.GetByID(ctx, noteIntegration.NoteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	integration, err := s.integrationRepo.GetByID(ctx, noteIntegration.IntegrationID)
	if err != nil {
		return fmt.Errorf("failed to get integration: %w", err)
	}

	// Get the integration with app information
	integrationWithApp, err := s.integrationRepo.GetByIDForUser(ctx, integration.IntegrationID, integration.ChatID)
	if err != nil {
		return fmt.Errorf("failed to get integration with app: %w", err)
	}

	// Get the appropriate sync provider
	provider, err := s.GetProvider(integrationWithApp.App.Provider)
	if err != nil {
		return fmt.Errorf("failed to get sync provider: %w", err)
	}

	// Increment retry count
	noteIntegration.RetryCount++
	noteIntegration.UpdatedAt = time.Now()

	// Perform the sync
	syncCtx, cancel := context.WithTimeout(ctx, s.syncTimeout)
	defer cancel()

	result, err := provider.SyncNote(syncCtx, note, integration)
	if err != nil {
		// Update retry count and error
		errorMsg := err.Error()
		noteIntegration.LastError = &errorMsg
		if updateErr := s.noteIntegrationRepo.Update(ctx, noteIntegration); updateErr != nil {
			log.Printf("Failed to update retry count: %v", updateErr)
		}
		return fmt.Errorf("retry sync failed: %w", err)
	}

	// Update sync status based on result
	if result.Success {
		noteIntegration.SyncStatus = "synced"
		noteIntegration.ExternalID = result.ExternalID
		noteIntegration.SyncedAt = result.SyncedAt
		noteIntegration.LastError = nil
	} else {
		noteIntegration.LastError = &result.Error
	}

	if updateErr := s.noteIntegrationRepo.Update(ctx, noteIntegration); updateErr != nil {
		log.Printf("Failed to update sync record: %v", updateErr)
	}

	return nil
}

// GetSyncStatus returns the sync status for a note
func (s *SyncServiceImpl) GetSyncStatus(ctx context.Context, noteID uuid.UUID) ([]*domain.NoteIntegration, error) {
	return s.noteIntegrationRepo.GetByNoteID(ctx, noteID)
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
