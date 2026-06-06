// Copyright 2026 The Mint Authors.

// Package translator defines the Translator interface for LLM translation backends.
package translator

import (
	"context"
	"io"
)

// Translator translates text into a target language.
// Implementations write tokens directly to w as they arrive.
type Translator interface {
	Translate(ctx context.Context, text, targetLang string, w io.Writer) error
}
