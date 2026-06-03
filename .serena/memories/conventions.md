# Conventions

- Package names: lowercase, short (e.g. `gemini`, `translator`)
- All new packages go under `internal/` or `cmd/mint/`
- Cobra command constructed in a `newRootCmd()` factory, not `init()`
- Viper instance (not global singleton) created in command wiring; `SetEnvPrefix("MINT")` + `AutomaticEnv()`
- Env var pattern: `MINT_<BACKEND>_API_KEY` (e.g. `MINT_GEMINI_API_KEY`)
- Translator backends implement `internal/translator.Translator` interface
- Errors returned up the call stack; cobra prints them to stderr with `RunE`
- No `log` package usage — use `fmt.Fprintf(os.Stderr, ...)` for error output
