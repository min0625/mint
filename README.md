🌐 [Other languages](https://github.com/min0625/mint/blob/main/LANGUAGES.md)

# 🌿 Mint

> Minimalist AI Translation CLI — Simple. Fast. Intuitive.

[![GitHub Release](https://img.shields.io/github/v/release/min0625/mint?logo=github)](https://github.com/min0625/mint/releases)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)
[![codecov](https://codecov.io/gh/min0625/mint/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/mint)

Mint is a single-binary, LLM-powered translation CLI. Set two environment variables and translate anything from the command line — files, piped output, or inline text. Built-in language detection, grammar correction, streaming output, and multi-language rotation.

```bash
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_key

mint -t ja "Good morning"         # おはようございます
echo "早安" | mint -t en          # Good morning
cat document.txt | mint -t fr     # translate a whole file
```

---

## ✨ Why Mint?

- **Zero-config** — Single binary; API keys via env vars, no config file pollution
- **Multi-provider** — Google Gemini, OpenAI, Anthropic, or local Ollama / LM Studio
- **Smart detection** — Auto-detects language on every call; language-neutral content (numbers, symbols) passes through unchanged
- **Smart correction** — Same-language input? Auto-corrects grammar & spelling instead of translating
- **Streaming** — Output streams in real-time, no waiting for long translations
- **Composable** — Pipe-friendly stdin/stdout; pairs seamlessly with `grep`, `sed`, `xargs`, and friends
- **Secure** — Untrusted input is isolated from model instructions via system/user message separation and per-request random-nonce delimiters; translating adversarial content cannot hijack the LLM's behavior

---

## 📋 Installation

### Automated install (recommended)

**macOS / Linux**

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

Auto-detects OS and architecture (Linux/macOS, x86_64/arm64), installs to `~/.local/bin`. Override with `MINT_INSTALL_DIR` or pin a version with `MINT_VERSION=v1.0.0`.

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/min0625/mint/main/script/install.ps1 | iex
```

Auto-detects architecture (x86_64/arm64) and installs to `$HOME\.local\bin`. Override with `$env:MINT_INSTALL_DIR` or pin a version with `$env:MINT_VERSION = 'v1.0.0'`.

### Homebrew (macOS / Linux)

```bash
brew install min0625/tap/mint-ai
```

### pipx

```bash
pipx install mint-ai
```

### npm

```bash
npm install -g mint-ai
```

### Manual download

Download the pre-built binary for your platform from [GitHub Releases](https://github.com/min0625/mint/releases), move it into a directory on your `PATH`, then verify:

```bash
mint --version
```

---

## 🚀 Quick Start

### 1. Set your provider

```bash
# Google Gemini (free tier available — https://aistudio.google.com/apikey)
export MINT_PROVIDER=google-genai
export MINT_API_KEY=your_gemini_api_key

# OpenAI
export MINT_PROVIDER=openai
export MINT_API_KEY=sk-...

# Anthropic
export MINT_PROVIDER=anthropic
export MINT_API_KEY=sk-ant-...

# Ollama (no API key needed)
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=qwen2.5:7b  # use any model loaded in Ollama

# LM Studio (no API key needed)
export MINT_PROVIDER=openai
export MINT_BASE_URL=http://localhost:1234
export MINT_MODEL_NAME=lmstudio-community/Qwen2.5-7B-Instruct-GGUF  # use any model loaded in LM Studio
```

### 2. Translate

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

Use `--verbose` / `-v` (or `MINT_VERBOSE=true`) to print diagnostic info and token usage to stderr:

```bash
mint -t ja -v "Good morning"
# [mint] provider: google-genai
# [mint] model: gemini-3.1-flash-lite
# [mint] single target — skipping language detection
# [mint] target language: ja
# おはようございます
# [mint] tokens: 113 in / 2 out
```

**Typical token usage** (measured on `gemini-3.1-flash-lite`):

| Mode | Input | Calls | Input tokens | Output tokens |
|------|-------|-------|-------------|---------------|
| Single-target (`-t` or single `MINT_TARGET_LANG`) | short word/sentence | 1 | ~110–130 | ~1–15 |
| Single-target | long article (`testdata/sample.txt`) | 1 | ~465–470 | ~450–560 |
| Multi-target rotation (comma-separated `MINT_TARGET_LANG`) | short sentence | 2 | ~250–260 | ~2–8 |
| Explicit source `-s` + rotation | short sentence | 1 | ~105–120 | ~1–2 |

> Token counts scale with input length. Output tokens vary by target language — Japanese and Chinese tend to produce more tokens than English for equivalent content.

**How far does 1M tokens go?** (input + output combined, derived from the measured usage above):

| Input | ~Tokens per translation | Translations per 1M tokens |
|-------|------------------------|----------------------------|
| Short word or phrase | ~120 | ~8,000 |
| 300-word article | ~1,000 | ~1,000 |

> Counts combine input and output tokens. Providers price input and output separately and many offer free tiers — check your provider's pricing page for current rates. Google Gemini's free tier at [Google AI Studio](https://aistudio.google.com/apikey) needs no credit card.

**Force the source language** with `--source` / `-s` to translate input that is also valid in the target language (cross-language homographs, romanized text):

```bash
mint -s fr -t en "pain"          # French → bread (without -s, treated as English "pain")
mint -s ja -t en "konnichiwa"    # romaji Japanese → hello
```

### 3. Smart language detection

**Translation with auto-detection:**

```bash
export MINT_TARGET_LANG=en

mint "早安"   # Detects Chinese → Good morning
```

**Grammar & spelling correction** — when input language matches the target, Mint corrects instead of translates:

```bash
export MINT_TARGET_LANG=en

mint "Good mooorning"          # Detects English → Good morning
mint "She don't know nothing"  # Detects English → She doesn't know anything
mint "i luv coding"            # Detects English → I love coding
```

**Language rotation** — translates to the next language in the list, wrapping around:

```bash
# Two languages
export MINT_TARGET_LANG=en,zh-TW
mint "Hello"   # en → zh-TW: 你好
mint "你好"    # zh-TW → en: Hello

# Three languages
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # en → zh-TW: 你好
mint "你好"        # zh-TW → ja: こんにちは
mint "こんにちは"   # ja → en: Hello
```

---

## 🔑 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` | — (required) |
| `MINT_API_KEY` | API key; required when using the default endpoint; optional when `MINT_BASE_URL` is set (proxy handles auth) | — |
| `MINT_BASE_URL` | Custom API base URL (domain only; each provider appends its own path); use with `openai` to target Ollama (`http://localhost:11434`), LM Studio (`http://localhost:1234`), or any other OpenAI-compatible endpoint | Provider default |
| `MINT_MODEL_NAME` | Model to use; required when `MINT_BASE_URL` is set | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` |
| `MINT_TARGET_LANG` | Target language(s), e.g. `en` or `en,zh-TW,ja` | System locale |
| `MINT_VERBOSE` | Set to `true` to enable verbose diagnostic output (equivalent to `--verbose`) | `false` |

---

## 🚩 CLI Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--target <lang>` | `-t` | Target language (BCP-47 tag, e.g. `ja`, `zh-TW`, `fr`). Overrides `MINT_TARGET_LANG`. |
| `--source <lang>` | `-s` | Source language (BCP-47 tag); skips auto-detection and forces translation from this language. |
| `--verbose` | `-v` | Print diagnostic info and token usage to stderr. Also enabled by `MINT_VERBOSE=true`. |
| `--version` | | Print version and exit. |

---

## 📅 Roadmap

- [x] Multi-LLM provider support (Google Gemini, OpenAI, Anthropic, local via Ollama / LM Studio)
- [x] Smart language detection and multi-language rotation via `MINT_TARGET_LANG`
- [x] Explicit target language via `--target` / `-t` flag
- [x] Explicit source language via `--source` / `-s` flag
- [x] Streaming output
- [x] GoReleaser multi-platform binary release (Linux / macOS / Windows)
- [ ] Batch translation mode
- [ ] Glossary / custom dictionary support
- [ ] Output format options (plain text, JSON, Markdown)
- [ ] Caching for repeated translations

---

## 📄 License

Apache License 2.0 — see [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) for details.
