# Mint — Core (Serena memory-graph root)

Minimalist LLM-powered translation CLI in Go. Module `github.com/min0625/mint`.

**Canonical project facts live in `AGENTS.md` at the repo root**, which is always loaded
into context (CLAUDE.md does `@AGENTS.md`). That file owns: build/test/lint commands, tech
stack + tool pins, the `MINT_*` env-var table, conventions, key design decisions (language
detection / source flag / rotation / neutral pass-through), and the multilingual README-sync
rules. Do **not** duplicate AGENTS.md here — these memories only add Serena-specific
navigation and code-level detail that AGENTS.md does not spell out.

## Source map (symbol-level)

```
cmd/mint/main.go                # entry; cobra root via newRootCmd() factory; viper wiring; target-lang resolution
internal/llm/llm.go             # Completer interface: Complete(ctx, system, user string, w io.Writer) (Usage, error); Usage = token counts
internal/llm/writer.go          # TrailingNewlineWriter: wraps io.Writer, guarantees exactly one trailing '\n'; providers stream tokens through it + call Done()
internal/httpx/httpx.go         # New() *http.Client — shared tuned transport (proxy-from-env, HTTP/2, conn pool); every provider uses it instead of http.DefaultClient
internal/provider/config.go     # Config struct; Config.ValidateConfig(); provider-name constants
internal/provider/provider.go   # NewCompleter(cfg) factory — dispatches on cfg.Provider
internal/provider/googlegenai/  # Google Gemini HTTP client (implements Completer)
internal/provider/openai/       # OpenAI GPT HTTP client (implements Completer)
internal/provider/anthropic/    # Anthropic Claude HTTP client (implements Completer)
```

Build injects `-X main.version=<tag-or-sha> -X main.commit=<sha>` via ldflags (Makefile).

## Memory graph

- Adding a provider + non-obvious code-level conventions not in AGENTS.md: `mem:conventions`
- How these memories are structured/maintained (meta): `mem:memory_maintenance`
