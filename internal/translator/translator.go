// Copyright 2026 The Mint Authors.

// Package translator defines the Translator interface for LLM translation backends.
package translator

import "context"

// Translator translates text into a target language.
type Translator interface {
	Translate(ctx context.Context, text, targetLang string) (string, error)
}
