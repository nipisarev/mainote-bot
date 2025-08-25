package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mainote-server/internal/ai"
	"mainote-server/internal/config"
	"mainote-server/internal/delivery/http/handler"
	"mainote-server/internal/delivery/http/middleware"
	"mainote-server/internal/domain"
	"mainote-server/internal/notifications"
	"mainote-server/internal/repository"
	"mainote-server/internal/sync"
	"mainote-server/internal/usecase"
	api "mainote-server/pkg/generated/api"

	"github.com/google/uuid"

	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// Initialize Sentry
	err = sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.Environment,
	})
	if err != nil {
		log.Printf("Sentry initialization failed: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	appRepo := repository.NewAppRepository(db)
	noteRepo := repository.NewNoteRepository(db)
	integrationRepo := repository.NewIntegrationRepository(db)
	noteIntegrationRepo := repository.NewNoteIntegrationRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// Initialize sync service
	syncService := sync.NewSyncService(noteIntegrationRepo, integrationRepo, noteRepo)

	// Initialize AI service and usecase
	var aiUsecase domain.AIUsecase
	if cfg.OpenAIAPIKey != "" {
		aiService := ai.NewOpenAIService(cfg.OpenAIAPIKey)
		aiUsecase = usecase.NewAIUsecase(aiService)
		log.Println("AI service initialized with OpenAI")
	} else {
		log.Println("OpenAI API key not provided - AI features disabled")
	}

	// Initialize notification scheduler
	scheduler := notifications.NewScheduler(notificationRepo, userRepo, noteRepo)

	// Inject syncService into scheduler's formatter so morning notifications can pull events
	// Note: this sets the internal formatter's syncService field to the running syncService instance
	// by creating a new formatter with syncService and replacing the scheduler.formatter.
	schedulerFormatter := notifications.NewMorningNotificationFormatter(noteRepo, syncService)
	scheduler.SetFormatter(schedulerFormatter)

	// Start notification scheduler in background
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	go func() {
		log.Println("Starting notification scheduler")
		scheduler.StartWorker(schedulerCtx)
		log.Println("Notification scheduler stopped")
	}()

	// Initialize use cases
	userUsecase := usecase.NewUserUsecase(userRepo)
	appUsecase := usecase.NewAppUsecase(appRepo)
	noteUsecase := usecase.NewNoteUsecase(noteRepo, userRepo, syncService, aiUsecase)
	integrationUsecase := usecase.NewIntegrationUsecase(integrationRepo, appRepo, userRepo)
	healthUsecase := usecase.NewHealthUseCase()

	// Initialize handlers
	userHandler := handler.NewUserHandler(userUsecase)
	appHandler := handler.NewAppHandler(appUsecase)
	noteHandler := handler.NewNoteHandler(noteUsecase)
	integrationHandler := handler.NewIntegrationHandler(integrationUsecase)
	healthHandler := handler.NewHealthHandler(healthUsecase)

	// Setup routes
	healthAPIService := healthHandler
	usersAPIService := userHandler
	appsAPIService := appHandler
	notesAPIService := noteHandler
	integrationsAPIService := integrationHandler

	healthAPIRouter := api.NewHealthAPIController(healthAPIService)
	usersAPIRouter := api.NewUsersAPIController(usersAPIService)
	appsAPIRouter := api.NewAppsAPIController(appsAPIService)
	notesAPIRouter := api.NewNotesAPIController(notesAPIService)
	integrationsAPIRouter := api.NewIntegrationsAPIController(integrationsAPIService)

	router := api.NewRouter(healthAPIRouter, usersAPIRouter, appsAPIRouter, notesAPIRouter, integrationsAPIRouter)

	// Apply middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.SentryMiddleware)

	// Create a mux so we can register custom admin endpoints alongside the generated router
	mux := http.NewServeMux()

	// Debug endpoint to test routing
	mux.HandleFunc("/debug/routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Routes are working! Users API should be at /api/v1/users/chat/{chat_id}"))
	})

	// Register the main router for all other routes
	mux.Handle("/", router)

	// Admin endpoint to trigger a morning notification for a specific chat_id or user_id
	mux.HandleFunc("/api/v1/admin/notifications/morning", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		type requestBody struct {
			ChatID string `json:"chat_id"`
			UserID string `json:"user_id"`
		}

		var body requestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		// Resolve chat_id from user_id if provided
		resolvedChatID := body.ChatID
		if resolvedChatID == "" {
			if body.UserID == "" {
				http.Error(w, "chat_id or user_id is required", http.StatusBadRequest)
				return
			}

			// Parse user id as UUID
			uid, err := uuid.Parse(body.UserID)
			if err != nil {
				http.Error(w, "invalid user_id format", http.StatusBadRequest)
				return
			}

			settings, err := userRepo.FindSettingsByUserID(r.Context(), uid)
			if err != nil {
				log.Printf("failed to find user settings by user_id %s: %v", body.UserID, err)
				http.Error(w, "failed to find user settings for provided user_id", http.StatusNotFound)
				return
			}

			resolvedChatID = settings.ChatID
			if resolvedChatID == "" {
				http.Error(w, "no chat_id associated with the provided user_id", http.StatusNotFound)
				return
			}
		}

		// Call scheduler to send notification immediately
		if err := scheduler.SendMorningNotificationForChat(r.Context(), resolvedChatID); err != nil {
			log.Printf("failed to send morning notification: %v", err)
			http.Error(w, "failed to send morning notification", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Start server
	log.Println("Starting server on port", cfg.Port)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", cfg.Port, err)
		}
	}()

	// Listen for an interrupt or termination signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Received shutdown signal")

	// Stop notification scheduler first
	schedulerCancel()

	// Create a deadline to wait for.
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
