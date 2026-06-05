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
	ProviderGoogle    = "google"
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderOllama    = "ollama"
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

	// Validate provider
	validProviders := map[string]bool{
		ProviderGoogle:    true,
		ProviderOpenAI:    true,
		ProviderAnthropic: true,
		ProviderOllama:    true,
	}
	if !validProviders[c.Provider] {
		return fmt.Errorf(
			"unsupported provider: %s. Supported: %s, %s, %s, %s",
			c.Provider,
			ProviderGoogle,
			ProviderOpenAI,
			ProviderAnthropic,
			ProviderOllama,
		)
	}

	// Check API key for non-local providers
	if c.Provider != ProviderOllama && c.APIKey == "" {
		return fmt.Errorf("MINT_API_KEY is required for provider: %s", c.Provider)
	}

	// Ollama requires BaseURL
	if c.Provider == ProviderOllama && c.BaseURL == "" {
		return errors.New("MINT_BASE_URL is required for ollama provider (e.g., http://localhost:11434)")
	}

	// Ollama requires ModelName
	if c.Provider == ProviderOllama && c.ModelName == "" {
		return errors.New("MINT_MODEL_NAME is required for ollama provider")
	}

	return nil
}
