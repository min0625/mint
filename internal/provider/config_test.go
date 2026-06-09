// Copyright 2026 The Mint Authors.
package provider_test

import (
	"strings"
	"testing"

	"github.com/min0625/mint/internal/provider"
)

const errUnsupportedProvider = "unsupported provider"

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     provider.Config
		wantErr string
	}{
		{"missing provider", provider.Config{}, "MINT_PROVIDER"},
		{"unsupported provider", provider.Config{Provider: "badprovider", APIKey: "k"}, errUnsupportedProvider},
		{"ollama no longer a valid provider", provider.Config{Provider: "ollama", APIKey: "k"}, errUnsupportedProvider},
		{"missing api key no base url", provider.Config{Provider: provider.ProviderOpenAI}, "MINT_API_KEY"},
		{"valid openai with key", provider.Config{Provider: provider.ProviderOpenAI, APIKey: "k"}, ""},
		{
			"valid openai with base url no key (proxy)",
			provider.Config{Provider: provider.ProviderOpenAI, BaseURL: "http://localhost:11434"},
			"",
		},
		{"valid google-genai with key", provider.Config{Provider: provider.ProviderGoogleGenAI, APIKey: "k"}, ""},
		{"valid anthropic with key", provider.Config{Provider: provider.ProviderAnthropic, APIKey: "k"}, ""},
		{"provider normalized to lowercase", provider.Config{Provider: "OpenAI", APIKey: "k"}, ""},
		{"provider all-caps normalized", provider.Config{Provider: "ANTHROPIC", APIKey: "k"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateConfig()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}
