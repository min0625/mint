// Copyright 2026 The Mint Authors.

package anthropic_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/min0625/mint/internal/provider/anthropic"
)

func TestCompleteStreamsTokens(t *testing.T) {
	const sse = `event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":" world"}}

data: {"type":"message_stop"}
`

	var gotKey, gotVersion string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Api-Key")
		gotVersion = r.Header.Get("Anthropic-Version")
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder
	if _, err := anthropic.New("secret-key", srv.URL, "").
		Complete(t.Context(), "system text", "user text", &sb); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if got, want := sb.String(), "Hello world\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	if gotKey != "secret-key" {
		t.Errorf("X-Api-Key = %q, want %q", gotKey, "secret-key")
	}

	if gotVersion == "" {
		t.Error("Anthropic-Version header was not set")
	}
}

func TestNewUsesDefaultBaseURL(t *testing.T) {
	c := anthropic.New("key", "", "custom-model")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCompleteReturnsUsage(t *testing.T) {
	const sse = `data: {"type":"message_start","message":{"usage":{"input_tokens":10,"output_tokens":0}}}

data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hi"}}

data: {"type":"message_delta","usage":{"output_tokens":3}}

data: {"type":"message_stop"}
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder

	usage, err := anthropic.New("key", srv.URL, "").Complete(t.Context(), "system text", "user text", &sb)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if usage.InputTokens != 10 {
		t.Errorf("InputTokens = %d, want 10", usage.InputTokens)
	}

	if usage.OutputTokens != 3 {
		t.Errorf("OutputTokens = %d, want 3", usage.OutputTokens)
	}
}

func TestCompleteReturnsErrorOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid key"}`))
	}))
	defer srv.Close()

	var sb strings.Builder

	_, err := anthropic.New("k", srv.URL, "").Complete(t.Context(), "system text", "user text", &sb)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error %q does not mention status 401", err.Error())
	}
}

// A malformed data line (e.g. truncated by a flaky proxy) must not abort the
// whole stream; it is skipped and the well-formed chunks around it still
// reach the caller.
func TestCompleteSkipsMalformedDataLine(t *testing.T) {
	const sse = `event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}

data: {not valid json}

event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":" world"}}

data: {"type":"message_stop"}
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder
	if _, err := anthropic.New("k", srv.URL, "").Complete(t.Context(), "sys", "usr", &sb); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if got, want := sb.String(), "Hello world\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestCompleteRoleSeparation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var req struct {
			System   string `json:"system"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}

		_ = json.Unmarshal(body, &req)

		if req.System != "my system instruction" {
			t.Errorf("system field = %q, want %q", req.System, "my system instruction")
		}

		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Errorf("expected one user message, got %+v", req.Messages)
		}

		if req.Messages[0].Content != "my user text" {
			t.Errorf("user content = %q, want %q", req.Messages[0].Content, "my user text")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"type\":\"message_stop\"}\n"))
	}))
	defer srv.Close()

	var sb strings.Builder

	_, _ = anthropic.New("k", srv.URL, "").Complete(t.Context(), "my system instruction", "my user text", &sb)
}

// The API reports output truncation only via message_delta's stop_reason;
// Complete must surface it as an error instead of returning partial text
// with a nil error.
func TestCompleteReturnsErrorOnMaxTokensTruncation(t *testing.T) {
	const sse = `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"partial"}}

data: {"type":"message_delta","delta":{"stop_reason":"max_tokens"},"usage":{"output_tokens":8192}}

data: {"type":"message_stop"}
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder

	usage, err := anthropic.New("k", srv.URL, "").Complete(t.Context(), "sys", "usr", &sb)
	if err == nil {
		t.Fatal("expected truncation error, got nil")
	}

	if !strings.Contains(err.Error(), "truncated") {
		t.Errorf("error %q does not mention truncation", err.Error())
	}

	// The partial text was already streamed to the caller and usage was
	// collected; both must still be delivered alongside the error.
	if got, want := sb.String(), "partial\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	if usage.OutputTokens != 8192 {
		t.Errorf("OutputTokens = %d, want 8192", usage.OutputTokens)
	}
}

// A mid-stream error event arrives after the HTTP status was already 200, so
// it is the only failure signal; Complete must return it as an error.
func TestCompleteReturnsErrorOnStreamErrorEvent(t *testing.T) {
	const sse = `data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hi"}}

data: {"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder

	_, err := anthropic.New("k", srv.URL, "").Complete(t.Context(), "sys", "usr", &sb)
	if err == nil {
		t.Fatal("expected stream error, got nil")
	}

	if !strings.Contains(err.Error(), "Overloaded") || !strings.Contains(err.Error(), "overloaded_error") {
		t.Errorf("error %q does not carry the stream error details", err.Error())
	}
}
