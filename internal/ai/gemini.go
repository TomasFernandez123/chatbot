package ai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Service encapsulates the Gemini AI client and generation logic.
type Service struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewService creates a new AI service configured with the given API key.
func NewService(ctx context.Context, apiKey string) (*Service, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash")

	// Low temperature for factual, grounded responses about real projects.
	temp := float32(0.3)
	model.Temperature = &temp

	topP := float32(0.9)
	model.TopP = &topP

	maxTokens := int32(1024)
	model.MaxOutputTokens = &maxTokens

	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockOnlyHigh},
	}

	return &Service{
		client: client,
		model:  model,
	}, nil
}

// Close releases resources held by the AI client.
func (s *Service) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// GenerateAnswer takes a project README (context) and a user question,
// then returns a grounded answer from Gemini.
func (s *Service) GenerateAnswer(ctx context.Context, projectContext, question string) (string, error) {
	systemPrompt := `Sos un asistente técnico virtual de Tomas Fernandez, un desarrollador de software.
Tu único objetivo es responder preguntas sobre los proyectos del portfolio de Tomas basándote EXCLUSIVAMENTE en el contexto que se te proporciona.

Reglas estrictas:
1. SOLO respondé con información que esté presente en el contexto proporcionado.
2. Si la pregunta no puede ser respondida con el contexto dado, decí: "No tengo información sobre eso en este proyecto. Te recomiendo contactar a Tomas directamente."
3. Sé conciso, profesional y técnico en tus respuestas.
4. Respondé en el mismo idioma en el que te preguntan.
5. No inventes funcionalidades, tecnologías ni detalles que no estén en el contexto.
6. Podés formatear las respuestas con Markdown cuando sea útil.`

	s.model.SystemInstruction = genai.NewUserContent(genai.Text(systemPrompt))

	prompt := fmt.Sprintf("## Contexto del Proyecto\n\n%s\n\n## Pregunta del Usuario\n\n%s", projectContext, question)

	log.Printf("[AI] Generating answer | question=%q | context_length=%d", question, len(projectContext))

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return extractText(resp), nil
}

// extractText pulls the text content from the Gemini response.
func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return "No pude generar una respuesta en este momento."
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			sb.WriteString(string(txt))
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "No pude generar una respuesta en este momento."
	}
	return result
}
