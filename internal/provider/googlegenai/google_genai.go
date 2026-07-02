// Copyright 2026 The Mint Authors.

// Package googlegenai provides Google Gemini LLM client for text translation.
package googlegenai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/min0625/mint/internal/httpx"
	"github.com/min0625/mint/internal/llm"
)

const (
	defaultBaseURL   = "https://generativelanguage.googleapis.com"
	defaultModelName = "gemini-3.1-flash-lite"
	temperature      = 0.3
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
		httpClient: httpx.New(),
	}
}

type requestBody struct {
	SystemInstruction systemInstruction `json:"systemInstruction"`
	Contents          []content         `json:"contents"`
	GenerationConfig  generationConfig  `json:"generationConfig"`
}

type systemInstruction struct {
	Parts []part `json:"parts"`
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

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type responseBody struct {
	Candidates    []candidate   `json:"candidates"`
	UsageMetadata usageMetadata `json:"usageMetadata"`
	Error         *apiError     `json:"error"`
}

type candidate struct {
	Content      content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}

// Complete calls the Google Gemini streaming API and writes tokens to w as they arrive.
// system is sent as systemInstruction; user is the single turn in contents.
func (c *Client) Complete(ctx context.Context, system, user string, w io.Writer) (llm.Usage, error) {
	body := requestBody{
		SystemInstruction: systemInstruction{Parts: []part{{Text: system}}},
		Contents:          []content{{Parts: []part{{Text: user}}}},
		GenerationConfig:  generationConfig{Temperature: temperature},
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

	var (
		usage     llm.Usage
		truncated bool
	)

	out := llm.NewTrailingNewlineWriter(w)

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

		// Mid-stream error object: the HTTP status was already 200, so this is
		// the only failure signal we get.
		if result.Error != nil {
			return llm.Usage{}, fmt.Errorf("API stream error %d: %s", result.Error.Code, result.Error.Message)
		}

		// usageMetadata accumulates across chunks; the last value is the total.
		if result.UsageMetadata.PromptTokenCount > 0 || result.UsageMetadata.CandidatesTokenCount > 0 {
			usage.InputTokens = result.UsageMetadata.PromptTokenCount
			usage.OutputTokens = result.UsageMetadata.CandidatesTokenCount
		}

		if len(result.Candidates) > 0 {
			if result.Candidates[0].FinishReason == "MAX_TOKENS" {
				truncated = true
			}

			if len(result.Candidates[0].Content.Parts) > 0 {
				if _, err := fmt.Fprint(out, result.Candidates[0].Content.Parts[0].Text); err != nil {
					return llm.Usage{}, err
				}
			}
		}
	}

	if err := out.Done(); err != nil {
		return llm.Usage{}, err
	}

	if err := scanner.Err(); err != nil {
		return usage, err
	}

	// Surface truncation instead of silently returning partial text: without
	// this, a long input would print an incomplete translation and exit 0. No
	// maxOutputTokens is requested, so the limit hit is the model default.
	if truncated {
		return usage, errors.New("output truncated: response hit the model's output-token limit")
	}

	return usage, nil
}
