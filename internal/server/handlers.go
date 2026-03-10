package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/TomasFernandez123/chatbot/internal/ai"
)

// slugRegex only allows safe characters, preventing any path-traversal attack.
var slugRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// masterContext is the fallback used when no project-specific file is found.
const masterContext = `# Tomas Fernandez — Desarrollador de Software

Tomas Fernandez es un desarrollador de software fullstack con experiencia en
tecnologías modernas como Go, Angular, React, Node.js, Docker y servicios cloud.

Ha trabajado en múltiples proyectos propios y freelance, construyendo plataformas
web completas desde el diseño hasta el despliegue en producción.

Si querés conocer más sobre su trabajo o discutir una oportunidad, podés
contactarlo directamente a través de su portfolio o LinkedIn.`

// ChatRequest represents the incoming JSON payload from the frontend.
// The frontend sends only the project slug; the backend resolves the context.
type ChatRequest struct {
	Message string `json:"message"`
	Project string `json:"project"`
}

// ChatResponse is the standardized JSON response.
type ChatResponse struct {
	Answer    string `json:"answer"`
	Timestamp string `json:"timestamp"`
}

// ErrorResponse is the standardized error JSON.
type ErrorResponse struct {
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
}

// Handlers holds the dependencies for HTTP handlers.
type Handlers struct {
	AI          *ai.Service
	contextsDir string
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(aiService *ai.Service, contextsDir string) *Handlers {
	return &Handlers{AI: aiService, contextsDir: contextsDir}
}

// loadContext reads the Markdown documentation file for the given project slug.
// If the file is missing or the slug is invalid, it returns the masterContext fallback.
func (h *Handlers) loadContext(slug string) (string, bool) {
	if !slugRegex.MatchString(slug) {
		return masterContext, false
	}

	// Build and verify the path stays inside the contexts directory.
	filePath := filepath.Join(h.contextsDir, slug+".md")
	absDir, err := filepath.Abs(h.contextsDir)
	if err != nil {
		return masterContext, false
	}
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return masterContext, false
	}
	if !strings.HasPrefix(absFile, absDir+string(filepath.Separator)) {
		return masterContext, false
	}

	data, err := os.ReadFile(absFile)
	if err != nil {
		return masterContext, false
	}
	return string(data), true
}

// HealthCheck responds with a simple status to verify the service is alive.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Chat handles POST /chat requests.
func (h *Handlers) Chat(w http.ResponseWriter, r *http.Request) {
	// Only accept POST
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate and sanitize inputs.
	req.Message = strings.TrimSpace(req.Message)
	req.Project = strings.TrimSpace(req.Project)

	if req.Message == "" {
		writeError(w, "message cannot be empty", http.StatusBadRequest)
		return
	}

	// Resolve project context from disk (RAG local).
	// An invalid/missing slug falls back to the master context silently.
	projectSlug := req.Project
	if projectSlug == "" {
		projectSlug = "unknown"
	}
	projectContext, found := h.loadContext(projectSlug)
	if !found {
		log.Printf("[WARN] Context file not found for project=%q, using master fallback", projectSlug)
	}

	start := time.Now()
	log.Printf("[CHAT] New query | project=%q | ip=%s | question=%q",
		projectSlug, r.RemoteAddr, req.Message)

	// Call the AI service — inherits the request context (handles client disconnect / timeouts).
	answer, err := h.AI.GenerateAnswer(r.Context(), projectContext, req.Message)
	if err != nil {
		log.Printf("[ERROR] AI generation failed | project=%q | elapsed=%s | err=%v",
			projectSlug, time.Since(start).Round(time.Millisecond), err)
		writeError(w, "failed to generate response", http.StatusInternalServerError)
		return
	}

	log.Printf("[CHAT] Response sent | project=%q | elapsed=%s | answer_length=%d",
		projectSlug, time.Since(start).Round(time.Millisecond), len(answer))

	writeJSON(w, http.StatusOK, ChatResponse{
		Answer:    answer,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a standardized error response.
func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, ErrorResponse{
		Error:     message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
