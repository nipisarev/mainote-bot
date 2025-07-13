package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"mainote-server/internal/domain"
	"mainote-server/internal/repository"
)

// Scheduler handles scheduling and sending notifications
type Scheduler struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	noteRepo         repository.NoteRepository
	formatter        *MorningNotificationFormatter
	botURL           string
	internalAPIKey   string
	cron             *cron.Cron
	httpClient       *http.Client
}

// NewScheduler creates a new notification scheduler
func NewScheduler(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	noteRepo repository.NoteRepository,
) *Scheduler {
	formatter := NewMorningNotificationFormatter(noteRepo)
	return &Scheduler{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		noteRepo:         noteRepo,
		formatter:        formatter,
		botURL:           getEnv("MAINOTE_BOT_URL", "http://mainote-bot:8080"),
		internalAPIKey:   getEnv("INTERNAL_API_KEY", ""),
		cron:             cron.New(),
		httpClient:       &http.Client{Timeout: 30 * time.Second},
	}
}

// Schedule schedules a notification to be sent at a specific user local time
func (s *Scheduler) Schedule(ctx context.Context, userID string, message string, userLocalTime time.Time) error {
	// Parse userID as UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Get user settings including timezone by user ID
	userSettings, err := s.userRepo.FindSettingsByUserID(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("failed to find user settings: %w", err)
	}

	// Convert user local time to UTC
	utcTime, err := s.convertToUTC(userLocalTime, userSettings.Timezone)
	if err != nil {
		return fmt.Errorf("failed to convert time to UTC: %w", err)
	}

	// Create scheduled notification
	notification := &domain.ScheduledNotification{
		ID:        uuid.New(),
		UserID:    userUUID,
		Message:   message,
		SendAt:    utcTime,
		Delivered: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	err = s.notificationRepo.Create(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Info().
		Str("notification_id", notification.ID.String()).
		Str("user_id", userID).
		Time("send_at_utc", utcTime).
		Time("send_at_local", userLocalTime).
		Msg("Scheduled notification created")

	return nil
}

// StartWorker starts the background worker that processes notifications
func (s *Scheduler) StartWorker(ctx context.Context) {
	// Add cron job to run every minute
	_, err := s.cron.AddFunc("* * * * *", func() {
		if err := s.processNotifications(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to process notifications")
		}
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to add cron job")
	}

	// Add cron job to schedule morning notifications every day at midnight
	_, err = s.cron.AddFunc("* * * * *", func() {
		if err := s.scheduleMorningNotifications(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to schedule morning notifications")
		}
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to add morning notification cron job")
	}

	// Schedule morning notifications immediately on startup
	go func() {
		if err := s.scheduleMorningNotifications(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to schedule morning notifications on startup")
		}
	}()

	// Start the cron scheduler
	s.cron.Start()
	log.Info().Msg("Notification scheduler started")

	// Wait for context cancellation
	<-ctx.Done()
	s.cron.Stop()
	log.Info().Msg("Notification scheduler stopped")
}

// processNotifications processes all pending notifications
func (s *Scheduler) processNotifications(ctx context.Context) error {
	// Get all notifications that should be sent now
	notifications, err := s.notificationRepo.GetPendingNotifications(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	if len(notifications) == 0 {
		return nil
	}

	log.Info().Int("count", len(notifications)).Msg("Processing pending notifications")

	// Process each notification
	for _, notification := range notifications {
		if err := s.sendNotification(ctx, notification); err != nil {
			log.Error().
				Err(err).
				Str("notification_id", notification.ID.String()).
				Msg("Failed to send notification")
			continue
		}

		// Mark as delivered
		if err := s.notificationRepo.MarkAsDelivered(ctx, notification.ID); err != nil {
			log.Error().
				Err(err).
				Str("notification_id", notification.ID.String()).
				Msg("Failed to mark notification as delivered")
		}

		// Schedule next day's morning notification if this was a morning notification
		if err := s.scheduleNextMorningNotification(ctx, notification); err != nil {
			log.Error().
				Err(err).
				Str("notification_id", notification.ID.String()).
				Str("user_id", notification.UserID.String()).
				Msg("Failed to schedule next morning notification")
		}
	}

	return nil
}

// scheduleMorningNotifications schedules morning notifications for all users who have morning_notification_time set
func (s *Scheduler) scheduleMorningNotifications(ctx context.Context) error {
	log.Info().Msg("Scheduling morning notifications for users")

	// Get all users with morning notification time set
	users, err := s.userRepo.FindUsersWithMorningNotifications(ctx)
	if err != nil {
		return fmt.Errorf("failed to find users with morning notifications: %w", err)
	}

	if len(users) == 0 {
		log.Info().Msg("No users with morning notifications found")
		return nil
	}

	log.Info().Int("user_count", len(users)).Msg("Found users with morning notifications")

	// Schedule morning notification for each user
	for _, userWithSettings := range users {
		if err := s.scheduleMorningNotificationForUser(ctx, userWithSettings); err != nil {
			log.Error().
				Err(err).
				Str("user_id", userWithSettings.User.ID.String()).
				Str("chat_id", userWithSettings.Settings.ChatID).
				Msg("Failed to schedule morning notification for user")
			continue
		}
	}

	return nil
}

// scheduleMorningNotificationForUser schedules a morning notification for a specific user
func (s *Scheduler) scheduleMorningNotificationForUser(ctx context.Context, userWithSettings domain.UserWithSettings) error {
	if userWithSettings.Settings.MorningNotificationTime == nil {
		return fmt.Errorf("user has no morning notification time set")
	}

	// Parse the morning notification time (format: "HH:MM")
	timeStr := *userWithSettings.Settings.MorningNotificationTime
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		return fmt.Errorf("invalid morning notification time format: %s", timeStr)
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return fmt.Errorf("invalid hour in morning notification time: %s", timeParts[0])
	}

	minute, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return fmt.Errorf("invalid minute in morning notification time: %s", timeParts[1])
	}

	// Get user's timezone
	var location *time.Location
	if userWithSettings.Settings.Timezone != nil && *userWithSettings.Settings.Timezone != "" {
		location, err = time.LoadLocation(*userWithSettings.Settings.Timezone)
		if err != nil {
			log.Warn().
				Str("user_id", userWithSettings.User.ID.String()).
				Str("timezone", *userWithSettings.Settings.Timezone).
				Msg("Invalid timezone, using UTC")
			location = time.UTC
		}
	} else {
		location = time.UTC
	}

	// Calculate next morning notification time
	now := time.Now().In(location)
	nextMorning := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, location)

	// If the time has already passed today, schedule for tomorrow
	if nextMorning.Before(now) {
		nextMorning = nextMorning.AddDate(0, 0, 1)
	}

	// Check if there's already a future morning notification scheduled for this user
	hasExistingNotification, err := s.hasFutureMorningNotification(ctx, userWithSettings.User.ID, nextMorning)
	if err != nil {
		log.Warn().
			Err(err).
			Str("user_id", userWithSettings.User.ID.String()).
			Msg("Failed to check for existing notifications, continuing with scheduling")
	} else if hasExistingNotification {
		log.Info().
			Str("user_id", userWithSettings.User.ID.String()).
			Time("scheduled_time", nextMorning).
			Msg("Morning notification already scheduled for user, skipping")
		return nil
	}

	// Cancel any existing morning notifications with different times
	err = s.cancelMorningNotificationsWithDifferentTime(ctx, userWithSettings.User.ID, nextMorning)
	if err != nil {
		log.Warn().
			Err(err).
			Str("user_id", userWithSettings.User.ID.String()).
			Msg("Failed to cancel old notifications with different time, continuing with scheduling")
	}

	// Generate morning notification message
	message := s.formatter.GenerateMessage(ctx, userWithSettings)

	// Schedule the notification
	err = s.Schedule(ctx, userWithSettings.User.ID.String(), message, nextMorning)
	if err != nil {
		return fmt.Errorf("failed to schedule morning notification: %w", err)
	}

	log.Info().
		Str("user_id", userWithSettings.User.ID.String()).
		Str("chat_id", userWithSettings.Settings.ChatID).
		Time("scheduled_time", nextMorning).
		Str("timezone", location.String()).
		Msg("Morning notification scheduled")

	return nil
}

// cancelPendingNotificationsForUser cancels all pending notifications for a specific user
func (s *Scheduler) cancelPendingNotificationsForUser(ctx context.Context, userID uuid.UUID) error {
	// Get all pending notifications for this user (within next 7 days to be safe)
	endTime := time.Now().Add(7 * 24 * time.Hour)
	pendingNotifications, err := s.notificationRepo.GetPendingNotifications(ctx, endTime)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	// Cancel notifications for this specific user
	cancelledCount := 0
	for _, notification := range pendingNotifications {
		if notification.UserID == userID {
			// Mark as delivered to effectively cancel it
			err := s.notificationRepo.MarkAsDelivered(ctx, notification.ID)
			if err != nil {
				log.Warn().
					Err(err).
					Str("notification_id", notification.ID.String()).
					Msg("Failed to cancel notification")
				continue
			}
			cancelledCount++
		}
	}

	if cancelledCount > 0 {
		log.Info().
			Str("user_id", userID.String()).
			Int("cancelled_count", cancelledCount).
			Msg("Cancelled existing pending notifications for user")
	}

	return nil
}

// hasFutureMorningNotification checks if there's already a future morning notification scheduled for the user
func (s *Scheduler) hasFutureMorningNotification(ctx context.Context, userID uuid.UUID, targetTime time.Time) (bool, error) {
	// Get all pending notifications for this user (within next 7 days)
	endTime := time.Now().Add(7 * 24 * time.Hour)
	pendingNotifications, err := s.notificationRepo.GetPendingNotifications(ctx, endTime)
	if err != nil {
		return false, fmt.Errorf("failed to get pending notifications: %w", err)
	}

	// Check if there's already a morning notification scheduled for this user
	for _, notification := range pendingNotifications {
		if notification.UserID == userID && strings.Contains(notification.Message, "🌅 Good morning!") {
			// Check if it's scheduled for around the same time (within 2 hours)
			timeDiff := notification.SendAt.Sub(targetTime)
			if timeDiff >= -1*time.Hour && timeDiff <= 1*time.Hour {
				return true, nil
			}
		}
	}

	return false, nil
}

// cancelMorningNotificationsWithDifferentTime cancels existing morning notifications that are scheduled for a different time
func (s *Scheduler) cancelMorningNotificationsWithDifferentTime(ctx context.Context, userID uuid.UUID, newTime time.Time) error {
	// Get all pending notifications for this user (within next 7 days)
	endTime := time.Now().Add(7 * 24 * time.Hour)
	pendingNotifications, err := s.notificationRepo.GetPendingNotifications(ctx, endTime)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	// Cancel morning notifications with different times
	cancelledCount := 0
	for _, notification := range pendingNotifications {
		if notification.UserID == userID && strings.Contains(notification.Message, "🌅 Good morning!") {
			// Check if it's scheduled for a different time (more than 1 hour difference)
			timeDiff := notification.SendAt.Sub(newTime)
			if timeDiff < -1*time.Hour || timeDiff > 1*time.Hour {
				// Cancel this notification (mark as delivered)
				err := s.notificationRepo.MarkAsDelivered(ctx, notification.ID)
				if err != nil {
					log.Warn().
						Err(err).
						Str("notification_id", notification.ID.String()).
						Msg("Failed to cancel notification with different time")
					continue
				}
				cancelledCount++
			}
		}
	}

	if cancelledCount > 0 {
		log.Info().
			Str("user_id", userID.String()).
			Int("cancelled_count", cancelledCount).
			Time("new_time", newTime).
			Msg("Cancelled morning notifications with different time")
	}

	return nil
}

// scheduleNextMorningNotification schedules the next day's morning notification if this was a morning notification
func (s *Scheduler) scheduleNextMorningNotification(ctx context.Context, deliveredNotification domain.ScheduledNotification) error {
	// Check if this was a morning notification by looking at the message content
	if !strings.Contains(deliveredNotification.Message, "🌅 Good morning!") {
		// Not a morning notification, skip
		return nil
	}

	// Get user settings including morning notification time
	userSettings, err := s.userRepo.FindSettingsByUserID(ctx, deliveredNotification.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user settings: %w", err)
	}

	// Check if user still has morning notifications enabled
	if userSettings.MorningNotificationTime == nil {
		log.Info().
			Str("user_id", deliveredNotification.UserID.String()).
			Msg("User no longer has morning notifications enabled, skipping next day scheduling")
		return nil
	}

	// Get user info
	user, err := s.userRepo.FindByID(ctx, deliveredNotification.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Create UserWithSettings object
	userWithSettings := domain.UserWithSettings{
		User:     *user,
		Settings: *userSettings,
	}

	// Calculate next morning notification time (tomorrow)
	timeStr := *userSettings.MorningNotificationTime
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		return fmt.Errorf("invalid morning notification time format: %s", timeStr)
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return fmt.Errorf("invalid hour in morning notification time: %s", timeParts[0])
	}

	minute, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return fmt.Errorf("invalid minute in morning notification time: %s", timeParts[1])
	}

	// Get user's timezone
	var location *time.Location
	if userSettings.Timezone != nil && *userSettings.Timezone != "" {
		location, err = time.LoadLocation(*userSettings.Timezone)
		if err != nil {
			log.Warn().
				Str("user_id", deliveredNotification.UserID.String()).
				Str("timezone", *userSettings.Timezone).
				Msg("Invalid timezone, using UTC")
			location = time.UTC
		}
	} else {
		location = time.UTC
	}

	// Calculate tomorrow's morning notification time
	now := time.Now().In(location)
	nextMorning := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, location)
	nextMorning = nextMorning.AddDate(0, 0, 1) // Always schedule for tomorrow

	// Generate morning notification message
	message := s.formatter.GenerateMessage(ctx, userWithSettings)

	// Schedule the notification
	err = s.Schedule(ctx, deliveredNotification.UserID.String(), message, nextMorning)
	if err != nil {
		return fmt.Errorf("failed to schedule next morning notification: %w", err)
	}

	log.Info().
		Str("user_id", deliveredNotification.UserID.String()).
		Time("scheduled_time", nextMorning).
		Str("timezone", location.String()).
		Msg("Next morning notification scheduled")

	return nil
}

