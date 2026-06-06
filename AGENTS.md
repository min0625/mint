# Mint — AI Agent Instructions

Mint is a minimalist LLM-powered translation CLI written in Go.
See [README.md](./README.md) for full project overview.

## Installation

### pipx (Recommended)

```bash
pipx install mint-ai
```

### npm

```bash
npm install -g mint-ai
```

### Quick Install (Binary)

```bash
curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh | bash
```

### go install

If you have Go 1.21+ installed:

```bash
go install github.com/min0625/mint/cmd/mint@latest
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
mint --to en "Hello, world!"
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
internal/translator/                # Translator interface (translator.Translator)
internal/provider/
  config.go                          # Config struct; provider validation
  provider.go                        # NewTranslator factory function
  googlegenai/google_genai.go        # Google Gemini HTTP client (implements Translator)
  openai/openai.go                   # OpenAI GPT HTTP client (implements Translator)
  anthropic/anthropic.go             # Anthropic Claude HTTP client (implements Translator)
  ollama/ollama.go                   # Ollama local LLM HTTP client (implements Translator)
bin/mint                             # compiled binary (gitignored)
.goreleaser.yaml                     # GoReleaser multi-platform release configuration
.github/workflows/release.yml        # GitHub Actions: triggered on v*.*.* tag push
```

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `MINT_PROVIDER` | LLM provider: `google-genai`, `openai`, `anthropic`, `ollama` | Yes | — |
| `MINT_API_KEY` | API key for the chosen provider | Conditional* | — |
| `MINT_BASE_URL` | Custom API endpoint; required for `ollama` (e.g., for self-hosted or local services) | Conditional* | Provider default |
| `MINT_MODEL_NAME` | LLM model name to use | Optional | Provider default** |
| `MINT_TARGET_LANG` | Target language(s) - single or comma-separated (e.g. `en`, `en,zh-TW,ja`) | Optional | System locale or `en` |

**Conditional:* `MINT_API_KEY` required for `google-genai`, `openai`, `anthropic`; not needed for `ollama`. `MINT_BASE_URL` required for `ollama`.*
**Default models:* `google-genai`: `gemini-3.1-flash-lite`, `openai`: `gpt-4o-mini`, `anthropic`: `claude-haiku-4-5`; `ollama`: none (must specify).*

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

- CLI framework: `github.com/spf13/cobra` — root command with optional `--target` / `-t` flag.
- Configuration: `github.com/spf13/viper` — reads env vars with `MINT_` prefix; no config files.
- LLM backends called directly via raw `net/http` (no heavy SDKs); keeps binary minimal.
- `Translator` interface in `internal/translator/` allows provider backends without breaking changes.
- Target language: use `--target` / `-t` explicitly, or set `MINT_TARGET_LANG` for smart auto-detection with multi-language rotation.
- Smart detection: if `MINT_TARGET_LANG` is set, automatically detects input language and translates to next target language in rotation.
- Input from positional arg or stdin (auto-detected).
- Language rotation: supports comma-separated language list in `MINT_TARGET_LANG` (e.g., `en,zh-TW,ja`); cycles through based on detected input language.
