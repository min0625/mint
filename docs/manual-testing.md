# Manual Test Cases

AI agents and humans can use these cases to verify correct behavior.
Assumes `MINT_PROVIDER` and `MINT_API_KEY` are already set in the environment.

**Debugging tip:** add `-v` / `--verbose` to any command to print diagnostic info to stderr.
Use it first whenever output is unexpected:

Single-target mode (the common case with `--target` or a single `MINT_TARGET_LANG`):

```
[mint] provider: google-genai
[mint] single target — skipping language detection
[mint] target language: zh-TW
```

Multi-target mode (language rotation with `MINT_TARGET_LANG=zh-TW,en,...`):

```
[mint] provider: google-genai
[mint] detected input language: "en"
[mint] target language: zh-TW
```

---

## 1. Single-word translation (`-t` flag)

```sh
mint -t zh-TW "apple"    # 蘋果
mint -t zh-TW "りんご"   # 蘋果
mint -t en    "蘋果"      # apple
mint -t en    "りんご"    # apple
mint -t ja    "apple"     # りんご
mint -t ja    "蘋果"      # りんご
```

## 2. Sentence translation (`-t` flag)

```sh
mint -t zh-TW "This is an apple."     # 這是一顆蘋果。
mint -t zh-TW "これはリンゴです。"    # 這是一顆蘋果。
mint -t en    "這是一顆蘋果。"        # This is an apple.
mint -t en    "これはリンゴです。"    # This is an apple.
mint -t ja    "This is an apple."     # これはリンゴです。
mint -t ja    "這是一顆蘋果。"        # これはリンゴです。
```

## 3. Long-form flags

`--target` and `--verbose` are the long forms of `-t` and `-v`:

```sh
mint --target zh-TW "apple"                    # 蘋果
mint --target zh-TW --verbose "apple"          # 蘋果  (diagnostics on stderr)
```

`-t` accepts only a single language tag; a comma in the value is silently truncated to the first tag:

```sh
mint -t "zh-TW,en" "apple"    # 蘋果  (only zh-TW is used; rotation requires MINT_TARGET_LANG)
```

## 4. Single target language via `MINT_TARGET_LANG`

```sh
export MINT_TARGET_LANG=zh-TW
mint "apple"                   # 蘋果
mint "This is an apple."       # 這是一顆蘋果。
mint "これはリンゴです。"      # 這是一顆蘋果。

export MINT_TARGET_LANG=en
mint "蘋果"                    # apple
mint "這是一顆蘋果。"          # This is an apple.
mint "これはリンゴです。"      # This is an apple.

export MINT_TARGET_LANG=ja
mint "apple"                   # りんご
mint "This is an apple."       # これはリンゴです。
mint "這是一顆蘋果。"          # これはリンゴです。
```

## 5. Same-language: spelling and grammar correction

When the detected input language matches the target, the tool corrects rather than translates.
`-v` shows `single target — skipping language detection` and `target language: en`.

```sh
export MINT_TARGET_LANG=en
mint "This are an apple."      # This is an apple.

export MINT_TARGET_LANG=zh-TW
mint "這事一科蘋果"            # 這是一顆蘋果。

export MINT_TARGET_LANG=ja
mint "これわリンゴです。"      # これはリンゴです。
```

## 6. Same primary subtag → correction / script conversion

BCP-47 tags sharing the same primary subtag (e.g. `zh-HK` and `zh-TW`, both `zh`) are
treated as one slot in language rotation. With a single target the text is always rewritten
in the target tag's standard form.

```sh
# zh-HK input, zh-TW target → standardize to Traditional Chinese (not a translation)
export MINT_TARGET_LANG=zh-TW
mint "這係一個蘋果"            # 這是一個蘋果

# zh-CN (Simplified Chinese) input, zh-TW target → convert script to Traditional Chinese
mint "这是一个苹果"            # 這是一個蘋果
```

`-v` shows `single target — skipping language detection` and `target language: zh-TW` — the
rewrite prompt handles standardization without needing to detect the input language first.

> **Rotation note:** in a multi-language list (e.g. `zh-TW,en`), zh-HK input occupies
> the zh-TW slot and rotates to `en`, not to `zh-TW`. In multi-target mode `-v` confirms:
> `detected input language: "zh-HK"` → `target language: en`. See section 9 below.

## 7. Language-neutral pass-through

Numbers, symbols, and other language-agnostic content are printed unchanged with no LLM call.
`-v` confirms: `language-neutral content — outputting unchanged`.

