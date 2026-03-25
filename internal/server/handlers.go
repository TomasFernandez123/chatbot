package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/TomasFernandez123/chatbot/internal/ai"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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

	generateAnswer       func(ctx context.Context, projectContext, question string) (string, error)
	generateAnswerStream func(ctx context.Context, projectContext, question string, onChunk func(string) error) error
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(aiService *ai.Service, contextsDir string) *Handlers {
	return &Handlers{
		AI:          aiService,
		contextsDir: contextsDir,
		generateAnswer: func(ctx context.Context, projectContext, question string) (string, error) {
			if aiService == nil {
				return "", errors.New("service unavailable")
			}
			return aiService.GenerateAnswer(ctx, projectContext, question)
		},
		generateAnswerStream: func(ctx context.Context, projectContext, question string, onChunk func(string) error) error {
			if aiService == nil {
				return errors.New("service unavailable")
			}
			return aiService.GenerateAnswerStream(ctx, projectContext, question, onChunk)
		},
	}
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

	if r.Header.Get("Accept") == "application/x-ndjson" {
		h.chatStream(w, r)
		return
	}

	h.chatClassic(w, r)
}

func (h *Handlers) chatClassic(w http.ResponseWriter, r *http.Request) {
	req, projectSlug, projectContext, ok := h.parseChatRequest(w, r)
	if !ok {
		return
	}

	requestID := chimiddleware.GetReqID(r.Context())

	start := time.Now()
	log.Printf("[CHAT] New query | request_id=%q | project=%q | ip=%s | question=%q",
		requestID, projectSlug, r.RemoteAddr, req.Message)

	answer, err := h.generateAnswer(r.Context(), projectContext, req.Message)
	if err != nil {
		log.Printf("[ERROR] AI generation failed | request_id=%q | project=%q | elapsed=%s | err=%v",
			requestID, projectSlug, time.Since(start).Round(time.Millisecond), err)
		writeError(w, "failed to generate response", http.StatusInternalServerError)
		return
	}

	log.Printf("[CHAT] Response sent | request_id=%q | project=%q | elapsed=%s | answer_length=%d",
		requestID, projectSlug, time.Since(start).Round(time.Millisecond), len(answer))

	writeJSON(w, http.StatusOK, ChatResponse{
		Answer:    answer,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handlers) chatStream(w http.ResponseWriter, r *http.Request) {
	req, projectSlug, projectContext, ok := h.parseChatRequest(w, r)
	if !ok {
		return
	}

	requestID := chimiddleware.GetReqID(r.Context())

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("X-Accel-Buffering", "off")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	flush := func() {
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	emit := func(frame streamFrame) error {
		if err := writeNDJSONFrame(w, frame); err != nil {
			return err
		}
		flush()
		return nil
	}

	start := time.Now()
	log.Printf("[CHAT] Streaming enabled | request_id=%q | project=%q | ip=%s | question=%q",
		requestID, projectSlug, r.RemoteAddr, req.Message)

	err := h.generateAnswerStream(r.Context(), projectContext, req.Message, func(chunk string) error {
		return emit(streamFrame{Type: "token", Text: chunk})
	})
	if err != nil {
		log.Printf("[ERROR] AI streaming failed | request_id=%q | project=%q | elapsed=%s | err=%v",
			requestID, projectSlug, time.Since(start).Round(time.Millisecond), err)
		_ = emit(streamFrame{Type: "error", Message: err.Error()})
		return
	}

	log.Printf("[CHAT] Streaming completed | request_id=%q | project=%q | elapsed=%s",
		requestID, projectSlug, time.Since(start).Round(time.Millisecond))
	_ = emit(streamFrame{Type: "done"})
}

func (h *Handlers) parseChatRequest(w http.ResponseWriter, r *http.Request) (ChatRequest, string, string, bool) {

	// Decode request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return ChatRequest{}, "", "", false
	}
	defer r.Body.Close()

	// Validate and sanitize inputs.
	req.Message = strings.TrimSpace(req.Message)
	req.Project = strings.TrimSpace(req.Project)

	if req.Message == "" {
		writeError(w, "message cannot be empty", http.StatusBadRequest)
		return ChatRequest{}, "", "", false
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

	return req, projectSlug, projectContext, true
}

type streamFrame struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Message string `json:"message,omitempty"`
}

func writeNDJSONFrame(w http.ResponseWriter, frame streamFrame) error {
	line, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	line = append(line, '\n')
	_, err = w.Write(line)
	return err
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
