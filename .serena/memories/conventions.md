# Conventions

## Packages & structure
- Package names: lowercase, short (e.g. `llm`, `googlegenai`, `openai`, `anthropic`)
- All new packages go under `internal/` or `cmd/mint/`
- LLM backends implement `internal/llm.Completer` (single method `Complete(ctx, prompt, w) (Usage, error)`)
- Adding a provider: new package `internal/provider/<name>/` exposing `New(apiKey, baseURL, modelName)`, plus a provider-name const + dispatch case in `NewCompleter` (internal/provider/provider.go)

## Config / env vars
- Single-key model: `MINT_PROVIDER` selects backend, `MINT_API_KEY` holds the key — NOT per-backend keys like `MINT_GEMINI_API_KEY`
- Others: `MINT_BASE_URL`, `MINT_MODEL_NAME`, `MINT_TARGET_LANG`, `MINT_VERBOSE` (full table in AGENTS.md)
- Viper instance (not global singleton) in command wiring; `SetEnvPrefix("MINT")` + `AutomaticEnv()`
- Validation centralized in `Config.ValidateConfig()` (internal/provider/config.go)

## CLI / errors
- Cobra root built in `newRootCmd()` factory, not `init()`
- Flags: `--target`/`-t`, `--source`/`-s`, `--verbose`/`-v`
- Errors returned up the stack; cobra prints to stderr via `RunE`
- No `log` package — use `fmt.Fprintf(os.Stderr, ...)` for diagnostics
