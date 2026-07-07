// Copyright 2026 The Mint Authors.
package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// reassemble concatenates text+sep over chunks; splitChunks guarantees this
// reproduces the original input exactly.
func reassemble(chunks []chunk) string {
	var sb strings.Builder
	for _, c := range chunks {
		sb.WriteString(c.text)
		sb.WriteString(c.sep)
	}

	return sb.String()
}

func TestSplitChunksShortInputSingleChunk(t *testing.T) {
	text := "Hello world\n\nSecond paragraph"

	chunks := splitChunks(text, 100)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	if chunks[0].text != text || chunks[0].sep != "" {
		t.Errorf("short input must pass through unchanged, got %+v", chunks[0])
	}
}

func TestSplitChunksParagraphBoundaries(t *testing.T) {
	text := "alpha one\n\nbravo two\n\ncandy six"

	chunks := splitChunks(text, 10)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d: %+v", len(chunks), chunks)
	}

	want := []chunk{
		{text: "alpha one", sep: "\n\n"},
		{text: "bravo two", sep: "\n\n"},
		{text: "candy six", sep: ""},
	}
	for i, c := range chunks {
		if c != want[i] {
			t.Errorf("chunk %d = %+v, want %+v", i, c, want[i])
		}
	}
}

func TestSplitChunksPacksSmallParagraphs(t *testing.T) {
	text := "aa\n\nbb\n\ncc\n\ndddddddddd"

	chunks := splitChunks(text, 10)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d: %+v", len(chunks), chunks)
	}

	// The three short paragraphs fit into one 10-rune chunk ("aa\n\nbb" is 6
	// runes; adding "\n\ncc" would make 10 — exactly at the limit).
	if chunks[0].text != "aa\n\nbb\n\ncc" {
		t.Errorf("chunk 0 = %q, want packed paragraphs", chunks[0].text)
	}

	if reassemble(chunks) != text {
		t.Errorf("reassembled = %q, want original", reassemble(chunks))
	}
}

func TestSplitChunksOversizedParagraphSplitsOnLines(t *testing.T) {
	text := "line one is here\nline two is here\n\nshort"

	chunks := splitChunks(text, 20)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d: %+v", len(chunks), chunks)
	}

	want := []chunk{
		{text: "line one is here", sep: "\n"},
		{text: "line two is here", sep: "\n\n"},
		{text: "short", sep: ""},
	}
	for i, c := range chunks {
		if c != want[i] {
			t.Errorf("chunk %d = %+v, want %+v", i, c, want[i])
		}
	}
}

func TestSplitChunksLongLineBreaksAtSpaces(t *testing.T) {
	text := strings.TrimSuffix(strings.Repeat("word ", 20), " ") // 99 runes

	chunks := splitChunks(text, 30)
	if len(chunks) < 4 {
		t.Fatalf("expected at least 4 chunks, got %d", len(chunks))
	}

	for i, c := range chunks {
		if n := utf8.RuneCountInString(c.text); n > 30 {
			t.Errorf("chunk %d has %d runes, exceeds max 30", i, n)
		}

		// Every piece except the last must break at a space, and that space
		// must live in sep (not text), so it's never sent to the LLM.
		if i < len(chunks)-1 {
			if strings.HasSuffix(c.text, " ") {
				t.Errorf("chunk %d text = %q retains the breaking space, want it moved to sep", i, c.text)
			}

			if c.sep != " " {
				t.Errorf("chunk %d sep = %q, want a single space", i, c.sep)
			}
		}
	}

	if reassemble(chunks) != text {
		t.Errorf("reassembled = %q, want original", reassemble(chunks))
	}
}

func TestSplitChunksCJKHardSplit(t *testing.T) {
	// No whitespace at all: must hard-split at the rune limit without
	// corrupting any multi-byte rune.
	text := strings.Repeat("翻譯測試", 25) // 100 runes

	chunks := splitChunks(text, 30)
	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}

	for i, c := range chunks {
		if !utf8.ValidString(c.text) {
			t.Errorf("chunk %d is not valid UTF-8", i)
		}

		if n := utf8.RuneCountInString(c.text); n > 30 {
			t.Errorf("chunk %d has %d runes, exceeds max 30", i, n)
		}
	}

	if reassemble(chunks) != text {
		t.Error("reassembled text differs from original")
	}
}

