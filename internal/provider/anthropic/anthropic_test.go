// Copyright 2026 The Mint Authors.

package anthropic_test

import (
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
	if _, err := anthropic.New("secret-key", srv.URL, "").Complete(t.Context(), "prompt", &sb); err != nil {
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

	usage, err := anthropic.New("key", srv.URL, "").Complete(t.Context(), "prompt", &sb)
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

	_, err := anthropic.New("k", srv.URL, "").Complete(t.Context(), "prompt", &sb)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error %q does not mention status 401", err.Error())
	}
}
