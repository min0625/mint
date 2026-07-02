// Copyright 2026 The Mint Authors.

// Package openai provides OpenAI GPT LLM client for text translation.
package openai

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
	defaultBaseURL   = "https://api.openai.com"
	defaultAPIPath   = "/v1/chat/completions"
	defaultModelName = "gpt-4o-mini"
	temperature      = 0.3
	// maxScanLineBytes raises bufio.Scanner's default 64KB line limit so a
	// large SSE data line or error body does not abort the stream early.
	maxScanLineBytes = 1 << 20
)

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
	Model         string         `json:"model"`
	Messages      []message      `json:"messages"`
	Temperature   float64        `json:"temperature"`
	Stream        bool           `json:"stream"`
	StreamOptions *streamOptions `json:"stream_options,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamDelta struct {
	Content string `json:"content"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type streamUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type streamChoice struct {
	Delta        streamDelta `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

type streamError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type streamResponse struct {
	Choices []streamChoice `json:"choices"`
	Usage   *streamUsage   `json:"usage"`
	Error   *streamError   `json:"error"`
}

// Complete calls the OpenAI API with streaming and writes tokens to w as they arrive.
// system is sent as a system-role message followed by user as a user-role message.
//
// stream_options.include_usage is always requested so token counts are
// available in verbose mode. It is an OpenAI extension, but every endpoint we
// document (OpenAI, Ollama, LM Studio) either supports it or ignores unknown
// fields. A strict OpenAI-compatible server that rejects it would surface a
// plain API error; graceful degradation can be added if such a case is ever
// reported.
func (c *Client) Complete(ctx context.Context, system, user string, w io.Writer) (llm.Usage, error) {
	body := requestBody{
		Model: c.modelName,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature:   temperature,
		Stream:        true,
		StreamOptions: &streamOptions{IncludeUsage: true},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return llm.Usage{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+defaultAPIPath, bytes.NewReader(jsonBody))
	if err != nil {
		return llm.Usage{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		if data == "[DONE]" {
			break
		}

		var sr streamResponse
		if err := json.Unmarshal([]byte(data), &sr); err != nil {
			continue
		}

		// Mid-stream error object: the HTTP status was already 200, so this is
		// the only failure signal we get (some proxies and local servers emit
		// errors this way).
		if sr.Error != nil {
			return llm.Usage{}, fmt.Errorf("API stream error: %s", sr.Error.Message)
		}

		if sr.Usage != nil {
			usage.InputTokens = sr.Usage.PromptTokens
			usage.OutputTokens = sr.Usage.CompletionTokens
		}

		if len(sr.Choices) > 0 {
			if sr.Choices[0].FinishReason == "length" {
				truncated = true
			}

			if _, err := fmt.Fprint(out, sr.Choices[0].Delta.Content); err != nil {
				return llm.Usage{}, err
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
	// max_tokens is requested, so the limit hit is the server or model default.
	if truncated {
		return usage, errors.New("output truncated: response hit the model's output-token limit")
	}

	return usage, nil
}
