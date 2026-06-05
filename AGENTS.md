# Mint — AI Agent Instructions

Mint is a minimalist LLM-powered translation CLI written in Go.
See [README.md](./README.md) for full project overview.

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/min0625/mint/main/script/install.sh | bash
```

Or with `go install` (requires Go 1.21+):

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
| `MINT_PROVIDER` | LLM provider (`google-genai`, `openai`, `anthropic`, `ollama`) | Yes | — |
| `MINT_API_KEY` | API key for the provider (not required for `ollama`) | Conditional* | — |
| `MINT_BASE_URL` | Custom API endpoint URL | Optional | provider default |
| `MINT_MODEL_NAME` | Model name to use | Optional | provider default** |
| `MINT_TARGET_LANG` | Target language(s) - single or comma-separated (e.g. `en`, `en,zh-TW`) | Optional | System locale or `en` |

**Conditional:* Required for `google-genai`, `openai`, `anthropic`; not required for `ollama`.*
**Default models:* `google-genai`: `gemini-3.1-flash-lite`; `openai`: `gpt-4o-mini`; `anthropic`: `claude-haiku-4-5`; `ollama`: none (must be specified).

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

- CLI framework: `github.com/spf13/cobra` — root command with optional `--to <lang>` flag.
- Configuration: `github.com/spf13/viper` — reads env vars with `MINT_` prefix; no config files.
- LLM backends called directly via raw `net/http` (no heavy SDKs); keeps binary minimal.
- `Translator` interface in `internal/translator/` allows provider backends without breaking changes.
- Target language: use `--to <lang>` explicitly, or set `MINT_PRIMARY_LANGUAGE` for smart auto-detection.
- Smart detection: if `MINT_PRIMARY_LANGUAGE` is set and text is in that language, automatically translate to `MINT_SECONDARY_LANGUAGE`.
- Input from positional arg or stdin (auto-detected).
- Planned additional backends: OpenAI (`MINT_OPENAI_API_KEY`), Anthropic (`MINT_ANTHROPIC_API_KEY`), Ollama (see roadmap in README).