// sendNotification sends a notification to mainote_bot
func (s *Scheduler) sendNotification(ctx context.Context, notification domain.ScheduledNotification) error {
	// Get user settings to find their chat_id
	chatID, err := s.getChatIDForUser(ctx, notification.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("notification_id", notification.ID.String()).
			Str("user_id", notification.UserID.String()).
			Msg("Failed to get chat_id for user")
		return fmt.Errorf("failed to get chat_id for user: %w", err)
	}

	if chatID == "" {
		log.Error().
			Str("notification_id", notification.ID.String()).
			Str("user_id", notification.UserID.String()).
			Msg("User has no chat_id")
		return fmt.Errorf("user has no chat_id")
	}

	log.Info().
		Str("notification_id", notification.ID.String()).
		Str("user_id", notification.UserID.String()).
		Str("chat_id", chatID).
		Str("bot_url", s.botURL).
		Str("api_key_set", fmt.Sprintf("%t", s.internalAPIKey != "")).
		Msg("Attempting to send notification to bot")

	// Create payload
	payload := domain.NotificationPayload{
		ChatID:  chatID,
		Type:    "morning_notification",
		Message: notification.Message,
		ID:      notification.ID.String(),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("notification_id", notification.ID.String()).
			Msg("Failed to marshal payload")
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	log.Info().
		Str("notification_id", notification.ID.String()).
		Str("payload", string(jsonData)).
		Msg("Sending notification payload")

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.botURL+"/notification", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().
			Err(err).
			Str("notification_id", notification.ID.String()).
			Str("url", s.botURL+"/notification").
			Msg("Failed to create HTTP request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-API-Key", s.internalAPIKey)

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Error().
			Err(err).
			Str("notification_id", notification.ID.String()).
			Str("url", s.botURL+"/notification").
			Msg("Failed to send HTTP request to bot")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	respBody := make([]byte, 1024)
	n, _ := resp.Body.Read(respBody)
	respBodyStr := string(respBody[:n])

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("response_body", respBodyStr).
			Str("notification_id", notification.ID.String()).
			Str("url", s.botURL+"/notification").
			Msg("Bot returned non-200 status")
		return fmt.Errorf("bot returned non-200 status: %d - %s", resp.StatusCode, respBodyStr)
	}

	log.Info().
		Str("notification_id", notification.ID.String()).
		Str("chat_id", chatID).
		Int("status_code", resp.StatusCode).
		Str("response_body", respBodyStr).
		Msg("Notification sent successfully")

	return nil
}

// convertToUTC converts user local time to UTC using their timezone
func (s *Scheduler) convertToUTC(userLocalTime time.Time, timezone *string) (time.Time, error) {
	if timezone == nil || *timezone == "" {
		// Default to UTC if no timezone is set
		return userLocalTime.UTC(), nil
	}

	// Load the timezone
	loc, err := time.LoadLocation(*timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", *timezone, err)
	}

	// Convert to UTC
	utcTime := userLocalTime.In(loc).UTC()
	return utcTime, nil
}

// getChatIDForUser gets the chat_id for a user from their settings
func (s *Scheduler) getChatIDForUser(ctx context.Context, userID uuid.UUID) (string, error) {
	settings, err := s.userRepo.FindSettingsByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to find user settings: %w", err)
	}
	return settings.ChatID, nil
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
