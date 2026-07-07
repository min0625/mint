// Copyright 2026 The Mint Authors.
package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxChunkRunes caps how much text goes into a single translation request.
// Longer input is split into chunks and translated chunk by chunk, so a long
// document cannot hit the model's output-token limit and come back truncated.
// CJK text runs roughly one token per rune, so a chunk stays near 2k output
// tokens — inside the default context of even small local models.
const maxChunkRunes = 2000

// chunk is one translation unit plus the exact separator text (a paragraph
// gap, a line break, a single space, or nothing) that followed it in the
// original input, kept so the output can restore paragraph and word
// separation exactly as it appeared, regardless of how a chunk was split.
type chunk struct {
	text string
	sep  string
}

// splitChunks splits text into translation units of at most limit runes each.
// Splitting prefers paragraph boundaries (blank lines), then line breaks,
// then whitespace within a line; only a single word longer than the limit is
// ever cut mid-word. Adjacent small units are re-packed greedily so short
// paragraphs still share one request. Concatenating text+sep over the result
// reproduces the input exactly.
func splitChunks(text string, limit int) []chunk {
	if limit < 1 {
		limit = 1
	}

	if utf8.RuneCountInString(text) <= limit {
		return []chunk{{text: text}}
	}

	var units []chunk
	for _, p := range splitOnNewlineRuns(text, 2) {
		units = append(units, subdivideUnit(p, limit)...)
	}

	return packUnits(units, limit)
}

// newlineWidth returns the byte length of the newline token starting at
// s[i]: 2 for "\r\n", 1 for a bare "\n", or 0 if s[i] does not start a
// newline. Treating "\r\n" as a single token means Windows line endings are
// recognized as paragraph and line boundaries exactly like Unix ones.
func newlineWidth(s string, i int) int {
	if i >= len(s) {
		return 0
	}

	if s[i] == '\n' {
		return 1
	}

	if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
		return 2
	}

	return 0
}

// indexNewlineRun returns the byte range [start, end) of the first run of at
// least minRun consecutive newline tokens ("\n" or "\r\n") in s, or (-1, -1)
// if there is none. The returned range extends through the whole run, not
// just its first minRun tokens, so e.g. three blank lines are captured
// together.
func indexNewlineRun(s string, minRun int) (start, end int) {
	start, run := -1, 0

	for i := 0; i < len(s); {
		w := newlineWidth(s, i)
		if w == 0 {
			run, start = 0, -1
			i++

			continue
		}

		if run == 0 {
			start = i
		}

		run++
		i += w

		if run == minRun {
			for {
				w := newlineWidth(s, i)
				if w == 0 {
					return start, i
				}

				i += w
			}
		}
	}

	return -1, -1
}

// splitOnNewlineRuns splits text into units at every run of at least minRun
// consecutive newline tokens; each unit keeps the newline run that followed
// it as sep. A trailing empty unit is never produced when text ends exactly
// on a newline run — that run is captured as the last real unit's sep
// instead, matching the invariant that every unit carries translatable text.
func splitOnNewlineRuns(text string, minRun int) []chunk {
	var units []chunk

	rest := text

	for {
		i, j := indexNewlineRun(rest, minRun)
		if i < 0 {
			break
		}

		units = append(units, chunk{text: rest[:i], sep: rest[i:j]})
		rest = rest[j:]
	}

	if rest != "" || len(units) == 0 {
		units = append(units, chunk{text: rest})
	}

	return units
}

// subdivideUnit splits a paragraph unit that exceeds limit runes at single line
// breaks, then hard-splits any line that is still too long. The unit's own
// trailing separator is reattached to its last piece.
func subdivideUnit(u chunk, limit int) []chunk {
	if utf8.RuneCountInString(u.text) <= limit {
		return []chunk{u}
	}

	var out []chunk
	for _, line := range splitOnNewlineRuns(u.text, 1) {
		out = append(out, splitLine(line, limit)...)
	}

	// The paragraph's own trailing separator (u.sep) and the last piece's own
	// sep are never both non-empty: u.sep is only non-empty when this unit is
	// followed by more of the original document, in which case
	// splitOnNewlineRuns(u.text, 1) already leaves the last piece's sep empty
	// (u.text ends on non-newline content); conversely a non-empty last-piece
	// sep only arises when u.text itself ends on a trailing newline token,
	// which only happens for the document's final unit, where u.sep is "".
	// Concatenating rather than overwriting keeps whichever one is real.
	out[len(out)-1].sep += u.sep

	return out
}

// splitLine splits a single line (no newline token) into pieces of at most
// limit runes, breaking at the last whitespace rune at or before the limit
// when one exists, so words stay intact. That whitespace rune becomes the
// piece's sep rather than staying in its text, so it is never sent to the
// LLM and reassembly rejoins the pieces exactly as the input had them. When
// no whitespace is found, the line is hard-split with sep = "". The line's
// own trailing separator is reattached to the last piece.
func splitLine(u chunk, limit int) []chunk {
	runes := []rune(u.text)
	if len(runes) <= limit {
		return []chunk{u}
	}

	var out []chunk

	for len(runes) > limit {
		cut, sep := limit, ""

		for k := limit; k >= 0; k-- {
			if unicode.IsSpace(runes[k]) {
				cut = k
				sep = string(runes[k])

				break
			}
		}

		out = append(out, chunk{text: string(runes[:cut]), sep: sep})
		runes = runes[cut+utf8.RuneCountInString(sep):]
	}

	return append(out, chunk{text: string(runes), sep: u.sep})
}

// packUnits greedily merges adjacent units into chunks of at most limit runes,
// so a long document of short paragraphs still needs only a few requests.
//
// A leading unit with empty text only occurs when the document starts with a
// paragraph-boundary newline run; unlike an empty unit produced between two
// real units (whose surrounding separators are meaningfully packed as literal
// text into one translation request), this one has no preceding real content
// to attach its separator to. Packing it in like any other unit would fold
// the document's leading whitespace into the first real chunk's text and
// send it to the LLM, which has no instruction to preserve it. Emitting it
// as its own chunk lets translateChunks reprint its separator verbatim
// before the first real chunk, exactly like an interior separator.
func packUnits(units []chunk, limit int) []chunk {
	if len(units) == 0 {
		return nil
	}

	var out []chunk

	start := 0

	if units[0].text == "" {
		out = append(out, units[0])
		start = 1
	}

	if start >= len(units) {
		return out
	}

	var b strings.Builder

	b.WriteString(units[start].text)

	curLen := utf8.RuneCountInString(units[start].text)
	curSep := units[start].sep

	for _, u := range units[start+1:] {
		sepLen := utf8.RuneCountInString(curSep)
		uLen := utf8.RuneCountInString(u.text)

		if curLen+sepLen+uLen <= limit {
			b.WriteString(curSep)
			b.WriteString(u.text)
			curSep = u.sep
			curLen += sepLen + uLen

			continue
		}

		out = append(out, chunk{text: b.String(), sep: curSep})

		b.Reset()
		b.WriteString(u.text)
		curLen, curSep = uLen, u.sep
	}

	return append(out, chunk{text: b.String(), sep: curSep})
}
