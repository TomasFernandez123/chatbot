package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type streamResponseIterator interface {
	Next() (*genai.GenerateContentResponse, error)
}

type sdkStreamResponseIterator struct {
	iter *genai.GenerateContentResponseIterator
}

func (s *sdkStreamResponseIterator) Next() (*genai.GenerateContentResponse, error) {
	return s.iter.Next()
}

// Service encapsulates the Gemini AI client and generation logic.
type Service struct {
	client        *genai.Client
	model         *genai.GenerativeModel
	streamFactory func(ctx context.Context, parts ...genai.Part) streamResponseIterator
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
		streamFactory: func(ctx context.Context, parts ...genai.Part) streamResponseIterator {
			return &sdkStreamResponseIterator{iter: model.GenerateContentStream(ctx, parts...)}
		},
	}, nil
}

// Close releases resources held by the AI client.
func (s *Service) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// GenerateAnswer takes a project context document and a user question,
// then returns a grounded answer from Gemini acting as Tomas-AI.
func (s *Service) GenerateAnswer(ctx context.Context, projectContext, question string) (string, error) {
	s.model.SystemInstruction = genai.NewUserContent(genai.Text(systemPrompt))
	prompt := buildPrompt(projectContext, question)

	log.Printf("[AI] Generating answer | question=%q", question)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return extractText(resp), nil
}

// GenerateAnswerStream streams Gemini response chunks and invokes onChunk for each chunk.
// Returns nil on successful completion, SDK iterator errors, callback errors,
// or context cancellation errors.
func (s *Service) GenerateAnswerStream(
	ctx context.Context,
	projectContext, question string,
	onChunk func(string) error,
) error {
	if onChunk == nil {
		return errors.New("onChunk callback is required")
	}

	if s.model != nil {
		s.model.SystemInstruction = genai.NewUserContent(genai.Text(systemPrompt))
	}

	stream := s.newStreamIterator(ctx, genai.Text(buildPrompt(projectContext, question)))
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		resp, err := stream.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return err
		}

		chunk := extractStreamText(resp)
		// Contract: whitespace-only chunks are emitted as-is; only truly empty chunks are skipped.
		if chunk == "" {
			continue
		}

		if err := onChunk(chunk); err != nil {
			return err
		}
	}
}

func (s *Service) newStreamIterator(ctx context.Context, parts ...genai.Part) streamResponseIterator {
	if s.streamFactory != nil {
		return s.streamFactory(ctx, parts...)
	}

	return &sdkStreamResponseIterator{iter: s.model.GenerateContentStream(ctx, parts...)}
}

func extractStreamText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0] == nil || resp.Candidates[0].Content == nil {
		return ""
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			sb.WriteString(string(txt))
		}
	}

	return sb.String()
}

const systemPrompt = `Sos **Tomas-AI**, el asistente técnico virtual de Tomas Fernandez, desarrollador de software fullstack.
Tu único objetivo es responder preguntas sobre los proyectos de su portfolio basándote EXCLUSIVAMENTE en la documentación inyectada como "Fuente de Verdad".

## Identidad
- Representás profesionalmente a Tomas Fernandez.
- Tu tono es técnico, conciso y amigable. Evitá respuestas vagas o geniales sin sustento.

## Reglas estrictas
1. Respondé SOLO con información presente en el contexto proporcionado (Única Fuente de Verdad).
2. Si la pregunta no puede responderse con ese contexto, respondé exactamente:
   "No tengo esa información en la documentación de este proyecto. Te recomiendo agendar una entrevista técnica con Tomas para profundizar en el tema."
3. No inventes funcionalidades, tecnologías ni métricas que no estén documentadas.
4. Respondé en el mismo idioma en que te escriben.
5. Usá Markdown para formatear listas, bloques de código o secciones cuando mejore la claridad.`

func buildPrompt(projectContext, question string) string {
	return fmt.Sprintf(
		"## Única Fuente de Verdad (documentación del proyecto)\n\n%s\n\n---\n\n## Pregunta\n\n%s",
		projectContext, question,
	)
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
