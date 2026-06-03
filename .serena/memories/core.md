# Mint — Core

Minimalist LLM-powered translation CLI written in Go.
Module: `github.com/min0625/mint`

## Source Map

```
cmd/mint/main.go              # entry point; cobra root command; viper config wiring
internal/translator/          # Translator interface
internal/gemini/              # Gemini HTTP client (implements Translator)
bin/mint                      # compiled binary (gitignored)
```

## Project-Wide Invariants

- CGO disabled (`CGO_ENABLED=0`), single static binary
- No config files — API keys from env vars only
- Pipe-friendly: input from args OR stdin; output to stdout; errors to stderr
- `version` and `commit` injected via ldflags at build time (see `mem:tech_stack`)
- Keep `go.mod` minimal — add dependencies only when necessary

## Key References

- Build/test/lint commands: `mem:suggested_commands`
- Language/toolchain pins: `mem:tech_stack`
- Code conventions: `mem:conventions`
- Task completion checklist: `mem:task_completion`
