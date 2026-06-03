**English** | [繁體中文](./README.zh-TW.md)

# 🌿 Mint

> Minimalist AI Translation CLI — Simple. Fast. Intuitive.

---

Mint is a lightweight, LLM-powered translation tool for the command line.
No complex setup, no cluttered interface — just one command, and you get fluent, natural translations instantly.

---

## ✨ Why Mint?

Most translation tools are either too bloated or too locked into a specific platform.
Mint is built around a single philosophy: **do less, do it well.**

- **Minimal** — One command, no noise
- **Fast** — Calls the LLM API directly, no intermediate layers
- **Flexible** — Supports any language pair, freely specify your target language
- **Composable** — Pipe-friendly stdin/stdout design, fits naturally into any workflow

---

## 🚀 Quick Start

### 1. Set your API key

```bash
export MINT_GEMINI_API_KEY=your_api_key_here
```

Get a free API key at [Google AI Studio](https://aistudio.google.com/apikey).

### 2. Translate

```bash
# Specify a target language (BCP-47 tag)
mint --to ja "Good morning"

# Pipe from stdin
echo "The quick brown fox" | mint --to zh-TW

# Translate a file
cat document.txt | mint --to fr
```

---

## 🔑 Environment Variables

| Variable | Description |
|----------|-------------|
| `MINT_GEMINI_API_KEY` | Google Gemini API key **(required)** |

---

## 🎯 Design Principles

Mint follows the Unix philosophy — **do one thing, and do it well.**

| Principle | Description |
|-----------|-------------|
| Zero-dependency install | Single binary, works out of the box |
| Composability | Pairs seamlessly with `grep`, `sed`, `xargs`, and friends |
| Transparent output | Results go to stdout, errors go to stderr |
| Environment-friendly | API keys managed via environment variables, no config file pollution |

---

## 🗺 Roadmap

- [x] Gemini LLM backend
- [ ] Additional LLM backends (OpenAI, Anthropic, Ollama local models)
- [ ] Automatic source language detection
- [ ] Batch translation mode
- [ ] Glossary / custom dictionary support
- [ ] Output format options (plain text, JSON, Markdown)

---

## 📄 License

Apache License 2.0 — see [LICENSE](./LICENSE) file for details.
