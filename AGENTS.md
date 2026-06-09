# Mint — AI Agent Instructions

Mint is a minimalist LLM-powered translation CLI written in Go.
See [README.md](./README.md) for full project overview.

## Installation

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

Auto-detects OS and architecture (Linux/macOS, x86_64/arm64), verifies SHA256 checksums,
and installs to `~/.local/bin`. Override with `MINT_INSTALL_DIR` or pin a version with `MINT_VERSION=v1.0.0`.

### go install

```bash
go install github.com/min0625/mint/cmd/mint@latest
```

Requires Go 1.26.4+. Binary lands in `$GOPATH/bin` (usually `~/go/bin`).

### Manual download

Pre-built binaries at [GitHub Releases](https://github.com/min0625/mint/releases):

```bash
# Linux x86_64
curl -L https://github.com/min0625/mint/releases/latest/download/mint_linux_amd64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# macOS Apple Silicon
curl -L https://github.com/min0625/mint/releases/latest/download/mint_darwin_arm64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# macOS Intel
curl -L https://github.com/min0625/mint/releases/latest/download/mint_darwin_amd64.tar.gz \
  | tar xz && sudo mv mint /usr/local/bin/

# Windows — download mint_windows_amd64.zip from releases and extract to a directory in PATH
```

## Build & Test Commands

```bash
make build        # compile → bin/mint (CGO_ENABLED=0, trimpath, ldflags injected)
make test         # go test -race -failfast -v ./...
make lint         # golangci-lint (only new violations since HEAD)
make fix          # golangci-lint --fix + go mod tidy
make check        # check-tidy + lint + test (CI gate)
make check-tidy   # verify go.mod/go.sum are tidy
make release-snapshot  # goreleaser release --snapshot --clean (test release locally)
```

Tool versions are pinned in [mise.toml](./mise.toml) (Go 1.26.4, golangci-lint 2.12.2, goreleaser 2.16.0).
Run `mise install` to set up the exact toolchain.

## PyPI Distribution (Python Wheels)

Mint is published to PyPI as `mint-ai` for easy installation via Python package managers:

```bash
pip install mint-ai
# or
pipx install mint-ai
```

After installation, the command is `mint` (not `mint-ai`), driven by the entry point in `python/pyproject.toml`.

### Local Testing of Wheels

```bash
# Step 1: Build platform-specific binaries (all platforms)
make release-snapshot

# Step 2: Build Python wheel(s)
cd python
# For current platform only (recommended for local testing to avoid platform mismatches)
python build_wheels.py \
  --version 0.1.0 \
  --dist-dir ../dist \
  --out-dir ../dist/wheels \
  --current-platform-only

# Step 3: Install and test locally
# Use --find-links so pip auto-selects the compatible wheel (avoids glob ambiguity)
pipx install mint-ai --pip-args="--find-links ../dist/wheels --no-index" --force

# Step 4: Verify
mint --target en "Hello, world!"
```

**PyPI Release Workflow:**
1. Push tag: `git tag v1.0.0 && git push origin v1.0.0`
2. GitHub Actions triggers `release.yml` → builds multi-platform binaries
3. GitHub Actions triggers `publish.yml` → assembles wheels → uploads to PyPI
4. Final users: `pip install mint-ai` (PyPI auto-selects correct platform)

### Directory Structure

```
python/
  mint/
    __init__.py              # Python wrapper; locates and execs bundled binary
    __main__.py              # Enables `python -m mint`
  pyproject.toml             # Package metadata; specifies entry point `mint = mint:main`
  build_wheels.py            # Assembles wheels from goreleaser dist/ binaries
.github/workflows/publish.yml # CI: download binaries → assemble wheels → upload to PyPI
.goreleaser.yaml             # Configure archive names for wheel building
```

## Project Layout

```
cmd/mint/main.go                     # entry point; cobra root command; viper config wiring
internal/llm/llm.go                  # Completer interface for LLM backends
internal/provider/
  config.go                          # Config struct; provider validation
  provider.go                        # NewCompleter factory function
  googlegenai/google_genai.go        # Google Gemini HTTP client (implements Completer)
  openai/openai.go                   # OpenAI GPT HTTP client (implements Completer)
  anthropic/anthropic.go             # Anthropic Claude HTTP client (implements Completer)
bin/mint                             # compiled binary (gitignored)
.goreleaser.yaml                     # GoReleaser multi-platform release configuration
.github/workflows/release.yml        # GitHub Actions: triggered on v*.*.* tag push
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `MINT_PROVIDER` | LLM provider: `google-genai`, `openai`, `anthropic` | Yes | — |
| `MINT_API_KEY` | API key for the chosen provider; optional when `MINT_BASE_URL` is set | Conditional* | — |
| `MINT_BASE_URL` | Custom API base URL (domain only; each provider appends its own path); use with `openai` to target local inference servers — Ollama (`http://localhost:11434`) or LM Studio (`http://localhost:1234`); optional for cloud providers to point to a proxy | Conditional* | Provider default |
| `MINT_MODEL_NAME` | LLM model name to use | Required when `MINT_BASE_URL` is set; optional otherwise | Provider default** |
| `MINT_TARGET_LANG` | Target language(s) - single or comma-separated (e.g. `en`, `en,zh-TW,ja`) | Optional | System locale or `en` |

**Conditional:* `MINT_API_KEY` required when using the default endpoint; optional when `MINT_BASE_URL` is set (proxy handles auth).*
**Default models:* `google-genai`: `gemini-3.1-flash-lite`, `openai`: `gpt-4o-mini`, `anthropic`: `claude-haiku-4-5`.*

## Documentation

- **Multilingual README sync** — `README.md` (English) and `README.zh-TW.md` (Traditional Chinese) must always be kept in sync. Whenever one is updated, apply the equivalent change to the other. New README language variants follow the pattern `README.<locale>.md`.

## Conventions

- **CGO disabled** — keep the binary fully static (`CGO_ENABLED=0` in Makefile).
- **Single binary** — no config files; API keys come from environment variables only.
- **Pipe-friendly** — translation input via args or stdin; results to stdout; errors to stderr.
- **Unix philosophy** — do one thing well; composable with `grep`, `sed`, `xargs`, etc.
- **No unnecessary dependencies** — keep `go.mod` minimal.
- Lint is checked only for *new* violations (`--new-from-rev=HEAD`); always run `make lint` before committing.
- **Release workflow** — push a tag matching `v*.*.*` to automatically trigger GoReleaser CI; creates GitHub Release with multi-platform binaries.
- **Local snapshot testing** — run `goreleaser release --snapshot --clean` to validate build configuration before publishing.

## Key Design Decisions

- CLI framework: `github.com/spf13/cobra` — root command with `--target` / `-t` (target language) and `--verbose` / `-v` (diagnostic output to stderr) flags.
- Configuration: `github.com/spf13/viper` — reads env vars with `MINT_` prefix; no config files.
- LLM backends called directly via raw `net/http` (no heavy SDKs); keeps binary minimal.
- `Completer` interface in `internal/llm/` allows provider backends without breaking changes.
- Language detection: always runs via LLM before translation; detects BCP-47 tag or `neutral` for language-agnostic content (numbers, symbols).
- Language-neutral pass-through: if detected language is `neutral`, input is printed unchanged with no translation call.
- Same-language behavior: if detected input language matches the target language, the tool performs grammar and spelling correction instead of translation.
- Target language priority: `--target` flag → `MINT_TARGET_LANG` env var → system locale (`$LANG` / `$LC_ALL`) → `en`.
- Language rotation: `MINT_TARGET_LANG` accepts a comma-separated list (e.g., `en,zh-TW,ja`); when the detected input language matches a tag in the list, the tool translates to the next tag (wraps around). BCP-47 variants sharing the same primary subtag (e.g., `zh-HK` and `zh-TW`) are treated as equivalent.
- Input from positional arg or stdin (auto-detected).
