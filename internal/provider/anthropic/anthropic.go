// Copyright 2026 The Mint Authors.

// Package anthropic provides Anthropic Claude LLM client for text translation.
package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	Stream    bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type streamEvent struct {
	Type  string      `json:"type"`
	Delta streamDelta `json:"delta"`
}

// Translate calls the Anthropic API with streaming and writes tokens to w as they arrive.
func (c *Client) Translate(ctx context.Context, text, targetLang string, w io.Writer) error {
	prompt := fmt.Sprintf(
		"Translate the following text to %s. Output only the translation, nothing else:\n\n%s",
		targetLang, text,
	)

	body := requestBody{
		Model:     c.modelName,
		MaxTokens: 1024,
		Messages:  []message{{Role: "user", Content: prompt}},
		Stream:    true,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var event streamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			if _, err := fmt.Fprint(w, event.Delta.Text); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return scanner.Err()
}
