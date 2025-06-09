package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mainote-backend/internal/config"
	"mainote-backend/internal/delivery/http/handler"
	"mainote-backend/internal/delivery/http/middleware"
	api "mainote-backend/pkg/generated"

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.Environment,
	})
	if err != nil {
		log.Printf("Sentry initialization failed: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Initialize use cases
	// healthUseCase := usecase.NewHealthUseCase() // TODO: Implement this
	// noteUsecase := usecase.NewNoteUsecase(nil)  // TODO: Implement repository

	// Initialize handlers
	// healthHandler := handler.NewHealthHandler(healthUseCase) // TODO: Implement this
	noteHandler := handler.NewNoteHandler(nil) // TODO: Pass proper usecase

	// Setup routes
	router := mux.NewRouter()

	// Apply middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.SentryMiddleware)

	// Create API controllers with the new generated interfaces
	healthController := api.NewHealthAPIController(noteHandler)
	notesController := api.NewNotesAPIController(noteHandler)

	// Register API routes using the new generated router
	apiRouter := api.NewRouter(healthController, notesController)

	// Mount the API routes under the main router
	router.PathPrefix("/").Handler(apiRouter)

	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Go backend server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
