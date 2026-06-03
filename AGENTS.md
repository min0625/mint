# Mint — AI Agent Instructions

Mint is a minimalist LLM-powered translation CLI written in Go.
See [README.md](./README.md) for full project overview.

## Build & Test Commands

```bash
make build        # compile → bin/mint (CGO_ENABLED=0, trimpath, ldflags injected)
make test         # go test -race -failfast -v ./...
make lint         # golangci-lint (only new violations since HEAD)
make fix          # golangci-lint --fix + go mod tidy
make check        # check-tidy + lint + test (CI gate)
make check-tidy   # verify go.mod/go.sum are tidy
```

Tool versions are pinned in [mise.toml](./mise.toml) (Go 1.26.4, golangci-lint 2.12.2).
Run `mise install` to set up the exact toolchain.

## Project Layout

```
cmd/mint/main.go              # entry point; cobra root command; viper config wiring
internal/translator/          # Translator interface (translator.Translator)
internal/gemini/              # Gemini HTTP client (implements Translator)
bin/mint                      # compiled binary (gitignored)
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `MINT_GEMINI_API_KEY` | Google Gemini API key (required for translation) |

## Conventions

- **CGO disabled** — keep the binary fully static (`CGO_ENABLED=0` in Makefile).
- **Single binary** — no config files; API keys come from environment variables only.
- **Pipe-friendly** — translation input via args or stdin; results to stdout; errors to stderr.
- **Unix philosophy** — do one thing well; composable with `grep`, `sed`, `xargs`, etc.
- **No unnecessary dependencies** — keep `go.mod` minimal.
- Lint is checked only for *new* violations (`--new-from-rev=HEAD`); always run `make lint` before committing.

## Key Design Decisions

- CLI framework: `github.com/spf13/cobra` — root command with `--to <lang>` flag (required).
- Configuration: `github.com/spf13/viper` — reads env vars with `MINT_` prefix; no config files.
- LLM backend called directly via raw `net/http` (no SDK); keeps binary minimal.
- `Translator` interface in `internal/translator/` allows future backends without breaking changes.
- Target language specified via `--to <lang>` flag (BCP-47 tags, e.g. `zh-TW`, `ja`, `fr`).
- Input from positional arg or stdin (auto-detected).
- Planned additional backends: OpenAI (`MINT_OPENAI_API_KEY`), Anthropic (`MINT_ANTHROPIC_API_KEY`), Ollama (see roadmap in README).
