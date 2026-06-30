// Copyright 2026 The Mint Authors.

// Package httpx builds the shared *http.Client used by every provider backend.
package httpx

import (
	"net"
	"net/http"
	"time"
)

// New returns an *http.Client tuned for streaming LLM responses.
//
// It bounds connection setup — DNS/dial and the TLS handshake — so an
// unreachable or half-open endpoint fails fast instead of hanging until the
// user presses Ctrl+C. It deliberately sets no overall Timeout and no
// ResponseHeaderTimeout: a slow or cold-starting local model may take a long
// time to load and stream, and the CLI is designed to wait as long as the
// backend needs. Per-request cancellation is handled by the context.
func New() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