func TestSplitChunksPreservesNewlineRuns(t *testing.T) {
	// Three blank lines between paragraphs (a run of 4 newlines) must be
	// preserved in the separator so output spacing can be restored.
	text := "alpha one\n\n\n\nbravo two"

	chunks := splitChunks(text, 10)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d: %+v", len(chunks), chunks)
	}

	if chunks[0].sep != "\n\n\n\n" {
		t.Errorf("sep = %q, want the full newline run", chunks[0].sep)
	}

	if reassemble(chunks) != text {
		t.Error("reassembled text differs from original")
	}
}

func TestSplitChunksReassemblyExact(t *testing.T) {
	inputs := []string{
		"plain short text",
		"para one\n\npara two\n\n\npara three",
		strings.Repeat("mixed 混合 text ", 40),
		"leading\n\n\n" + strings.Repeat("x", 50) + "\ntrailing",
		"\n\nstarts with blank lines",
		"trailing newline run\n\n",
		"trailing newline run\n\n\n\n",
		"para one\r\n\r\npara two",
		"line one\r\nline two\r\n\r\nshort",
		// Oversized paragraph whose text itself ends in a single trailing
		// newline (no blank-line run anywhere), the case subdivideUnit must
		// not lose when reattaching the (empty) paragraph-level separator.
		strings.Repeat("x", 15) + " " + strings.Repeat("y", 15) + "\n",
		"line one is here\nline two is here\n",
	}
	for _, text := range inputs {
		for _, max := range []int{10, 25, 80} {
			if got := reassemble(splitChunks(text, max)); got != text {
				t.Errorf("splitChunks(%q, %d): reassembled %q != original", text, max, got)
			}
		}
	}
}

func TestSplitChunksTrailingNewlineRunHasNoEmptyChunk(t *testing.T) {
	// A document whose last paragraph is followed by a newline run must not
	// produce a trailing empty chunk — that run belongs to the last real
	// chunk's sep.
	text := "alpha one\n\nbravo two\n\n"

	chunks := splitChunks(text, 10)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d: %+v", len(chunks), chunks)
	}

	if chunks[1].text == "" {
		t.Errorf("chunk 1 = %+v, want no trailing empty chunk", chunks[1])
	}

	if chunks[1].sep != "\n\n" {
		t.Errorf("chunk 1 sep = %q, want the trailing newline run", chunks[1].sep)
	}

	if reassemble(chunks) != text {
		t.Errorf("reassembled = %q, want original", reassemble(chunks))
	}
}

// TestSplitChunksLeadingBlankLineRunKeptAsSeparator verifies that a blank-line
// run before a long document's first paragraph becomes its own empty-text
// chunk carrying the run as sep, rather than being folded into the first
// real chunk's text — which would send the document's leading whitespace to
// the LLM with no instruction to preserve it.
func TestSplitChunksLeadingBlankLineRunKeptAsSeparator(t *testing.T) {
	para := strings.TrimSuffix(strings.Repeat("hello world ", 200), " ") // > 2000 runes
	text := "\n\n" + para

	chunks := splitChunks(text, 2000)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}

	if chunks[0].text != "" || chunks[0].sep != "\n\n" {
		t.Errorf("chunk 0 = %+v, want the leading blank-line run isolated as its own empty-text chunk", chunks[0])
	}

	if strings.HasPrefix(chunks[1].text, "\n") {
		t.Errorf("chunk 1 text = %q, must not carry the leading separator into real content", chunks[1].text)
	}

	if reassemble(chunks) != text {
		t.Errorf("reassembled = %q, want original", reassemble(chunks))
	}
}

