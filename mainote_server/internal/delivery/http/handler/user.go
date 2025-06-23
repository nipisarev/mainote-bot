package handler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"mainote-server/internal/usecase"
	api "mainote-server/pkg/generated/api"

	"golang.org/x/crypto/bcrypt"
)

// UserHandler handles HTTP requests for users and implements UsersAPIServicer.
type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// CreateUser implements UsersAPIServicer interface.
func (h *UserHandler) CreateUser(ctx context.Context, req api.CreateUserRequest) (api.ImplResponse, error) {
	user, err := h.userUsecase.CreateUser(ctx, req.Email, req.Password)
	if err != nil {
		// Handle specific errors
		if err == sql.ErrNoRows {
			return api.Response(409, api.ErrorResponse{
				Error:   "User already exists",
				Message: "A user with this email already exists",
			}), nil
		}
		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain user to API response
	userResponse := api.UserResponse{
		Id:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return api.Response(201, userResponse), nil
}

// AuthUser implements UsersAPIServicer interface for user authentication with chat_id.
func (h *UserHandler) AuthUser(ctx context.Context, req api.AuthUserRequest) (api.ImplResponse, error) {
	userWithSettings, err := h.userUsecase.AuthUser(ctx, req.Email, req.Password, req.ChatId)
	if err != nil {
		if err == sql.ErrNoRows {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user found with the provided email",
			}), nil
		}
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return api.Response(401, api.ErrorResponse{
				Error:   "Invalid credentials",
				Message: "Email or password is incorrect",
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain model to API response
	authResponse := api.AuthUserResponse{
		UserId:    userWithSettings.User.ID.String(),
		Email:     userWithSettings.User.Email,
		ChatId:    userWithSettings.Settings.ChatID,
		CreatedAt: userWithSettings.User.CreatedAt,
		UpdatedAt: userWithSettings.User.UpdatedAt,
	}

	// Return 201 for new user, 200 for existing user
	statusCode := 200
	if userWithSettings.User.CreatedAt.After(time.Now().Add(-1 * time.Minute)) {
		statusCode = 201 // User was just created
	}

	return api.Response(statusCode, authResponse), nil
}

// GetUserByChatId implements UsersAPIServicer interface.
func (h *UserHandler) GetUserByChatId(ctx context.Context, chatId string) (api.ImplResponse, error) {
	userWithSettings, err := h.userUsecase.GetUserByChatID(ctx, chatId)
	if err != nil {
		// Handle user not found
		if err == sql.ErrNoRows {
			return api.Response(404, api.ErrorResponse{
				Error:   "User not found",
				Message: "No user is associated with chat_id " + chatId,
			}), nil
		}

		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}

	// Convert domain model to API response
	userResponse := api.UserResponse{
		Id:        userWithSettings.User.ID.String(),
		Email:     userWithSettings.User.Email,
		ChatId:    userWithSettings.Settings.ChatID,
		CreatedAt: userWithSettings.User.CreatedAt,
		UpdatedAt: userWithSettings.User.UpdatedAt,
	}

	return api.Response(200, userResponse), nil
}

// UpdateUserSettings implements UsersAPIServicer interface.
func (h *UserHandler) UpdateUserSettings(ctx context.Context, chatId string, req api.UserSettingsRequest) (api.ImplResponse, error) {
	// Convert string fields to pointers for partial updates
	var morningNotificationTime *string
	var timezone *string

	if req.MorningNotificationTime != "" {
		morningNotificationTime = &req.MorningNotificationTime
	}

	if req.Timezone != "" {
		timezone = &req.Timezone
	}

	// Update user settings
	settings, err := h.userUsecase.UpdateUserSettings(ctx, chatId, morningNotificationTime, timezone)
	if err != nil {
		if err.Error() == "no fields provided for update" {
			return api.Response(400, api.ErrorResponse{
				Error:   "Bad request",
				Message: "At least one field must be provided for update",
			}), nil
		}
		if err.Error() == fmt.Sprintf("user settings not found for chat_id: %s", chatId) {
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

	// Convert domain model to API response
	settingsResponse := api.UserSettingsResponse{
		UserSettingsId:          int32(settings.UserSettingsID),
		UserId:                  settings.UserID.String(),
		ChatId:                  settings.ChatID,
		MorningNotificationTime: settings.MorningNotificationTime,
		Timezone:                settings.Timezone,
		CreatedAt:               settings.CreatedAt,
		UpdatedAt:               settings.UpdatedAt,
	}

	return api.Response(200, settingsResponse), nil
}

// CreateUserSettings implements UsersAPIServicer interface for creating user settings.
func (h *UserHandler) CreateUserSettings(ctx context.Context, req api.UserSettingsRequest) (api.ImplResponse, error) {
	// Parse userID from string to UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "Invalid user_id format",
		}), nil
	}

	// Convert string fields to pointers for partial insert
	var morningNotificationTime *string
	var timezone *string
	if req.MorningNotificationTime != "" {
		morningNotificationTime = &req.MorningNotificationTime
	}
	if req.Timezone != "" {
		timezone = &req.Timezone
	}
	if req.ChatId == "" {
		return api.Response(400, api.ErrorResponse{
			Error:   "Bad request",
			Message: "chat_id is required",
		}), nil
	}
	
	settings, err := h.userUsecase.CreateUserSettings(ctx, userID, req.ChatId, morningNotificationTime, timezone)
	if err != nil {
		return api.Response(500, api.ErrorResponse{
			Error:   "Internal server error",
			Message: err.Error(),
		}), nil
	}
	
	settingsResponse := api.UserSettingsResponse{
		UserSettingsId:          int32(settings.UserSettingsID),
		UserId:                  settings.UserID.String(),
		ChatId:                  settings.ChatID,
		MorningNotificationTime: settings.MorningNotificationTime,
		Timezone:                settings.Timezone,
		CreatedAt:               settings.CreatedAt,
		UpdatedAt:               settings.UpdatedAt,
	}
	return api.Response(201, settingsResponse), nil
}
