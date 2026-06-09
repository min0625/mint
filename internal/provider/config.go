// Copyright 2026 The Mint Authors.

// Package provider manages LLM provider configuration and initialization.
package provider

import (
	"errors"
	"fmt"
	"strings"
)

// Provider constants.
const (
	ProviderGoogleGenAI = "google-genai"
	ProviderOpenAI      = "openai"
	ProviderAnthropic   = "anthropic"
)

// Config holds provider configuration loaded from environment variables.
type Config struct {
	Provider   string
	APIKey     string
	BaseURL    string
	ModelName  string
	TargetLang string // Comma-separated target languages (e.g., "en", "en,zh-TW")
}

// ValidateConfig validates the provider configuration.
func (c *Config) ValidateConfig() error {
	if c.Provider == "" {
		return errors.New("MINT_PROVIDER environment variable is required")
	}

	c.Provider = strings.ToLower(c.Provider)

	switch c.Provider {
	case ProviderGoogleGenAI, ProviderOpenAI, ProviderAnthropic:
		// valid
	default:
		return fmt.Errorf(
			"unsupported provider: %s. Supported: %s, %s, %s",
			c.Provider,
			ProviderGoogleGenAI,
			ProviderOpenAI,
			ProviderAnthropic,
		)
	}

	if c.BaseURL == "" && c.APIKey == "" {
		return fmt.Errorf("MINT_API_KEY is required for provider: %s", c.Provider)
	}

	return nil
}
