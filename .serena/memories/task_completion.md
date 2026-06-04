# Task Completion

Run these before committing any coding task:

1. `make build` — must compile cleanly
2. `make lint` — no new violations (checks only since HEAD)
3. `make test` — all tests pass
4. `make check-tidy` — go.mod/go.sum must be tidy

Quick combined gate: `make check`
