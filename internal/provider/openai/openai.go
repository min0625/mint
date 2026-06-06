// Copyright 2026 The Mint Authors.

// Package openai provides OpenAI GPT LLM client for text translation.
package openai

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

const defaultAPIEndpoint = "https://api.openai.com/v1/chat/completions"

// Client is an OpenAI API client.
type Client struct {
	apiKey     string
	baseURL    string
	modelName  string
	httpClient *http.Client
}

// New creates a new OpenAI client.
func New(apiKey, baseURL, modelName string) *Client {
	if modelName == "" {
		modelName = "gpt-4o-mini"
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
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamDelta struct {
	Content string `json:"content"`
}

type streamChoice struct {
	Delta streamDelta `json:"delta"`
}

type streamResponse struct {
	Choices []streamChoice `json:"choices"`
}

// Translate calls the OpenAI API with streaming and writes tokens to w as they arrive.
func (c *Client) Translate(ctx context.Context, text, targetLang string, w io.Writer) error {
	prompt := fmt.Sprintf(
		"Translate the following text to %s. Output only the translation, nothing else:\n\n%s",
		targetLang, text,
	)

	body := requestBody{
		Model:       c.modelName,
		Messages:    []message{{Role: "user", Content: prompt}},
		Temperature: 0.3,
		Stream:      true,
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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		if data == "[DONE]" {
			break
		}

		var sr streamResponse
		if err := json.Unmarshal([]byte(data), &sr); err != nil {
			continue
		}

		if len(sr.Choices) > 0 {
			if _, err := fmt.Fprint(w, sr.Choices[0].Delta.Content); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return scanner.Err()
}
