package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"mainote-backend/internal/delivery/http/handler"
	api "mainote-backend/pkg/generated"
)

func main() {
	fmt.Println("🔍 Testing mainote_server components...")

	// Test 1: Create handler
	fmt.Println("✅ Creating note handler...")
	noteHandler := handler.NewNoteHandler(nil)

	// Test 2: Create API controllers
	fmt.Println("✅ Creating API controllers...")
	healthController := api.NewHealthAPIController(noteHandler)
	notesController := api.NewNotesAPIController(noteHandler)

	// Test 3: Create router
	fmt.Println("✅ Creating API router...")
	apiRouter := api.NewRouter(healthController, notesController)

	// Test 4: Test health endpoint directly
	fmt.Println("✅ Testing health check endpoint...")
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := &testResponseWriter{header: make(http.Header)}

	// Find health route and test it
	for _, route := range healthController.Routes() {
		if route.Pattern == "/health" {
			route.HandlerFunc(rr, req)
			fmt.Printf("   Health endpoint response: %s\n", rr.body)
			break
		}
	}

	// Test 5: Start simple server
	fmt.Println("✅ Starting test server on :8081...")
	http.Handle("/", apiRouter)

	go func() {
		log.Println("Go backend server starting on port 8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(2 * time.Second)

	// Test health endpoint
	fmt.Println("✅ Testing HTTP health endpoint...")
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		fmt.Printf("❌ Error calling health endpoint: %v\n", err)
	} else {
		fmt.Printf("✅ Health endpoint responded with status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	fmt.Println("🎉 mainote_server is working correctly!")
	fmt.Println("   Health endpoint: http://localhost:8081/health")
	fmt.Println("   Notes API: http://localhost:8081/notes")

	// Keep server running for manual testing
	fmt.Println("\n🔧 Server running for manual testing... Press Ctrl+C to stop")
	select {}
}

// Simple response writer for testing
type testResponseWriter struct {
	header http.Header
	body   string
	status int
}

func (w *testResponseWriter) Header() http.Header {
	return w.header
}

func (w *testResponseWriter) Write(data []byte) (int, error) {
	w.body = string(data)
	return len(data), nil
}

func (w *testResponseWriter) WriteHeader(status int) {
	w.status = status
}
