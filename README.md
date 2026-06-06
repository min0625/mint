**English** | [繁體中文](https://github.com/min0625/mint/blob/main/README.zh-TW.md)

# 🌿 Mint

> Minimalist AI Translation CLI — Simple. Fast. Intuitive.

[![GitHub](https://img.shields.io/badge/GitHub-min0625%2Fmint-blue?logo=github)](https://github.com/min0625/mint)
[![Go Reference](https://pkg.go.dev/badge/github.com/min0625/mint.svg)](https://pkg.go.dev/github.com/min0625/mint)
[![PyPI](https://img.shields.io/pypi/v/mint-ai?logo=pypi&logoColor=white)](https://pypi.org/project/mint-ai/)
[![npm](https://img.shields.io/npm/v/mint-ai?logo=npm)](https://www.npmjs.com/package/mint-ai)

Mint is a lightweight, LLM-powered translation tool for the command line.
Choose your provider (Google Gemini, OpenAI, Anthropic, or local Ollama)
and get fluent translations instantly with optional smart language detection.

---

## ✨ Why Mint?

- **Minimal** — One command, no noise
- **Multi-provider** — Google Gemini, OpenAI, Anthropic, or local Ollama
- **Flexible** — Any language pair; smart auto-detection optional
- **Composable** — Pipe-friendly stdin/stdout, fits any workflow

---

## 📋 Installation

### pipx

```bash
pipx install mint-ai
```

### npm

```bash
npm install -g mint-ai
```

### Automated install (one-liner)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh)"
```

Auto-detects OS and architecture (Linux/macOS, x86_64/arm64),
and installs to `~/.local/bin`. Override with `MINT_INSTALL_DIR` or pin a version with `MINT_VERSION=v1.0.0`.

### go install

```bash
go install github.com/min0625/mint/cmd/mint@latest
```

Requires Go 1.21+. Binary lands in `$GOPATH/bin` (usually `~/go/bin`).

### Manual download

Pre-built binaries at [GitHub Releases](https://github.com/min0625/mint/releases)

### Verify installation

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

# Local Ollama (no API key needed)
export MINT_PROVIDER=ollama
export MINT_BASE_URL=http://localhost:11434
export MINT_MODEL_NAME=llama2
```

### 2. Translate

```bash
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

echo "The quick brown fox" | mint -t fr
cat document.txt | mint -t zh-TW
```

### 3. Smart language detection

```bash
export MINT_TARGET_LANG=en
mint "早安"             # Detects Chinese → translates to en
mint "Good mooorning"  # Detects English → corrects grammar & spelling

# Rotate across multiple targets
export MINT_TARGET_LANG=en,zh-TW,ja
mint "Hello"       # → zh-TW
mint "你好"        # → ja
mint "こんにちは"   # → en
```

---

## 🔑 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MINT_PROVIDER` | `google-genai` \| `openai` \| `anthropic` \| `ollama` | — (required) |
| `MINT_API_KEY` | API key (required for all providers except `ollama`) | — |
| `MINT_BASE_URL` | Custom endpoint (required for `ollama`) | Provider default |
| `MINT_MODEL_NAME` | Model to use | `gemini-3.1-flash-lite` / `gpt-4o-mini` / `claude-haiku-4-5` / none |
| `MINT_TARGET_LANG` | Target language(s), e.g. `en` or `en,zh-TW,ja` | System locale or `en` |

---

## 🎯 Design Principles

| Principle | Description |
|-----------|-------------|
| Zero-dependency install | Single binary, works out of the box |
| Multi-provider | Supports major LLM services plus local alternatives |
| Composability | Pairs seamlessly with `grep`, `sed`, `xargs`, and friends |
| Transparent output | Results to stdout, errors to stderr |
| Environment-friendly | API keys via env vars, no config file pollution |

---

## 🗺 Roadmap

- [x] Multi-LLM provider support (Google Gemini, OpenAI, Anthropic, Ollama)
- [x] Smart language detection and multi-language rotation via `MINT_TARGET_LANG`
- [x] Explicit target language via `--target` / `-t` flag
- [x] GoReleaser multi-platform binary release (Linux / macOS / Windows)
- [ ] Batch translation mode
- [ ] Glossary / custom dictionary support
- [ ] Output format options (plain text, JSON, Markdown)
- [ ] Caching for repeated translations

---

## 📄 License

Apache License 2.0 — see [LICENSE](https://github.com/min0625/mint/blob/main/LICENSE) for details.