func TestSplitChunksCRLFParagraphBoundaries(t *testing.T) {
	// CRLF blank lines must be recognized as paragraph boundaries just like
	// bare "\n\n", with the "\r" bytes captured in sep rather than leaking
	// into the text sent to the LLM.
	text := "alpha one\r\n\r\nbravo two"

	chunks := splitChunks(text, 10)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d: %+v", len(chunks), chunks)
	}

	want := []chunk{
		{text: "alpha one", sep: "\r\n\r\n"},
		{text: "bravo two", sep: ""},
	}
	for i, c := range chunks {
		if c != want[i] {
			t.Errorf("chunk %d = %+v, want %+v", i, c, want[i])
		}
	}

	if reassemble(chunks) != text {
		t.Errorf("reassembled = %q, want original", reassemble(chunks))
	}
}

// TestSplitChunksOversizedParagraphTrailingNewlinePreserved verifies that an
// oversized paragraph whose text ends in a single trailing newline (only
// possible for the document's final unit, where there is no paragraph-level
// separator to reattach) keeps that newline instead of losing it when
// subdivideUnit reattaches its own (empty) separator.
func TestSplitChunksOversizedParagraphTrailingNewlinePreserved(t *testing.T) {
	text := "line one is here\nline two is here\n"

	chunks := splitChunks(text, 20)
	if got := reassemble(chunks); got != text {
		t.Errorf("reassemble(splitChunks(%q, 20)) = %q, want original (trailing newline lost)", text, got)
	}
}

// TestSplitChunksLineSplitDoesNotCorruptWordWhenSpaceExists verifies that
// splitLine finds a break point anywhere in the window (not just its second
// half), so a word shorter than the limit is never cut mid-word just because
// its only available space sits early in the window.
func TestSplitChunksLineSplitDoesNotCorruptWordWhenSpaceExists(t *testing.T) {
	chunks := splitChunks("a bcde", 4)
	for _, c := range chunks {
		if c.text == "a bc" || c.text == "de" {
			t.Errorf("chunks = %+v, want the word \"bcde\" (4 runes, within limit) kept intact", chunks)
		}
	}

	if got := reassemble(chunks); got != "a bcde" {
		t.Errorf("reassembled = %q, want original", got)
	}
}

// TestSplitChunksLineSplitFindsSpaceAtLimitBoundary verifies that a
// whitespace rune sitting exactly at the limit boundary is used as the break
// point, rather than being missed and left as a leading space on the next
// piece (which would fuse translated pieces together, since the piece's own
// sep would then be empty).
func TestSplitChunksLineSplitFindsSpaceAtLimitBoundary(t *testing.T) {
	chunks := splitChunks("abcd efg", 4)

	want := []chunk{
		{text: "abcd", sep: " "},
		{text: "efg", sep: ""},
	}
	for i, c := range chunks {
		if c != want[i] {
			t.Errorf("chunk %d = %+v, want %+v", i, c, want[i])
		}
	}

	if got := reassemble(chunks); got != "abcd efg" {
		t.Errorf("reassembled = %q, want original", got)
	}
}

func TestIndexNewlineRun(t *testing.T) {
	tests := []struct {
		s         string
		minRun    int
		wantStart int
		wantEnd   int
	}{
		{"abc", 1, -1, -1},
		{"a\nb", 1, 1, 2},
		{"a\nb", 2, -1, -1},
		{"a\n\nb", 2, 1, 3},
		{"ab\n\n\ncd", 2, 2, 5},
		{"\n\nab", 2, 0, 2},
		{"a\nb\n\nc", 2, 3, 5},
		// CRLF: "\r\n" is one newline token.
		{"a\r\nb", 1, 1, 3},
		{"a\r\n\r\nb", 2, 1, 5},
		{"a\r\nb", 2, -1, -1},
		// A lone "\r" not followed by "\n" is not a newline token.
		{"a\rb", 1, -1, -1},
	}
	for _, tt := range tests {
		start, end := indexNewlineRun(tt.s, tt.minRun)
		if start != tt.wantStart || end != tt.wantEnd {
			t.Errorf("indexNewlineRun(%q, %d) = (%d, %d), want (%d, %d)",
				tt.s, tt.minRun, start, end, tt.wantStart, tt.wantEnd)
		}
	}
}
