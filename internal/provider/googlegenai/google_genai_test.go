// Copyright 2026 The Mint Authors.

package googlegenai_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/min0625/mint/internal/provider/googlegenai"
)

func TestCompleteStreamsTokens(t *testing.T) {
	const sse = `data: {"candidates":[{"content":{"parts":[{"text":"Hello"}]}}]}

data: {"candidates":[{"content":{"parts":[{"text":" world"}]}}]}
`

	var gotKey, gotQueryKey string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Goog-Api-Key")
		gotQueryKey = r.URL.Query().Get("key")

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder
	if _, err := googlegenai.New("secret-key", srv.URL, "").Complete(t.Context(), "prompt", &sb); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if got, want := sb.String(), "Hello world\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	if gotKey != "secret-key" {
		t.Errorf("X-Goog-Api-Key = %q, want %q", gotKey, "secret-key")
	}

	// The key must not leak into the URL query string.
	if gotQueryKey != "" {
		t.Errorf("api key leaked into query string: %q", gotQueryKey)
	}
}

func TestNewUsesDefaultBaseURL(t *testing.T) {
	c := googlegenai.New("key", "", "custom-model")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCompleteReturnsUsage(t *testing.T) {
	const sse = `data: {"candidates":[{"content":{"parts":[{"text":"Hi"}]}}],"usageMetadata":{"promptTokenCount":8,"candidatesTokenCount":2}}
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(sse))
	}))
	defer srv.Close()

	var sb strings.Builder

	usage, err := googlegenai.New("key", srv.URL, "").Complete(t.Context(), "prompt", &sb)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	if usage.InputTokens != 8 {
		t.Errorf("InputTokens = %d, want 8", usage.InputTokens)
	}

	if usage.OutputTokens != 2 {
		t.Errorf("OutputTokens = %d, want 2", usage.OutputTokens)
	}
}

func TestCompleteReturnsErrorOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"denied"}`))
	}))
	defer srv.Close()

	var sb strings.Builder

	_, err := googlegenai.New("k", srv.URL, "").Complete(t.Context(), "prompt", &sb)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error %q does not mention status 403", err.Error())
	}
}
