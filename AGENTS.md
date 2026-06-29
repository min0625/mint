# Mint — AI Agent Instructions

Mint is a minimalist LLM-powered translation CLI written in Go.
See [README.md](./README.md) for full project overview.

## Installation

End-user install methods (pipx, npm, one-liner script, manual download) are documented in
[README.md](./README.md). The install script lives at [script/install.sh](./script/install.sh).

For local development, build from source with `make build` (see below) or:

```bash
go install github.com/min0625/mint/cmd/mint@latest   # Go 1.26.4+; binary → $GOPATH/bin
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

## Distribution

Mint is published to PyPI as `mint-ai` and to npm as `mint-ai` (the command stays `mint`).
Wheels wrap the goreleaser-built binaries; the assembly script and per-platform packaging
live under [pypi/](./pypi/) (`build_wheels.py`) and [npm/](./npm/). For local wheel-build/test
steps see [pypi/](./pypi/).

**Release trigger:** push a tag matching `v*.*.*` → `release.yml` builds multi-platform
binaries → `publish-pypi.yml` assembles wheels and uploads to PyPI.

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
| `MINT_VERBOSE` | Set to `true` to enable verbose diagnostic output to stderr (equivalent to `--verbose`) | Optional | `false` |

**Conditional:* `MINT_API_KEY` required when using the default endpoint; optional when `MINT_BASE_URL` is set (proxy handles auth).*
**Default models:* `google-genai`: `gemini-3.1-flash-lite`, `openai`: `gpt-4o-mini`, `anthropic`: `claude-haiku-4-5`.*

## Documentation

- **Multilingual README sync** — **all** README variants must be kept in sync: `README.md` (English, the canonical source) and every `README.<locale>.md` translation. [LANGUAGES.md](./LANGUAGES.md) is the authoritative list of variants — consult it first to confirm the full set, since new languages are added over time. Whenever one README is updated, apply the equivalent change to *every* other variant listed there. New README language variants follow the pattern `README.<locale>.md`.
- **Language list** — the list of available languages lives **only** in [LANGUAGES.md](./LANGUAGES.md). Each README links to it with a single static line, written **in that README's own language** (e.g. English: `🌐 Other languages`; Traditional Chinese: `🌐 其他語言`). This line never changes when languages are added. To add a language: create `README.<locale>.md` and add one entry to `LANGUAGES.md` — do **not** add a per-language switcher row to every README.
- **Absolute URLs in README headers** — `README.md` is shipped as the PyPI long-description and the npm package readme. PyPI does **not** rewrite relative links, so the `LANGUAGES.md` link in each README must be an **absolute** GitHub URL (`https://github.com/min0625/mint/blob/main/LANGUAGES.md`). Links *inside* `LANGUAGES.md` may stay relative — that file is GitHub-only (not packaged into PyPI/npm).

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

- CLI framework: `github.com/spf13/cobra` — root command with `--target` / `-t` (target language), `--source` / `-s` (source language), and `--verbose` / `-v` (diagnostic output to stderr) flags.
- Configuration: `github.com/spf13/viper` — reads env vars with `MINT_` prefix; no config files.
- LLM backends called directly via raw `net/http` (no heavy SDKs); keeps binary minimal.
- `Completer` interface in `internal/llm/` allows provider backends without breaking changes. `Complete(ctx, system, user, w)` keeps task instructions (`system`) separate from untrusted user input (`user`) so translated content cannot contaminate the instruction context. Each request also embeds a random nonce as a delimiter in the system prompt; the same nonce wraps the user text, preventing injected content from escaping its boundary even if it mimics delimiter syntax. Weaker models that echo the nonce lines back are filtered before the output reaches the caller.
- Language detection: when no source language is given, the input language is inferred — by the LLM in rotation mode, or implicitly by the rewrite prompt for a single target.
- Source language: optional `--source` / `-s` flag (BCP-47 tag); flag-only, no env var (a source is per-input, not a persistent preference). When set it skips detection and anchors the rewrite prompt to translate *from* that language, so cross-language homographs (e.g. French `chat` → English `cat`) and romanized input (e.g. `konnichiwa` → `hello`) are translated rather than treated as already-target text. Empty (the default) preserves the original auto-detect behavior. Pure language-neutral input still passes through unchanged regardless.
- Language-neutral pass-through: if detected language is `neutral`, input is printed unchanged with no translation call.
- Same-language behavior: if detected input language matches the target language, the tool performs grammar and spelling correction instead of translation.
- Target language priority: `--target` flag → `MINT_TARGET_LANG` env var → system locale (`$LC_ALL` / `$LANG`) → `en`.
- Language rotation: `MINT_TARGET_LANG` accepts a comma-separated list (e.g., `en,zh-TW,ja`); when the detected input language matches a tag in the list, the tool translates to the next tag (wraps around). BCP-47 variants sharing the same primary subtag (e.g., `zh-HK` and `zh-TW`) are treated as equivalent.
- Input from positional arg or stdin (auto-detected).
