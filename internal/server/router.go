package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates and configures the Chi router with all middleware and routes.
func NewRouter(handlers *Handlers, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()

	// --- Middleware Stack ---

	// Request ID for tracing
	r.Use(middleware.RequestID)

	// Real IP from proxy headers
	r.Use(middleware.RealIP)

	// Structured logger
	r.Use(middleware.Logger)

	// Recover from panics gracefully
	r.Use(middleware.Recoverer)

	// Timeout: cut connections if AI takes too long (30s)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS: only allow configured origins
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Security headers
	r.Use(securityHeaders)

	// --- Routes ---
	r.Get("/health", handlers.HealthCheck)
	r.Post("/chat", handlers.Chat)

	return r
}

// securityHeaders adds basic security headers to every response.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
