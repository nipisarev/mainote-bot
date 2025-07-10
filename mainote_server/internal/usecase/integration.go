package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"mainote-server/internal/domain"
	"mainote-server/internal/repository"
)

// IntegrationUsecase defines the interface for integration business logic.
type IntegrationUsecase interface {
	CreateIntegration(ctx context.Context, chatID string, appID uuid.UUID, authType string, authData interface{}, config interface{}) (*domain.IntegrationWithApp, error)
	GetIntegrationByID(ctx context.Context, integrationID uuid.UUID, chatID string) (*domain.IntegrationWithApp, error)
	GetIntegrationsForUser(ctx context.Context, chatID string, status *string, appID *uuid.UUID, limit, offset int) (*domain.IntegrationsListResult, error)
	UpdateIntegration(ctx context.Context, integrationID uuid.UUID, chatID string, status *string, authType *string, authData interface{}, config interface{}) (*domain.IntegrationWithApp, error)
	DeleteIntegration(ctx context.Context, integrationID uuid.UUID, chatID string) error
}

// NewIntegrationUsecase creates a new instance of IntegrationUsecase.
func NewIntegrationUsecase(integrationRepo repository.IntegrationRepository, appRepo repository.AppRepository, userRepo repository.UserRepository) IntegrationUsecase {
	return &integrationUsecase{
		integrationRepo: integrationRepo,
		appRepo:         appRepo,
		userRepo:        userRepo,
	}
}

type integrationUsecase struct {
	integrationRepo repository.IntegrationRepository
	appRepo         repository.AppRepository
	userRepo        repository.UserRepository
}

// CreateIntegration creates a new integration for a user.
func (uc *integrationUsecase) CreateIntegration(ctx context.Context, chatID string, appID uuid.UUID, authType string, authData interface{}, config interface{}) (*domain.IntegrationWithApp, error) {
	// Verify user exists
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		return nil, errors.New("user not found for chat_id: " + chatID)
	}

	// Verify app exists
	app, err := uc.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, errors.New("app not found")
	}

	// Set default config if not provided
	if config == nil {
		config = map[string]interface{}{}
	}

	// Convert config to json.RawMessage
	var configJSON *json.RawMessage
	if config != nil {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return nil, errors.New("invalid config format")
		}
		configJSON = (*json.RawMessage)(&configBytes)
	}

	// Convert authData to json.RawMessage
	var authDataJSON *json.RawMessage
	if authData != nil {
		authDataBytes, err := json.Marshal(authData)
		if err != nil {
			return nil, errors.New("invalid auth_data format")
		}
		authDataJSON = (*json.RawMessage)(&authDataBytes)
	}

	// Create integration
	integration := &domain.Integration{
		IntegrationID: uuid.New(),
		AppID:         appID,
		ChatID:        chatID,
		Status:        "active",
		Config:        configJSON,
		AuthType:      authType,
		AuthData:      authDataJSON,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = uc.integrationRepo.Create(ctx, integration)
	if err != nil {
		return nil, err
	}

	return &domain.IntegrationWithApp{
		Integration: *integration,
		App:         *app,
	}, nil
}

// GetIntegrationByID retrieves an integration by its ID for a specific user.
func (uc *integrationUsecase) GetIntegrationByID(ctx context.Context, integrationID uuid.UUID, chatID string) (*domain.IntegrationWithApp, error) {
	// Verify user exists
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		return nil, errors.New("user not found for chat_id: " + chatID)
	}

	integration, err := uc.integrationRepo.GetByIDForUser(ctx, integrationID, chatID)
	if err != nil {
		return nil, errors.New("integration not found")
	}

	return integration, nil
}

// GetIntegrationsForUser retrieves integrations for a specific user with optional filtering.
func (uc *integrationUsecase) GetIntegrationsForUser(ctx context.Context, chatID string, status *string, appID *uuid.UUID, limit, offset int) (*domain.IntegrationsListResult, error) {
	// Verify user exists
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		return nil, errors.New("user not found for chat_id: " + chatID)
	}

	return uc.integrationRepo.GetIntegrationsForUser(ctx, chatID, status, appID, limit, offset)
}

// UpdateIntegration updates an existing integration for a user.
func (uc *integrationUsecase) UpdateIntegration(ctx context.Context, integrationID uuid.UUID, chatID string, status *string, authType *string, authData interface{}, config interface{}) (*domain.IntegrationWithApp, error) {
	// Verify user exists
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		return nil, errors.New("user not found for chat_id: " + chatID)
	}

	// Get existing integration
	existing, err := uc.integrationRepo.GetByIDForUser(ctx, integrationID, chatID)
	if err != nil {
		return nil, errors.New("integration not found")
	}

	// Update fields if provided
	updatedIntegration := existing.Integration
	if status != nil {
		updatedIntegration.Status = *status
	}
	if authType != nil {
		updatedIntegration.AuthType = *authType
	}
	if authData != nil {
		authDataBytes, err := json.Marshal(authData)
		if err != nil {
			return nil, errors.New("invalid auth_data format")
		}
		updatedIntegration.AuthData = (*json.RawMessage)(&authDataBytes)
	}
	if config != nil {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return nil, errors.New("invalid config format")
		}
		updatedIntegration.Config = (*json.RawMessage)(&configBytes)
	}
	updatedIntegration.UpdatedAt = time.Now()

	err = uc.integrationRepo.UpdateForUser(ctx, integrationID, chatID, &updatedIntegration)
	if err != nil {
		return nil, err
	}

	return &domain.IntegrationWithApp{
		Integration: updatedIntegration,
		App:         existing.App,
	}, nil
}

// DeleteIntegration deletes an integration for a user.
func (uc *integrationUsecase) DeleteIntegration(ctx context.Context, integrationID uuid.UUID, chatID string) error {
	// Verify user exists
	_, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil {
		return errors.New("user not found for chat_id: " + chatID)
	}

	err = uc.integrationRepo.DeleteForUser(ctx, integrationID, chatID)
	if err != nil {
		return errors.New("integration not found")
	}

	return nil
}
