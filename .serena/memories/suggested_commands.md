# Suggested Commands

```bash
make build        # compile → bin/mint
make test         # go test -race -failfast -v ./...
make lint         # golangci-lint --new-from-rev=HEAD
make fix          # golangci-lint --fix + go mod tidy
make check        # check-tidy + lint + test (CI gate)
make check-tidy   # verify go.mod/go.sum are tidy
mise install      # set up pinned Go + golangci-lint versions
```
