package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/TomasFernandez123/chatbot/internal/ai"
	"github.com/TomasFernandez123/chatbot/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (ignored in production where env vars are set directly).
	if err := godotenv.Load(); err != nil {
		log.Println("[INIT] No .env file found, using system environment variables")
	}

	// Read configuration from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("[FATAL] GEMINI_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	allowedOriginsRaw := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsRaw == "" {
		allowedOriginsRaw = "http://localhost:4200"
	}
	allowedOrigins := strings.Split(allowedOriginsRaw, ",")
	for i, o := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(o)
	}

	// Initialize AI service
	ctx := context.Background()
	aiService, err := ai.NewService(ctx, apiKey)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize AI service: %v", err)
	}
	defer aiService.Close()

	// Build the HTTP server
	handlers := server.NewHandlers(aiService)
	router := server.NewRouter(handlers, allowedOrigins)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 35 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		log.Printf("[SHUTDOWN] Received signal %v, shutting down gracefully...", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("[FATAL] Server forced to shutdown: %v", err)
		}
	}()

	// Start serving
	log.Printf("╔══════════════════════════════════════════╗")
	log.Printf("║  🤖 Chatbot Backend - Tomas Fernandez   ║")
	log.Printf("║  📡 Listening on port %s                ║", port)
	log.Printf("║  🌐 CORS origins: %v", allowedOrigins)
	log.Printf("╚══════════════════════════════════════════╝")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] Server failed: %v", err)
	}

	log.Println("[SHUTDOWN] Server stopped cleanly")
}
