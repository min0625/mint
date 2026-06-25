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

	"github.com/min0625/mint/internal/llm"
)

const (
	defaultBaseURL   = "https://generativelanguage.googleapis.com"
	defaultModelName = "gemini-3.1-flash-lite"
	// maxScanLineBytes raises bufio.Scanner's default 64KB line limit so a
	// large SSE data line or error body does not abort the stream early.
	maxScanLineBytes = 1 << 20
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
	Candidates    []candidate   `json:"candidates"`
	UsageMetadata usageMetadata `json:"usageMetadata"`
}

type candidate struct {
	Content content `json:"content"`
}

type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}

// Complete calls the Google Gemini streaming API and writes tokens to w as they arrive.
func (c *Client) Complete(ctx context.Context, prompt string, w io.Writer) (llm.Usage, error) {
	body := requestBody{
		Contents: []content{
			{Parts: []part{{Text: prompt}}},
		},
		GenerationConfig: generationConfig{Temperature: 0.3},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return llm.Usage{}, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse", c.baseURL, c.modelName)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return llm.Usage{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Pass the API key via header rather than the URL query string so it does
	// not leak into proxy or server access logs.
	if c.apiKey != "" {
		req.Header.Set("X-Goog-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return llm.Usage{}, fmt.Errorf("call API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return llm.Usage{}, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxScanLineBytes)

	var usage llm.Usage

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

		// usageMetadata accumulates across chunks; the last value is the total.
		if result.UsageMetadata.PromptTokenCount > 0 || result.UsageMetadata.CandidatesTokenCount > 0 {
			usage.InputTokens = result.UsageMetadata.PromptTokenCount
			usage.OutputTokens = result.UsageMetadata.CandidatesTokenCount
		}

		if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
			if _, err := fmt.Fprint(w, result.Candidates[0].Content.Parts[0].Text); err != nil {
				return llm.Usage{}, err
			}
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return llm.Usage{}, err
	}

	return usage, scanner.Err()
}
