// Copyright 2026 The Mint Authors.

// Package ollama provides Ollama local LLM client for text translation.
package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is an Ollama API client.
type Client struct {
	baseURL    string
	modelName  string
	httpClient *http.Client
}

// New creates a new Ollama client.
func New(baseURL, modelName string) *Client {
	return &Client{
		baseURL:    baseURL,
		modelName:  modelName,
		httpClient: &http.Client{},
	}
}

type requestBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type streamResponseBody struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Translate calls the Ollama API with streaming and writes tokens to w as they arrive.
func (c *Client) Translate(ctx context.Context, text, targetLang string, w io.Writer) error {
	prompt := fmt.Sprintf(
		"Translate the following text to %s. Output only the translation, nothing else:\n\n%s",
		targetLang, text,
	)

	body := requestBody{
		Model:  c.modelName,
		Prompt: prompt,
		Stream: true,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/api/generate"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		var sr streamResponseBody
		if err := json.Unmarshal(scanner.Bytes(), &sr); err != nil {
			continue
		}

		if _, err := fmt.Fprint(w, sr.Response); err != nil {
			return err
		}

		if sr.Done {
			break
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return scanner.Err()
}
