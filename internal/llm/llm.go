// Copyright 2026 The Mint Authors.

// Package llm defines the Completer interface for LLM backends.
package llm

import (
	"context"
	"io"
)

// Usage holds token consumption for a single LLM call.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// Completer sends a prompt to an LLM and streams the response.
// Implementations write tokens directly to w as they arrive.
type Completer interface {
	Complete(ctx context.Context, prompt string, w io.Writer) (Usage, error)
}
