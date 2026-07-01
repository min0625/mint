// Copyright 2026 The Mint Authors.

package httpx_test

import (
	"net/http"
	"testing"

	"github.com/min0625/mint/internal/httpx"
)

func TestNew(t *testing.T) {
	c := httpx.New()
	if c == nil {
		t.Fatal("expected non-nil client")
	}

	// No overall timeout: a slow or cold-starting backend must be allowed to
	// stream for as long as it needs; cancellation is the caller's job via ctx.
	if c.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0 (no overall timeout)", c.Timeout)
	}

	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport = %T, want *http.Transport", c.Transport)
	}

	// Connection setup (DNS/dial, TLS handshake) must still be bounded so an
	// unreachable endpoint fails fast instead of hanging indefinitely.
	if tr.TLSHandshakeTimeout == 0 {
		t.Error("TLSHandshakeTimeout must be set")
	}

	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = false, want true")
	}
}

func TestNewReturnsDistinctClients(t *testing.T) {
	a := httpx.New()
	b := httpx.New()

	if a == b {
		t.Error("expected each call to New to return a distinct client")
	}
}
