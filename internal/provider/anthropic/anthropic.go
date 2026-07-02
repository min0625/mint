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

	"github.com/min0625/mint/internal/httpx"
	"github.com/min0625/mint/internal/llm"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	defaultAPIPath   = "/v1/messages"
	defaultModelName = "claude-haiku-4-5"
	anthropicVersion = "2023-06-01"
	maxTokens        = 8192
	temperature      = 0.3
	// maxScanLineBytes raises bufio.Scanner's default 64KB line limit so a
	// large SSE data line or error body does not abort the stream early.
	maxScanLineBytes = 1 << 20
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
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	System      string    `json:"system,omitempty"`
	Messages    []message `json:"messages"`
	Stream      bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamDelta struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	StopReason string `json:"stop_reason"`
}

type streamError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type streamUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type streamMessage struct {
	Usage streamUsage `json:"usage"`
}

type streamEvent struct {
	Type    string        `json:"type"`
	Delta   streamDelta   `json:"delta"`
	Message streamMessage `json:"message"`
	Usage   streamUsage   `json:"usage"`
	Error   *streamError  `json:"error"`
}

// Complete calls the Anthropic API with streaming and writes tokens to w as they arrive.
// system is sent as the top-level system field; user is the single user message.
func (c *Client) Complete(ctx context.Context, system, user string, w io.Writer) (llm.Usage, error) {
	body := requestBody{
		Model:       c.modelName,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		System:      system,
		Messages:    []message{{Role: "user", Content: user}},
		Stream:      true,
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
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", anthropicVersion)

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

		var event streamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "error":
			// Mid-stream error event (e.g. overloaded_error): the HTTP status
			// was already 200, so this is the only failure signal we get.
			if event.Error != nil {
				return llm.Usage{}, fmt.Errorf("API stream error: %s: %s", event.Error.Type, event.Error.Message)
			}
		case "message_start":
			usage.InputTokens = event.Message.Usage.InputTokens
		case "message_delta":
			usage.OutputTokens = event.Usage.OutputTokens

			if event.Delta.StopReason == "max_tokens" {
				truncated = true
			}
		case "content_block_delta":
			if event.Delta.Type == "text_delta" {
				if _, err := fmt.Fprint(out, event.Delta.Text); err != nil {
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
	// this, a long input would print an incomplete translation and exit 0.
	if truncated {
		return usage, fmt.Errorf("output truncated: response hit the %d output-token limit", maxTokens)
	}

	return usage, nil
}
