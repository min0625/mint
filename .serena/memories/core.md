# Mint — Core

Minimalist LLM-powered translation CLI written in Go.
Module: `github.com/min0625/mint`

## Source Map

```
cmd/mint/main.go                # entry; cobra root cmd; viper wiring; target-language resolution
internal/llm/llm.go             # Completer interface — Complete(ctx, prompt, w) (Usage, error); Usage struct
internal/provider/config.go     # Config struct; ValidateConfig; provider-name constants
internal/provider/provider.go   # NewCompleter(cfg) factory — dispatches on cfg.Provider
internal/provider/googlegenai/  # Google Gemini HTTP client (implements Completer)
internal/provider/openai/       # OpenAI GPT HTTP client (implements Completer)
internal/provider/anthropic/    # Anthropic Claude HTTP client (implements Completer)
bin/mint                        # compiled binary (gitignored)
```

## Project-Wide Invariants

- CGO disabled (`CGO_ENABLED=0`), single static binary
- No config files — config from `MINT_`-prefixed env vars (+ CLI flags) only
- Pipe-friendly: input from args OR stdin; output to stdout; errors to stderr
- `version` and `commit` injected via ldflags at build time (see `mem:tech_stack`)
- Keep `go.mod` minimal — add dependencies only when necessary
- README sync: every `README.<locale>.md` mirrors canonical `README.md`; language list lives only in LANGUAGES.md (see AGENTS.md "Documentation")

## Key References

- Build/test/lint commands: `mem:suggested_commands`
- Language/toolchain pins: `mem:tech_stack`
- Code conventions + provider/env design: `mem:conventions`
- Task completion checklist: `mem:task_completion`
