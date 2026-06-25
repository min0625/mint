# Tech Stack

- **Language**: Go 1.26.4 (pinned via `mise.toml`)
- **Linter**: golangci-lint 2.12.2 (pinned via `mise.toml`)
- **Release**: goreleaser 2.16.0 (pinned via `mise.toml`)
- **Toolchain manager**: mise (`mise install` to set up exact versions)
- **CLI framework**: `github.com/spf13/cobra`
- **Config/env**: `github.com/spf13/viper`
- **HTTP**: raw `net/http` per provider — no vendor SDKs (keeps binary minimal)
- **Build**: `make build` → `CGO_ENABLED=0 go build -trimpath -ldflags=...`
- **ldflags**: `-X main.version=<tag-or-sha> -X main.commit=<sha>`
- No test framework beyond stdlib `testing`
