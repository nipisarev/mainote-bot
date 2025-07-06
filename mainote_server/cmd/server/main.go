package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mainote-server/internal/config"
	"mainote-server/internal/delivery/http/handler"
	"mainote-server/internal/delivery/http/middleware"
	"mainote-server/internal/repository"
	"mainote-server/internal/usecase"
	api "mainote-server/pkg/generated/api"

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

	// Initialize use cases
	userUsecase := usecase.NewUserUsecase(userRepo)
	appUsecase := usecase.NewAppUsecase(appRepo)
	noteUsecase := usecase.NewNoteUsecase(noteRepo, userRepo)
	healthUsecase := usecase.NewHealthUseCase()

	// Initialize handlers
	userHandler := handler.NewUserHandler(userUsecase)
	appHandler := handler.NewAppHandler(appUsecase)
	noteHandler := handler.NewNoteHandler(noteUsecase)
	healthHandler := handler.NewHealthHandler(healthUsecase)

	// Setup routes
	healthAPIService := healthHandler
	usersAPIService := userHandler
	appsAPIService := appHandler
	notesAPIService := noteHandler

	healthAPIRouter := api.NewHealthAPIController(healthAPIService)
	usersAPIRouter := api.NewUsersAPIController(usersAPIService)
	appsAPIRouter := api.NewAppsAPIController(appsAPIService)
	notesAPIRouter := api.NewNotesAPIController(notesAPIService)

	router := api.NewRouter(healthAPIRouter, usersAPIRouter, appsAPIRouter, notesAPIRouter)

	// Apply middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.SentryMiddleware)

	// Start server
	log.Println("Starting server on port", cfg.Port)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
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

	// Create a deadline to wait for.
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	server.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
