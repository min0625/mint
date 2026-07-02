// Copyright 2026 The Mint Authors.

// Signal-delivery tests live behind a unix build tag: they send real POSIX
// signals to the test process via syscall.Kill, which does not exist on
// Windows (and neither does meaningful SIGINT/SIGTERM delivery there).

//go:build unix

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"
)

// TestRunInterrupted covers both signals run() registers (os.Interrupt and
// syscall.SIGTERM): either one must cancel an in-flight request and exit
// quietly with the conventional 128+N code (130 for SIGINT, 143 for SIGTERM).
func TestRunInterrupted(t *testing.T) {
	signals := []struct {
		name     string
		sig      syscall.Signal
		wantCode int
	}{
		{"SIGINT", syscall.SIGINT, 130},
		{"SIGTERM", syscall.SIGTERM, 143},
	}

	for _, tt := range signals {
		t.Run(tt.name, func(t *testing.T) {
			started := make(chan struct{})
			done := make(chan struct{})

			srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				close(started)
				<-done // held open until the test releases it, regardless of client-side cancellation
			}))

			defer func() {
				close(done)
				srv.Close()
			}()

			t.Setenv("MINT_PROVIDER", "openai")
			t.Setenv("MINT_API_KEY", "test")
			t.Setenv("MINT_BASE_URL", srv.URL)
			t.Setenv("MINT_MODEL_NAME", "test-model")

			old := os.Args

			os.Args = []string{"mint", "--target", "en", "hello"}
			defer func() { os.Args = old }()

			flushOut := captureStdout(t)
			flushErr := captureStderr(t)

			codeCh := make(chan int, 1)

			go func() { codeCh <- run() }()

			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("request never reached the server")
			}

			if err := syscall.Kill(os.Getpid(), tt.sig); err != nil {
				t.Fatalf("failed to send %s: %v", tt.name, err)
			}

			var code int

			select {
			case code = <-codeCh:
			case <-time.After(5 * time.Second):
				t.Fatalf("run() did not return after %s", tt.name)
			}

			stderrOutput := flushErr()
			_ = flushOut()

			if code != tt.wantCode {
				t.Errorf("expected exit code %d, got %d", tt.wantCode, code)
			}

			if stderrOutput != "" {
				t.Errorf("expected no stderr output on interrupt, got: %q", stderrOutput)
			}
		})
	}
}
