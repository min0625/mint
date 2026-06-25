# Conventions (additive to AGENTS.md)

Only code-level conventions that `AGENTS.md` does not already state. Env-var table, CLI
flags, design decisions, and build/test commands live in AGENTS.md — not here.

## Adding an LLM provider
1. New package `internal/provider/<name>/` exposing `New(apiKey, baseURL, modelName)` that
   returns an `llm.Completer` (single method `Complete(ctx, system, user, w) (Usage, error)` —
   system/user are split messages for prompt-injection isolation).
2. Add a provider-name const + a dispatch case in `NewCompleter` (internal/provider/provider.go).
- Single-key config model: `MINT_PROVIDER` selects the backend, `MINT_API_KEY` holds the key
  — NOT per-backend keys like `MINT_GEMINI_API_KEY`.

## Code-level conventions
- Packages: lowercase, short (`llm`, `googlegenai`, `openai`, `anthropic`); all new packages
  under `internal/` or `cmd/mint/`.
- Viper: use a local viper instance (not the global singleton) in command wiring;
  `SetEnvPrefix("MINT")` + `AutomaticEnv()`.
- Cobra root built in a `newRootCmd()` factory, not `init()`.
- Config validation centralized in `Config.ValidateConfig()` (internal/provider/config.go).
- No `log` package — write diagnostics with `fmt.Fprintf(os.Stderr, ...)`; errors are returned
  up the stack and cobra prints them via `RunE`.
