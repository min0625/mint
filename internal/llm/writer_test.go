// Copyright 2026 The Mint Authors.

package llm_test

import (
	"strings"
	"testing"

	"github.com/min0625/mint/internal/llm"
)

func TestTrailingNewlineWriter(t *testing.T) {
	const want = "Hello\n"

	tests := []struct {
		name   string
		chunks []string
		want   string
	}{
		{name: "no trailing newline appends one", chunks: []string{"Hello"}, want: want},
		{name: "existing trailing newline kept as is", chunks: []string{"Hello\n"}, want: want},
		{name: "newline split across chunks not doubled", chunks: []string{"Hello", "\n"}, want: want},
		{name: "empty final chunk does not reset state", chunks: []string{"Hello\n", ""}, want: want},
		{name: "empty stream still terminated", chunks: nil, want: "\n"},
		{name: "internal newlines preserved", chunks: []string{"a\nb"}, want: "a\nb\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder

			out := llm.NewTrailingNewlineWriter(&sb)
			for _, c := range tt.chunks {
				if _, err := out.Write([]byte(c)); err != nil {
					t.Fatalf("Write returned error: %v", err)
				}
			}

			if err := out.Done(); err != nil {
				t.Fatalf("Done returned error: %v", err)
			}

			if got := sb.String(); got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
