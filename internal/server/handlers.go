package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/TomasFernandez123/chatbot/internal/ai"
)

// ChatRequest represents the incoming JSON payload from the frontend.
type ChatRequest struct {
	Message string `json:"message"`
	Context string `json:"context"`
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
	AI *ai.Service
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(aiService *ai.Service) *Handlers {
	return &Handlers{AI: aiService}
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

	// Validate inputs — don't waste tokens on empty messages
	req.Message = strings.TrimSpace(req.Message)
	req.Context = strings.TrimSpace(req.Context)
	req.Project = strings.TrimSpace(req.Project)

	if req.Message == "" {
		writeError(w, "message cannot be empty", http.StatusBadRequest)
		return
	}

	if req.Context == "" {
		writeError(w, "context cannot be empty", http.StatusBadRequest)
		return
	}

	// Log the consultation
	projectName := req.Project
	if projectName == "" {
		projectName = "unknown"
	}
	log.Printf("[CHAT] New query | project=%q | question=%q | ip=%s",
		projectName, req.Message, r.RemoteAddr)

	// Call the AI service with a timeout context
	answer, err := h.AI.GenerateAnswer(r.Context(), req.Context, req.Message)
	if err != nil {
		log.Printf("[ERROR] AI generation failed: %v", err)
		writeError(w, "failed to generate response", http.StatusInternalServerError)
		return
	}

	log.Printf("[CHAT] Response sent | project=%q | answer_length=%d", projectName, len(answer))

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
