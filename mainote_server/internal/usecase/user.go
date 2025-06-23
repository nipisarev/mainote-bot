package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"mainote-server/internal/domain"
	"mainote-server/internal/repository"
)

// UserUsecase defines the interface for user business logic.
type UserUsecase interface {
	CreateUser(ctx context.Context, email, password string) (*domain.User, error)
	AuthUser(ctx context.Context, email, password, chatID string) (*domain.UserWithSettings, error)
	GetUserByChatID(ctx context.Context, chatID string) (*domain.UserWithSettings, error)
	UpdateUserSettings(ctx context.Context, chatID string, morningNotificationTime, timezone *string) (*domain.UserSettings, error)
	CreateUserSettings(ctx context.Context, userID uuid.UUID, chatID string, morningNotificationTime, timezone *string) (*domain.UserSettings, error)
}

// NewUserUsecase creates a new instance of UserUsecase.
func NewUserUsecase(userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

type userUsecase struct {
	userRepo repository.UserRepository
}

// CreateUser creates a new user.
func (uc *userUsecase) CreateUser(ctx context.Context, email, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthUser authenticates existing user or creates new user and links to chat_id.
func (uc *userUsecase) AuthUser(ctx context.Context, email, password, chatID string) (*domain.UserWithSettings, error) {
	// First, check if user exists with this chat_id
	existingUserWithSettings, err := uc.userRepo.FindByChatID(ctx, chatID)
	if err != nil && err != sql.ErrNoRows {
		// Database error
		return nil, err
	}

	// If user exists with this chat_id, verify password
	if err != sql.ErrNoRows && existingUserWithSettings != nil {
		err = bcrypt.CompareHashAndPassword([]byte(existingUserWithSettings.User.Password), []byte(password))
		if err != nil {
			return nil, err // Invalid password
		}
		return existingUserWithSettings, nil
	}

	// Check if user exists by email
	existingUser, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// User doesn't exist or database error
		return nil, err
	}

	// User exists, verify password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password))
	if err != nil {
		return nil, err // Invalid password
	}

	// Return user with updated settings
	return &domain.UserWithSettings{
		User: *existingUser,
		Settings: domain.UserSettings{
			UserID:    existingUser.ID,
			ChatID:    chatID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}, nil
}

// GetUserByChatID retrieves a user by their chat ID.
func (uc *userUsecase) GetUserByChatID(ctx context.Context, chatID string) (*domain.UserWithSettings, error) {
	return uc.userRepo.FindByChatID(ctx, chatID)
}

// UpdateUserSettings updates user settings for a given chat ID.
func (uc *userUsecase) UpdateUserSettings(ctx context.Context, chatID string, morningNotificationTime, timezone *string) (*domain.UserSettings, error) {
	// Create update request with only non-nil fields
	update := &domain.UserSettings{}

	update.ChatID = chatID

	if morningNotificationTime != nil && *morningNotificationTime != "" {
		update.MorningNotificationTime = morningNotificationTime
	}

	if timezone != nil && *timezone != "" {
		update.Timezone = timezone
	}

	// Check if at least one field is provided for update
	if update.MorningNotificationTime == nil && update.Timezone == nil {
		return nil, fmt.Errorf("no fields provided for update")
	}

	return uc.userRepo.UpdateUserSettings(ctx, update)
}

// CreateUserSettings creates user settings for a given user ID and chat ID.
func (uc *userUsecase) CreateUserSettings(ctx context.Context, userID uuid.UUID, chatID string, morningNotificationTime, timezone *string) (*domain.UserSettings, error) {
	// Use the repository to create the settings with the provided userID
	return uc.userRepo.CreateUserSettings(ctx, userID, chatID, morningNotificationTime, timezone)
}
