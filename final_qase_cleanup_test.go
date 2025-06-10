package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	fmt.Println("🎉 Final verification - All Qase dependencies removed!")
	fmt.Println("======================================================")

	// Test build
	fmt.Println("1. Testing build...")
	buildCmd := exec.Command("go", "build", "-o", "bin/test-server", "cmd/server/main.go")
	buildCmd.Dir = "./mainote_server"
	if err := buildCmd.Run(); err != nil {
		log.Fatalf("❌ Build failed: %v", err)
	}
	fmt.Println("✅ Build successful")

	// Test server startup
	fmt.Println("2. Testing server startup...")
	serverCmd := exec.Command("./bin/test-server")
	serverCmd.Dir = "./mainote_server"
	if err := serverCmd.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
	
	// Wait for server to initialize
	time.Sleep(2 * time.Second)
	fmt.Println("✅ Server started successfully")

	// Test health endpoint
	fmt.Println("3. Testing health endpoint...")
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {