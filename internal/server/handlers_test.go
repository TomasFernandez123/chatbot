package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChat_WithNDJSONAccept_StreamsTokenAndDoneFrames(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswerStream = func(ctx context.Context, projectContext, question string, onChunk func(string) error) error {
		if err := onChunk("hola"); err != nil {
			return err
		}
		if err := onChunk(" mundo"); err != nil {
			return err
		}
		return nil
	}

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	req.Header.Set("Accept", "application/x-ndjson")
	rec := httptest.NewRecorder()

	h.Chat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/x-ndjson" {
		t.Fatalf("content-type = %q, want %q", got, "application/x-ndjson")
	}
	if got := rec.Header().Get("X-Accel-Buffering"); got != "off" {
		t.Fatalf("X-Accel-Buffering = %q, want %q", got, "off")
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("frames = %d, want 3", len(lines))
	}

	var f1, f2, f3 streamFrame
	if err := json.Unmarshal([]byte(lines[0]), &f1); err != nil {
		t.Fatalf("frame 1 unmarshal error: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &f2); err != nil {
		t.Fatalf("frame 2 unmarshal error: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[2]), &f3); err != nil {
		t.Fatalf("frame 3 unmarshal error: %v", err)
	}

	if f1.Type != "token" || f1.Text != "hola" {
		t.Fatalf("frame 1 = %#v, want token/hola", f1)
	}
	if f2.Type != "token" || f2.Text != " mundo" {
		t.Fatalf("frame 2 = %#v, want token/' mundo'", f2)
	}
	if f3.Type != "done" {
		t.Fatalf("frame 3 = %#v, want done", f3)
	}
}

func TestChat_WithNDJSONAccept_OnStreamError_EmitsErrorFrame(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswerStream = func(ctx context.Context, projectContext, question string, onChunk func(string) error) error {
		_ = onChunk("partial")
		return errors.New("stream failed")
	}

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	req.Header.Set("Accept", "application/x-ndjson")
	rec := httptest.NewRecorder()

	h.Chat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("frames = %d, want 2", len(lines))
	}

	var tokenFrame, errorFrame streamFrame
	if err := json.Unmarshal([]byte(lines[0]), &tokenFrame); err != nil {
		t.Fatalf("token frame unmarshal error: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &errorFrame); err != nil {
		t.Fatalf("error frame unmarshal error: %v", err)
	}

	if tokenFrame.Type != "token" {
		t.Fatalf("first frame = %#v, want token", tokenFrame)
	}
	if errorFrame.Type != "error" || errorFrame.Message != "stream failed" {
		t.Fatalf("second frame = %#v, want error/stream failed", errorFrame)
	}
}

func TestChat_WithoutStreamingHeader_ReturnsClassicJSON(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswer = func(ctx context.Context, projectContext, question string) (string, error) {
		return "respuesta tradicional", nil
	}

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	rec := httptest.NewRecorder()

	h.Chat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want %q", got, "application/json")
	}

	var resp ChatResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response unmarshal error: %v", err)
	}
	if resp.Answer != "respuesta tradicional" {
		t.Fatalf("answer = %q, want %q", resp.Answer, "respuesta tradicional")
	}
}

func TestChat_EndToEnd_WithoutStreamingHeader_RemainsBufferedJSON(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswer = func(ctx context.Context, projectContext, question string) (string, error) {
		if _, hasDeadline := ctx.Deadline(); hasDeadline {
			return "", errors.New("chat endpoint should not have middleware timeout deadline")
		}
		return "respuesta e2e", nil
	}

	ts := httptest.NewServer(NewRouter(h, []string{"http://localhost:4200"}))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	if err != nil {
		t.Fatalf("new request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("http do error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want %q", got, "application/json")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body error: %v", err)
	}

	var out ChatResponse
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("response unmarshal error: %v", err)
	}
	if out.Answer != "respuesta e2e" {
		t.Fatalf("answer = %q, want %q", out.Answer, "respuesta e2e")
	}
}

func TestChat_EndToEnd_NDJSON_FirstTokenArrivesBeforeCompletionAndEndsWithDone(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswerStream = func(ctx context.Context, projectContext, question string, onChunk func(string) error) error {
		if err := onChunk("primer-token"); err != nil {
			return err
		}
		time.Sleep(150 * time.Millisecond)
		if err := onChunk("segundo-token"); err != nil {
			return err
		}
		return nil
	}

	ts := httptest.NewServer(NewRouter(h, []string{"http://localhost:4200"}))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	if err != nil {
		t.Fatalf("new request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	start := time.Now()
	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("http do error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/x-ndjson" {
		t.Fatalf("content-type = %q, want %q", got, "application/x-ndjson")
	}

	r := bufio.NewReader(resp.Body)
	firstLine, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("read first frame error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("first frame arrived after %s, want <= 1s", elapsed)
	}

	var first streamFrame
	if err := json.Unmarshal([]byte(strings.TrimSpace(firstLine)), &first); err != nil {
		t.Fatalf("unmarshal first frame error: %v", err)
	}
	if first.Type != "token" || first.Text != "primer-token" {
		t.Fatalf("first frame = %#v, want token/primer-token", first)
	}

	secondLine, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("read second frame error: %v", err)
	}
	var second streamFrame
	if err := json.Unmarshal([]byte(strings.TrimSpace(secondLine)), &second); err != nil {
		t.Fatalf("unmarshal second frame error: %v", err)
	}
	if second.Type != "token" || second.Text != "segundo-token" {
		t.Fatalf("second frame = %#v, want token/segundo-token", second)
	}

	doneLine, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("read done frame error: %v", err)
	}
	var done streamFrame
	if err := json.Unmarshal([]byte(strings.TrimSpace(doneLine)), &done); err != nil {
		t.Fatalf("unmarshal done frame error: %v", err)
	}
	if done.Type != "done" {
		t.Fatalf("done frame = %#v, want done", done)
	}
}

func TestChat_WithNDJSONAccept_ServiceUnavailable_EmitsErrorFrame(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	req.Header.Set("Accept", "application/x-ndjson")
	rec := httptest.NewRecorder()

	h.Chat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("frames = %d, want 1", len(lines))
	}

	var errorFrame streamFrame
	if err := json.Unmarshal([]byte(lines[0]), &errorFrame); err != nil {
		t.Fatalf("error frame unmarshal error: %v", err)
	}
	if errorFrame.Type != "error" || errorFrame.Message != "service unavailable" {
		t.Fatalf("frame = %#v, want error/service unavailable", errorFrame)
	}
}

func TestChat_WithNDJSONAccept_ContextCancelled_EmitsErrorFrame(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswerStream = func(ctx context.Context, projectContext, question string, onChunk func(string) error) error {
		return context.Canceled
	}

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	req.Header.Set("Accept", "application/x-ndjson")
	rec := httptest.NewRecorder()

	h.Chat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	lines := strings.Split(strings.TrimSpace(rec.Body.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("frames = %d, want 1", len(lines))
	}

	var frame streamFrame
	if err := json.Unmarshal([]byte(lines[0]), &frame); err != nil {
		t.Fatalf("frame unmarshal error: %v", err)
	}
	if frame.Type != "error" || frame.Message != context.Canceled.Error() {
		t.Fatalf("frame = %#v, want error/%q", frame, context.Canceled.Error())
	}
}

func TestChat_EndToEnd_StallBeyondWriteTimeout_ConnectionClosedWithoutDone(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswerStream = func(ctx context.Context, projectContext, question string, onChunk func(string) error) error {
		if err := onChunk("primer-token"); err != nil {
			return err
		}

		// Simula stall de generación mayor al WriteTimeout del server.
		time.Sleep(250 * time.Millisecond)
		return onChunk("token-tarde")
	}

	ts := httptest.NewUnstartedServer(NewRouter(h, []string{"http://localhost:4200"}))
	ts.Config.WriteTimeout = 200 * time.Millisecond
	ts.Start()
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	if err != nil {
		t.Fatalf("new request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("http do error: %v", err)
	}
	defer resp.Body.Close()

	r := bufio.NewReader(resp.Body)
	firstLine, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("read first frame error: %v", err)
	}

	var first streamFrame
	if err := json.Unmarshal([]byte(strings.TrimSpace(firstLine)), &first); err != nil {
		t.Fatalf("unmarshal first frame error: %v", err)
	}
	if first.Type != "token" || first.Text != "primer-token" {
		t.Fatalf("first frame = %#v, want token/primer-token", first)
	}

	rest, err := io.ReadAll(r)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("read remaining stream error: %v", err)
	}
	stream := firstLine + string(rest)

	if strings.Contains(stream, `"type":"done"`) {
		t.Fatalf("unexpected done frame after write-timeout stall: %q", stream)
	}
	if strings.Contains(stream, "token-tarde") {
		t.Fatalf("unexpected late token after write-timeout stall: %q", stream)
	}
}

func TestChat_EndToEnd_ProxyTimeoutBeforeServerTimeout_ClosesStreamEarly(t *testing.T) {
	h := NewHandlers(nil, t.TempDir())
	h.generateAnswerStream = func(ctx context.Context, projectContext, question string, onChunk func(string) error) error {
		if err := onChunk("token-inicial"); err != nil {
			return err
		}

		// Simula pausa de backend mayor al timeout del proxy pero menor al del server.
		time.Sleep(700 * time.Millisecond)
		return onChunk("token-final")
	}

	backend := httptest.NewUnstartedServer(NewRouter(h, []string{"http://localhost:4200"}))
	backend.Config.WriteTimeout = 2 * time.Second
	backend.Start()
	defer backend.Close()

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
		defer cancel()

		upReq, err := http.NewRequestWithContext(ctx, r.Method, backend.URL+r.URL.Path, bytes.NewReader([]byte(`{"message":"pregunta","project":"demo"}`)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		upReq.Header.Set("Content-Type", "application/json")
		upReq.Header.Set("Accept", "application/x-ndjson")

		upResp, err := http.DefaultClient.Do(upReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusGatewayTimeout)
			return
		}
		defer upResp.Body.Close()

		for k, values := range upResp.Header {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(upResp.StatusCode)

		reader := bufio.NewReader(upResp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				if _, writeErr := w.Write(line); writeErr != nil {
					return
				}
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			}
			if err != nil {
				return
			}
		}
	}))
	defer proxy.Close()

	req, err := http.NewRequest(http.MethodPost, proxy.URL+"/chat", strings.NewReader(`{"message":"pregunta","project":"demo"}`))
	if err != nil {
		t.Fatalf("new request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	start := time.Now()
	resp, err := proxy.Client().Do(req)
	if err != nil {
		t.Fatalf("proxy request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read proxy body error: %v", err)
	}
	elapsed := time.Since(start)

	if !strings.Contains(string(body), "token-inicial") {
		t.Fatalf("expected initial token frame, got: %q", string(body))
	}
	if strings.Contains(string(body), "token-final") {
		t.Fatalf("unexpected final token; proxy should cut stream early: %q", string(body))
	}
	if strings.Contains(string(body), `"type":"done"`) {
		t.Fatalf("unexpected done frame; proxy should cut stream early: %q", string(body))
	}
	if elapsed >= 2*time.Second {
		t.Fatalf("proxy did not cut before backend timeout, elapsed=%s", elapsed)
	}
}
