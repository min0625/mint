// Copyright 2026 The Mint Authors.

package openai_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/min0625/mint/internal/provider/openai"
)

func TestCompleteStreamsTokens(t *testing.T) {
	const sse = `data: {"choices":[{"delta":{"content":"Hello"}}]}

data: {"choices":[{"delta":{"content":" world"}}]}

data: [DONE]
`

	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder
	if _, err := openai.New("secret-key", srv.URL, "").Complete(t.Context(), "prompt", &sb); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if got, want := sb.String(), "Hello world\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	if got, want := gotAuth, "Bearer secret-key"; got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
}

func TestNewUsesDefaultBaseURL(t *testing.T) {
	c := openai.New("key", "", "custom-model")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCompleteReturnsUsage(t *testing.T) {
	const sse = `data: {"choices":[{"delta":{"content":"Hi"}}]}

data: {"choices":[],"usage":{"prompt_tokens":5,"completion_tokens":2}}

data: [DONE]
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder

	usage, err := openai.New("key", srv.URL, "").Complete(t.Context(), "prompt", &sb)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if usage.InputTokens != 5 {
		t.Errorf("InputTokens = %d, want 5", usage.InputTokens)
	}

	if usage.OutputTokens != 2 {
		t.Errorf("OutputTokens = %d, want 2", usage.OutputTokens)
	}
}

func TestCompleteReturnsErrorOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	var sb strings.Builder

	_, err := openai.New("k", srv.URL, "").Complete(t.Context(), "prompt", &sb)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error %q does not mention status 429", err.Error())
	}
}