- **Single-target mode** (`-t` or single `MINT_TARGET_LANG`): a local heuristic (no letters
  in the text) detects neutral content before any LLM call is made.
- **Multi-target mode** (rotation list): language detection runs first; if the model returns
  `neutral`, the text is output immediately with no second LLM call.

```sh
mint -t zh-TW "42"        # 42
mint -t zh-TW "3.14"      # 3.14
mint -t zh-TW "!@#$%"     # !@#$%
mint -t ja    "123-456"   # 123-456
```

## 8. Stdin / pipe input

Text can be piped from stdin instead of passed as a positional argument.

```sh
echo "apple"              | mint -t zh-TW   # 蘋果
echo "This is an apple."  | mint -t en      # This is an apple.
cat file.txt              | mint -t ja      # file contents translated to Japanese

# Multiline input
printf "First line.\nSecond line." | mint -t zh-TW   # two-line Traditional Chinese output
```

## 9. Language rotation — two languages

```sh
export MINT_TARGET_LANG=zh-TW,en

mint "This is an apple."     # 這是一顆蘋果。  (en matched at index 1 → next: zh-TW at index 0)
mint "これはリンゴです。"    # 這是一顆蘋果。  (ja not in list → first: zh-TW)
mint "這是一顆蘋果。"        # This is an apple.  (zh-TW matched at index 0 → next: en at index 1)
```

> **zh-HK edge case:** zh-HK input matches the zh-TW slot (same primary subtag `zh`) and
> therefore rotates to `en`. Use `-v` to confirm:
> `detected input language: "zh-HK"` → `target language: en`.

## 10. Language rotation — three languages (wrap-around)

```sh
export MINT_TARGET_LANG=zh-TW,en,ja

mint "This is an apple."     # これはリンゴです。  (en at index 1 → next: ja at index 2)
mint "這是一顆蘋果。"        # This is an apple.  (zh-TW at index 0 → next: en at index 1)
mint "これはリンゴです。"    # 這是一顆蘋果。  (ja at index 2 → wraps to: zh-TW at index 0)
```

## 11. `-t` flag overrides `MINT_TARGET_LANG`

```sh
export MINT_TARGET_LANG=ja
mint -t zh-TW "This is an apple."    # 這是一顆蘋果。  (flag wins over env var)
```

## 12. Force the source language (`-s` / `--source` flag)

`-s` / `--source` skips auto-detection and anchors the rewrite prompt to translate *from* the
given language. Use it for input that is also valid in the target language — cross-language
homographs and romanized text — which auto-detection would otherwise leave unchanged.

`-v` shows `source language: <tag>` plus `single target — skipping language detection` (single
target) or `explicit source — skipping language detection` (rotation).

```sh
# Cross-language homograph: French "pain" (bread) is spelled like English "pain" (ache)
mint -s fr -t en "pain"          # bread      (without -s → "pain", treated as English)
mint    -t en "pain"             # pain       (auto-detect leaves the English word unchanged)

# Romanized input that auto-detect would treat as already-English
mint -s ja -t en "konnichiwa"    # hello

# Source == target (exact same tag): a no-op translation falls back to correction
mint -s en -t en "This are an apple."    # This is an apple.

# Distinct tags sharing a primary subtag: a deliberate script conversion keeps the anchor
mint -s zh-CN -t zh-TW "这是一个苹果"    # 這是一個蘋果

# Rotation: an explicit source skips the detection call and picks the next tag
export MINT_TARGET_LANG=zh-TW,en
mint -s en "Hello"               # 你好  (en matched → next: zh-TW; -v: explicit source — skipping language detection)
```

`-s` accepts only a single language tag; like `-t`, a comma in the value is truncated to the first tag.
`--source` is the long form of `-s`. There is no `MINT_SOURCE_LANG` env var — a source is per-input, not a persistent preference.

> **Model note:** whether the homograph is actually translated depends on model quality, not
> on the CLI — `-v` confirms the anchor is sent (`source language: fr`) either way. Observed
> (2026-07): `gemini-3.1-flash-lite`, `claude-haiku-4-5`, and `gemma-4-e4b` return `bread`;
> `gpt-4o-mini` and `llama3.1:8b` keep `pain` unchanged. Treat a wrong result here as a
> model limitation, not a `-s` regression, unless the verbose anchor line is missing.

## 13. Error cases

All errors go to stderr; the process exits with code 1.

