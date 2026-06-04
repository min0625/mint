// Copyright 2026 The Mint Authors.

// Package anthropic provides Anthropic Claude LLM client for text translation.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultAPIEndpoint = "https://api.anthropic.com/v1/messages"
	anthropicVersion   = "2023-06-01"
)

// Client is an Anthropic Claude API client.
type Client struct {
	apiKey     string
	baseURL    string
	modelName  string
	httpClient *http.Client
}

// New creates a new Anthropic client.
func New(apiKey, baseURL, modelName string) *Client {
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}

	if baseURL == "" {
		baseURL = defaultAPIEndpoint
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		modelName:  modelName,
		httpClient: &http.Client{},
	}
}

type requestBody struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseBody struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Translate calls the Anthropic API to translate text into targetLang (BCP-47 tag).
func (c *Client) Translate(ctx context.Context, text, targetLang string) (string, error) {
	prompt := fmt.Sprintf(
		"Translate the following text to %s. Output only the translation, nothing else:\n\n%s",
		targetLang, text,
	)

	body := requestBody{
		Model:     c.modelName,
		MaxTokens: 1024,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	var result responseBody
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", errors.New("no content in response")
	}

	return result.Content[0].Text, nil
}
