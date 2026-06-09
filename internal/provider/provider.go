// Copyright 2026 The Mint Authors.

package provider

import (
	"fmt"

	"github.com/min0625/mint/internal/llm"
	"github.com/min0625/mint/internal/provider/anthropic"
	"github.com/min0625/mint/internal/provider/googlegenai"
	"github.com/min0625/mint/internal/provider/openai"
)

// NewCompleter creates a new Completer based on the provider configuration.
func NewCompleter(cfg Config) (llm.Completer, error) {
	if err := cfg.ValidateConfig(); err != nil {
		return nil, err
	}

	switch cfg.Provider {
	case ProviderGoogleGenAI:
		return googlegenai.New(cfg.APIKey, cfg.BaseURL, cfg.ModelName), nil
	case ProviderOpenAI:
		return openai.New(cfg.APIKey, cfg.BaseURL, cfg.ModelName), nil
	case ProviderAnthropic:
		return anthropic.New(cfg.APIKey, cfg.BaseURL, cfg.ModelName), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}
