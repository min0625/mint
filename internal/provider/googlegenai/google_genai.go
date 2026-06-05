// Copyright 2026 The Mint Authors.

// Package googlegenai provides Google Gemini LLM client for text translation.
package googlegenai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const apiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

// Client is a Google Gemini API client.
type Client struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
}

// New creates a new Google Gemini client.
func New(apiKey, modelName string) *Client {
	if modelName == "" {
		modelName = "gemini-3.1-flash-lite"
	}

	return &Client{
		apiKey:     apiKey,
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

// Translate calls the Google Gemini API to translate text into targetLang (BCP-47 tag).
func (c *Client) Translate(ctx context.Context, text, targetLang string) (string, error) {
	prompt := fmt.Sprintf(
		"Translate the following text to %s. Output only the translation, nothing else:\n\n%s",
		targetLang, text,
	)

	body := requestBody{
		Contents: []content{
			{Parts: []part{{Text: prompt}}},
		},
		GenerationConfig: generationConfig{Temperature: 0.3},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", fmt.Sprintf(apiEndpoint, c.modelName), c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

	if len(result.Candidates) == 0 {
		return "", errors.New("no candidates in response")
	}

	if len(result.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no text in response")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
