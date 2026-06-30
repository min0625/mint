// Copyright 2026 The Mint Authors.

package llm

import "io"

// TrailingNewlineWriter wraps an io.Writer and guarantees the stream ends with
// exactly one '\n'. Provider backends stream model tokens through it and call
// Done once the stream is complete: Done appends a newline unless the model
// already ended on one, avoiding a spurious blank trailing line. When nothing
// was written, Done still emits a single newline so callers always get a
// terminated line.
//
// Centralizing this here keeps every provider's streaming loop identical and
// removes the per-byte index bookkeeping (and its empty-chunk edge case) from
// each backend.
type TrailingNewlineWriter struct {
	w        io.Writer
	lastByte byte
	wrote    bool
}

// NewTrailingNewlineWriter returns a TrailingNewlineWriter that writes to w.
func NewTrailingNewlineWriter(w io.Writer) *TrailingNewlineWriter {
	return &TrailingNewlineWriter{w: w}
}

// Write forwards p to the underlying writer, tracking the last byte actually
// written so Done can decide whether a terminating newline is needed.
func (t *TrailingNewlineWriter) Write(p []byte) (int, error) {
	n, err := t.w.Write(p)
	if n > 0 {
		t.lastByte = p[n-1]
		t.wrote = true
	}

	return n, err
}

// Done writes a terminating newline unless the stream already ended with one.
// It must be called once, after the final Write.
func (t *TrailingNewlineWriter) Done() error {
	if t.wrote && t.lastByte == '\n' {
		return nil
	}

	_, err := io.WriteString(t.w, "\n")

	return err
}
