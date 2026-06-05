**English** | [繁體中文](./README.zh-TW.md)

# 🌿 Mint

> Minimalist AI Translation CLI — Simple. Fast. Intuitive.

---

Mint is a lightweight, LLM-powered translation tool for the command line.
Choose your favorite LLM provider (Google Gemini, OpenAI, Anthropic, or local Ollama),
and get fluent, natural translations instantly with optional smart language detection.

---

## ✨ Why Mint?

Most translation tools are either too bloated or too locked into a specific platform.
Mint is built around a single philosophy: **do less, do it well.**

- **Minimal** — One command, no noise
- **Fast** — Calls the LLM API directly, no intermediate layers
- **Multi-provider** — Choose between Google Gemini, OpenAI, Anthropic, or local Ollama models
- **Flexible** — Supports any language pair, freely specify your target language
- **Smart detection** — Optionally detect input language and auto-translate to your preference
- **Composable** — Pipe-friendly stdin/stdout design, fits naturally into any workflow

---

## 📋 Installation
### Automated Install (One-liner)

The easiest way to install — downloads the latest binary automatically:

```bash
curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh | bash
```

Features:
- Detects your OS and architecture automatically (Linux/macOS, x86_64/arm64)
- Verifies SHA256 checksums
- Installs to `~/.local/bin` by default (override with `MINT_INSTALL_DIR`)
- Shows PATH setup hints if needed
- Supports pinning a specific version: `MINT_VERSION=v1.0.0 bash script/install.sh`

### go install

If you have Go 1.21+ installed:

```bash
go install github.com/min0625/mint/cmd/mint@latest
```

The binary will be available as `mint` in your `$GOPATH/bin` directory (usually `~/go/bin`).

### Manual Download from GitHub Releases

Download pre-built binaries directly from [GitHub Releases](https://github.com/min0625/mint/releases):

```bash
# Linux x86_64
curl -L https://github.com/min0625/mint/releases/latest/download/mint_Linux_x86_64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# macOS arm64 (Apple Silicon)
curl -L https://github.com/min0625/mint/releases/latest/download/mint_Darwin_arm64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# macOS x86_64 (Intel)
curl -L https://github.com/min0625/mint/releases/latest/download/mint_Darwin_x86_64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# Windows x86_64 (PowerShell)
# Download mint_Windows_x86_64.zip from releases page and extract to a directory in your PATH
```

### Verify Installation

```bash
mint --version
```

---

## 🚀 Quick Start

### 1. Choose your provider

```bash
# Google Gemini (free tier available)
export MINT_PROVIDER=google
export MINT_API_KEY=your_gemini_api_key
# Get a free API key at: https://aistudio.google.com/apikey

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

### 2. Translate with explicit target language

```bash
# Specify a target language (BCP-47 tag) using --target or -t flag
mint --target ja "Good morning"
mint -t zh-TW "Good morning"

# Pipe from stdin
echo "The quick brown fox" | mint -t fr

# Translate a file
cat document.txt | mint -t zh-TW
```

### 3. Smart language detection (optional)

Set your target language preference using `MINT_TARGET_LANG`:

```bash
# Single target language
export MINT_TARGET_LANG=zh-TW
mint "Good morning"    # Detects English → translates to zh-TW
mint "早安"            # Detects Chinese → grammar & spelling correction

# Multiple target languages (language rotation)
export MINT_TARGET_LANG=en,zh-TW,ja

mint "Hello"           # English input → translates to zh-TW (next in rotation)
mint "你好"            # Chinese input → translates to ja (next in rotation)
mint "こんにちは"     # Japanese input → translates to en (wraps around)
```

The tool automatically detects the input language and applies the appropriate transformation.

---

## 🔑 Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `MINT_PROVIDER` | LLM provider: `google`, `openai`, `anthropic`, `ollama` | Yes | — |
| `MINT_API_KEY` | API key for the chosen provider | Conditional* | — |
| `MINT_BASE_URL` | Custom API endpoint; required for `ollama` (e.g., for self-hosted or local services) | Conditional* | Provider default |
| `MINT_MODEL_NAME` | LLM model name to use | Optional | Provider default** |
| `MINT_TARGET_LANG` | Target language(s) - single or comma-separated (e.g. `en`, `en,zh-TW,ja`) | Optional | System locale or `en` |

**Conditional:* `MINT_API_KEY` required for `google`, `openai`, `anthropic`; not needed for `ollama`. `MINT_BASE_URL` required for `ollama`.*
**Default models:* `google`: `gemini-3.1-flash-lite`, `openai`: `gpt-4o-mini`, `anthropic`: `claude-haiku-4-5`; `ollama`: none (must specify).*

### Language Resolution Priority

The tool uses the following priority order to determine the target language(s):

1. **Flag**: `--target` / `-t` CLI flag (highest priority)
2. **Config**: `MINT_TARGET_LANG` environment variable
3. **System**: Operating system locale
4. **Default**: `en` (lowest priority)

---

## 🎯 Design Principles

Mint follows the Unix philosophy — **do one thing, and do it well.**

| Principle | Description |
|-----------|-------------|
| Zero-dependency install | Single binary, works out of the box |
| Multi-provider | Supports major LLM services plus local alternatives |
| Composability | Pairs seamlessly with `grep`, `sed`, `xargs`, and friends |
| Transparent output | Results go to stdout, errors go to stderr |
| Environment-friendly | API keys managed via environment variables, no config file pollution |

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

Apache License 2.0 — see [LICENSE](./LICENSE) file for details.
