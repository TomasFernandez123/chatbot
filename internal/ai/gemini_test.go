package ai

import (
	"context"
	"errors"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

type fakeStreamIterator struct {
	responses []*genai.GenerateContentResponse
	errs      []error
	idx       int
}

func (f *fakeStreamIterator) Next() (*genai.GenerateContentResponse, error) {
	if f.idx >= len(f.responses) {
		return nil, iterator.Done
	}

	resp := f.responses[f.idx]
	err := f.errs[f.idx]
	f.idx++
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func textChunkResponse(s string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{Parts: []genai.Part{genai.Text(s)}},
			},
		},
	}
}

func TestGenerateAnswerStream_SuccessInvokesCallbackPerChunk(t *testing.T) {
	iter := &fakeStreamIterator{
		responses: []*genai.GenerateContentResponse{
			textChunkResponse("Hola"),
			textChunkResponse(" mundo"),
			textChunkResponse("!"),
		},
		errs: []error{nil, nil, nil},
	}

	svc := &Service{
		streamFactory: func(ctx context.Context, parts ...genai.Part) streamResponseIterator {
			return iter
		},
	}

	var got []string
	err := svc.GenerateAnswerStream(context.Background(), "ctx", "pregunta", func(chunk string) error {
		got = append(got, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("GenerateAnswerStream() error = %v, want nil", err)
	}

	if len(got) != 3 {
		t.Fatalf("callback calls = %d, want 3", len(got))
	}
	if got[0] != "Hola" || got[1] != " mundo" || got[2] != "!" {
		t.Fatalf("chunks = %#v, want [\"Hola\", \" mundo\", \"!\"]", got)
	}
}

func TestGenerateAnswerStream_IteratorErrorPropagation(t *testing.T) {
	sdkErr := errors.New("sdk stream failed")
	iter := &fakeStreamIterator{
		responses: []*genai.GenerateContentResponse{
			textChunkResponse("primer"),
			nil,
			textChunkResponse("tercero"),
		},
		errs: []error{nil, sdkErr, nil},
	}

	svc := &Service{
		streamFactory: func(ctx context.Context, parts ...genai.Part) streamResponseIterator {
			return iter
		},
	}

	called := 0
	err := svc.GenerateAnswerStream(context.Background(), "ctx", "pregunta", func(chunk string) error {
		called++
		return nil
	})

	if !errors.Is(err, sdkErr) {
		t.Fatalf("GenerateAnswerStream() error = %v, want %v", err, sdkErr)
	}
	if called != 1 {
		t.Fatalf("callback calls after sdk error = %d, want 1", called)
	}
}

func TestGenerateAnswerStream_CallbackErrorPropagation(t *testing.T) {
	cbErr := errors.New("callback failed")
	iter := &fakeStreamIterator{
		responses: []*genai.GenerateContentResponse{
			textChunkResponse("primer"),
			textChunkResponse("segundo"),
		},
		errs: []error{nil, nil},
	}

	svc := &Service{
		streamFactory: func(ctx context.Context, parts ...genai.Part) streamResponseIterator {
			return iter
		},
	}

	called := 0
	err := svc.GenerateAnswerStream(context.Background(), "ctx", "pregunta", func(chunk string) error {
		called++
		if called == 1 {
			return cbErr
		}
		return nil
	})

	if !errors.Is(err, cbErr) {
		t.Fatalf("GenerateAnswerStream() error = %v, want %v", err, cbErr)
	}
	if called != 1 {
		t.Fatalf("callback calls after callback error = %d, want 1", called)
	}
}

func TestGenerateAnswerStream_ContextCancellation(t *testing.T) {
	iter := &fakeStreamIterator{
		responses: []*genai.GenerateContentResponse{textChunkResponse("uno")},
		errs:      []error{nil},
	}

	svc := &Service{
		streamFactory: func(ctx context.Context, parts ...genai.Part) streamResponseIterator {
			return iter
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.GenerateAnswerStream(ctx, "ctx", "pregunta", func(chunk string) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GenerateAnswerStream() error = %v, want %v", err, context.Canceled)
	}
}

func TestGenerateAnswerStream_WhitespaceOnlyChunk_IsEmittedAsToken(t *testing.T) {
	iter := &fakeStreamIterator{
		responses: []*genai.GenerateContentResponse{
			textChunkResponse("   \t"),
			textChunkResponse("hola"),
		},
		errs: []error{nil, nil},
	}

	svc := &Service{
		streamFactory: func(ctx context.Context, parts ...genai.Part) streamResponseIterator {
			return iter
		},
	}

	var got []string
	err := svc.GenerateAnswerStream(context.Background(), "ctx", "pregunta", func(chunk string) error {
		got = append(got, chunk)
		return nil
	})
	if err != nil {
		t.Fatalf("GenerateAnswerStream() error = %v, want nil", err)
	}

	if len(got) != 2 {
		t.Fatalf("callback calls = %d, want 2", len(got))
	}
	if got[0] != "   \t" || got[1] != "hola" {
		t.Fatalf("chunks = %#v, want [\"   \\t\", \"hola\"]", got)
	}
}
