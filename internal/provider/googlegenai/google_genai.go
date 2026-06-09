// Copyright 2026 The Mint Authors.

// Package googlegenai provides Google Gemini LLM client for text translation.
package googlegenai

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
	defaultBaseURL   = "https://generativelanguage.googleapis.com"
	defaultModelName = "gemini-3.1-flash-lite"
)

// Client is a Google Gemini API client.
type Client struct {
	apiKey     string
	baseURL    string
	modelName  string
	httpClient *http.Client
}

// New creates a new Google Gemini client.
func New(apiKey, baseURL, modelName string) *Client {
	if modelName == "" {
		modelName = defaultModelName
	}

	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		modelName:  modelName,
		httpClient: &http.Client{},
	}
}

type requestBody struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	Temperature float64 `json:"temperature"`
}

type responseBody struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content content `json:"content"`
}

// Complete calls the Google Gemini streaming API and writes tokens to w as they arrive.
func (c *Client) Complete(ctx context.Context, prompt string, w io.Writer) error {
	body := requestBody{
		Contents: []content{
			{Parts: []part{{Text: prompt}}},
		},
		GenerationConfig: generationConfig{Temperature: 0.3},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", c.baseURL, c.modelName, c.apiKey)

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
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var result responseBody
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			continue
		}

		if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
			if _, err := fmt.Fprint(w, result.Candidates[0].Content.Parts[0].Text); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	return scanner.Err()
}