Config validation runs before input validation: if `MINT_PROVIDER` is unset, empty input
reports the provider error, not `no input text provided`. The cases below assume a valid
provider config is already in place (per the header of this file).

```sh
# Empty or whitespace-only input
mint -t zh-TW ""           # Error: no input text provided
mint -t zh-TW "   "        # Error: no input text provided
echo "" | mint -t zh-TW    # Error: no input text provided

# Missing or invalid provider
unset MINT_PROVIDER
mint -t zh-TW "apple"      # Error: MINT_PROVIDER environment variable is required

MINT_PROVIDER=invalid mint -t zh-TW "apple"
# Error: unsupported provider: invalid. Supported: google-genai, openai, anthropic

# Missing API key (no MINT_BASE_URL set)
unset MINT_API_KEY
mint -t zh-TW "apple"      # Error: MINT_API_KEY is required for provider: <provider>

# MINT_BASE_URL set: API key is optional (proxy handles auth), but a model
# name is required — a custom endpoint has no meaningful default model.
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=translategemma:4b
unset MINT_API_KEY
mint -t zh-TW "hello"      # 你好  (no API key required)

unset MINT_MODEL_NAME
mint -t zh-TW "hello"      # Error: MINT_MODEL_NAME is required when MINT_BASE_URL is set
```

Ctrl+C sends SIGINT and exits with code 130; SIGTERM exits with code 143 — the
conventional 128+N mapping, applied while an HTTP request is in progress so the two
signals stay distinguishable. A signal sent before the request starts is handled by
the OS default (exit 143 for SIGTERM, exit 130 for SIGINT).

## 14. `MINT_VERBOSE` environment variable

`MINT_VERBOSE=true` is equivalent to the `-v` / `--verbose` flag.

```sh
MINT_VERBOSE=true mint -t zh-TW "apple"   # 蘋果  (diagnostics on stderr)
```

Verbose stderr output (single-target mode):
```
[mint] provider: google-genai
[mint] model: gemini-3.1-flash-lite
[mint] single target — skipping language detection
[mint] target language: zh-TW
[mint] tokens: 125 in / 8 out
```

The `model:` line appears only when `MINT_MODEL_NAME` is set (likewise `base_url:` for
`MINT_BASE_URL`); with provider defaults both are omitted, as in the abbreviated example at
the top of this file.

## 15. Long-input chunked translation

Input longer than 2000 runes is split at paragraph boundaries (falling back to line
breaks, then spaces) and translated chunk by chunk — one LLM request per chunk — so a
long document cannot hit the model's output-token limit and come back truncated. Blank
lines between paragraphs are preserved in the output. Shorter input is unaffected:
one request, behavior identical to before.

```sh
# Build a ~4500-rune three-paragraph document, then translate it
python3 -c 'p = ("hello world " * 125).strip(); print(p + "\n\n" + p + "\n\n" + p, end="")' > /tmp/long.txt
mint -t zh-TW -v < /tmp/long.txt
# [mint] long input — split into 3 chunks (max 2000 runes each)
# ... three translated paragraphs, separated by blank lines
```

> **Against a real model** (not a mock), prefer a varied-content document over the
> repeated-phrase one above — e.g. `testdata/sample.txt` duplicated once
> (`cat testdata/sample.txt testdata/sample.txt`). A single phrase repeated dozens of
> times is degenerate input that can send some models (observed: `gpt-4o-mini`,
> `llama3.1:8b`) into a runaway repetition loop, producing a chunk many times longer
> than the input and taking minutes instead of seconds — a model quirk, not a chunking
> bug, but it can look like a hang during manual testing.

In rotation mode, language detection samples the first chunk that contains letters
(falling back to the first chunk if none do), so detection cost does not grow with
document length. If that sample comes back language-neutral, a single-chunk document is
printed unchanged as usual, but a multi-chunk document falls back to the first
configured target instead — a neutral sample doesn't mean the rest of a longer document
is neutral too. A language-neutral chunk (e.g. a block of numbers) is printed unchanged
with no LLM call. If a request fails mid-document, the error names the chunk:
`Error: translation failed: ... (chunk 2/3)`, and any tokens already spent on earlier
chunks — including the failing chunk's own usage, if the provider returned any alongside
the error (e.g. hitting the output-token limit) — are still reported in verbose mode. A
local output-write failure (e.g. a closed pipe) is named the same way but is not
mislabeled as a translation failure: `Error: write output: ... (chunk 2/3)`.
