package handler

import (
	"context"
	"encoding/json"

	"mainote-server/internal/domain"
	"mainote-server/internal/usecase"
	api "mainote-server/pkg/generated/api"

	"github.com/google/uuid"
)

// IntegrationHandler handles HTTP requests for integrations and implements IntegrationsAPIServicer.
type IntegrationHandler struct {
	integrationUsecase usecase.IntegrationUsecase
}

// NewIntegrationHandler creates a new instance of IntegrationHandler.
func NewIntegrationHandler(integrationUsecase usecase.IntegrationUsecase) *IntegrationHandler {
	return &IntegrationHandler{integrationUsecase: integrationUsecase}
}

// CreateIntegration implements IntegrationsAPIServicer interface.
func (h *IntegrationHandler) CreateIntegration(ctx context.Context, req api.CreateIntegrationRequest) (api.ImplResponse, error) {
	// Parse app_id
	appID, err := uuid.Parse(req.AppId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid app_id format",
		}), nil
	}

	// Convert config if provided
	var config interface{}
	if req.Config != nil {
		config = req.Config
	}

	integration, err := h.integrationUsecase.CreateIntegration(ctx, req.ChatId, appID, req.AuthType, req.AuthData, config)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+req.ChatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}
		if err.Error() == "app not found" {
			return api.Response(404, api.ErrorResponse{
				Error:   "App not found",
				Message: "No app found with the provided app_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain integration to API response
	integrationResponse := domainIntegrationToAPIResponse(integration)
	return api.Response(201, integrationResponse), nil
}

// GetIntegrations implements IntegrationsAPIServicer interface.
func (h *IntegrationHandler) GetIntegrations(ctx context.Context, chatId string, status string, appId string, limit int32, offset int32) (api.ImplResponse, error) {
	// Convert optional parameters
	var statusPtr *string
	var appIDPtr *uuid.UUID

	if status != "" {
		statusPtr = &status
	}

	if appId != "" {
		appUUID, err := uuid.Parse(appId)
		if err != nil {
			return api.Response(400, api.ErrorResponse{
				Error:   "Bad request",
				Message: "Invalid app_id format",
			}), nil
		}
		appIDPtr = &appUUID
	}

	result, err := h.integrationUsecase.GetIntegrationsForUser(ctx, chatId, statusPtr, appIDPtr, int(limit), int(offset))
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
	integrations := make([]api.IntegrationResponse, len(result.Integrations))
	for i, integration := range result.Integrations {
		integrations[i] = domainIntegrationToAPIResponse(&integration)
	}

	listResponse := api.IntegrationsListResponse{
		Integrations: integrations,
		Pagination: api.IntegrationsListResponsePagination{
			TotalCount: int32(result.Total),
			Limit:      int32(result.Limit),
			Offset:     int32(result.Offset),
			HasMore:    result.HasMore,
		},
	}

	return api.Response(200, listResponse), nil
}

// GetIntegrationById implements IntegrationsAPIServicer interface.
func (h *IntegrationHandler) GetIntegrationById(ctx context.Context, integrationId string, chatId string) (api.ImplResponse, error) {
	// Parse integrationId
	integrationUUID, err := uuid.Parse(integrationId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid integration_id format",
		}), nil
	}

	integration, err := h.integrationUsecase.GetIntegrationByID(ctx, integrationUUID, chatId)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+chatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}
		if err.Error() == "integration not found" {
			return api.Response(404, api.ErrorResponse{
				Error:   "Integration not found",
				Message: "No integration found with the provided integration_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain integration to API response
	integrationResponse := domainIntegrationToAPIResponse(integration)
	return api.Response(200, integrationResponse), nil
}

// UpdateIntegration implements IntegrationsAPIServicer interface.
func (h *IntegrationHandler) UpdateIntegration(ctx context.Context, integrationId string, chatId string, req api.UpdateIntegrationRequest) (api.ImplResponse, error) {
	// Parse integrationId
	integrationUUID, err := uuid.Parse(integrationId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid integration_id format",
		}), nil
	}

	// Convert optional fields to pointers
	var status, authType *string
	var authData, config interface{}

	if req.Status != "" {
		status = &req.Status
	}
	if req.AuthType != "" {
		authType = &req.AuthType
	}
	if req.AuthData != nil {
		authData = req.AuthData
	}
	if req.Config != nil {
		config = req.Config
	}

	// Check if at least one field is provided
	if status == nil && authType == nil && authData == nil && config == nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "At least one field must be provided for update",
		}), nil
	}

	integration, err := h.integrationUsecase.UpdateIntegration(ctx, integrationUUID, chatId, status, authType, authData, config)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+chatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}
		if err.Error() == "integration not found" {
			return api.Response(404, api.ErrorResponse{
				Error:   "Integration not found",
				Message: "No integration found with the provided integration_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain integration to API response
	integrationResponse := domainIntegrationToAPIResponse(integration)
	return api.Response(200, integrationResponse), nil
}

// DeleteIntegration implements IntegrationsAPIServicer interface.
func (h *IntegrationHandler) DeleteIntegration(ctx context.Context, integrationId string, chatId string) (api.ImplResponse, error) {
	// Parse integrationId
	integrationUUID, err := uuid.Parse(integrationId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid integration_id format",
		}), nil
	}

	err = h.integrationUsecase.DeleteIntegration(ctx, integrationUUID, chatId)
	if err != nil {
		if err.Error() == "user not found for chat_id: "+chatId {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided chat_id",
			}), nil
		}
		if err.Error() == "integration not found" {
			return api.Response(404, api.ErrorResponse{
				Error:   "Integration not found",
				Message: "No integration found with the provided integration_id",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	return api.Response(204, nil), nil
}

// domainIntegrationToAPIResponse converts domain integration to API response.
func domainIntegrationToAPIResponse(integration *domain.IntegrationWithApp) api.IntegrationResponse {
	// Convert auth_data from json.RawMessage to map[string]interface{}
	authDataMap := make(map[string]interface{})
	if integration.Integration.AuthData != nil {
		var tempAuthData map[string]interface{}
		if err := json.Unmarshal(*integration.Integration.AuthData, &tempAuthData); err == nil {
			// Mask sensitive data
			maskedAuthData := maskSensitiveData(tempAuthData)
			if m, ok := maskedAuthData.(map[string]interface{}); ok {
				authDataMap = m
			}
		}
	}

	// Convert config from json.RawMessage to map[string]interface{}
	configMap := make(map[string]interface{})
	if integration.Integration.Config != nil {
		var tempConfig map[string]interface{}
		if err := json.Unmarshal(*integration.Integration.Config, &tempConfig); err == nil {
			configMap = tempConfig
		}
	}

	return api.IntegrationResponse{
		IntegrationId: integration.Integration.IntegrationID.String(),
		AppId:         integration.Integration.AppID.String(),
		App: api.IntegrationResponseApp{
			AppId:       integration.App.AppID.String(),
			Provider:    integration.App.Provider,
			Name:        integration.App.Name,
			Description: integration.App.Description,
		},
		ChatId:    integration.Integration.ChatID,
		Status:    integration.Integration.Status,
		AuthType:  integration.Integration.AuthType,
		AuthData:  authDataMap,
		Config:    configMap,
		CreatedAt: integration.Integration.CreatedAt,
		UpdatedAt: integration.Integration.UpdatedAt,
	}
}

// maskSensitiveData masks sensitive fields in auth_data.
func maskSensitiveData(authData interface{}) interface{} {
	if authData == nil {
		return nil
	}

	// Convert to map for processing
	var authDataMap map[string]interface{}
	if m, ok := authData.(map[string]interface{}); ok {
		authDataMap = m
	} else {
		// If not already a map, try to convert it
		authDataBytes, err := json.Marshal(authData)
		if err != nil {
			return authData
		}
		if err := json.Unmarshal(authDataBytes, &authDataMap); err != nil {
			return authData
		}
	}

	// Mask sensitive fields
	sensitiveFields := []string{"api_key", "secret", "token", "password", "private_key"}
	for _, field := range sensitiveFields {
		if value, exists := authDataMap[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				authDataMap[field] = "***masked***"
			}
		}
	}

	return authDataMap
}
