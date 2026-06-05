// Copyright 2026 The Mint Authors.

package provider

import (
	"context"
	"fmt"

	"github.com/min0625/mint/internal/provider/anthropic"
	"github.com/min0625/mint/internal/provider/googlegenai"
	"github.com/min0625/mint/internal/provider/ollama"
	"github.com/min0625/mint/internal/provider/openai"
	"github.com/min0625/mint/internal/translator"
)

// NewTranslator creates a new Translator based on the provider configuration.
func NewTranslator(_ context.Context, cfg Config) (translator.Translator, error) {
	if err := cfg.ValidateConfig(); err != nil {
		return nil, err
	}

	switch cfg.Provider {
	case ProviderGoogleGenAI:
		return googlegenai.New(cfg.APIKey, cfg.ModelName), nil
	case ProviderOpenAI:
		return openai.New(cfg.APIKey, cfg.BaseURL, cfg.ModelName), nil
	case ProviderAnthropic:
		return anthropic.New(cfg.APIKey, cfg.BaseURL, cfg.ModelName), nil
	case ProviderOllama:
		return ollama.New(cfg.BaseURL, cfg.ModelName), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}
