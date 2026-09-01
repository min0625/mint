# Leading "v" is stripped to match goreleaser's {{.Version}}, so `mint --version`
# prints the same string for local builds and released binaries.
VERSION ?= $(shell (git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD) | sed 's/^v//')
COMMIT ?= $(shell git rev-parse HEAD)
LDFLAGS ?= -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)
NEW_FROM_REV ?= HEAD

.PHONY: build
build:
	mkdir -p ./bin/
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o ./bin/ ./cmd/mint

# An unresolvable rev is not silent: golangci-lint warns, reports every issue, and
# exits 1. This guard is for the message, not the exit code.
.PHONY: check-rev
check-rev:
	@git rev-parse --verify --quiet '$(NEW_FROM_REV)^{commit}' >/dev/null || \
		{ echo "NEW_FROM_REV=$(NEW_FROM_REV) is not a valid revision" >&2; exit 1; }

# Tidy first: a stale go.mod fails typecheck, which is reported past --new-from-rev
# while every other linter falls silent -- so --fix would quietly repair nothing.
# Tidy again after: exptostd rewrites x/exp imports to stdlib and strips a require.
# The closing lint reports whatever --fix could not repair.
.PHONY: fix
fix: check-rev
	go mod tidy
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) --fix ./...
	go mod tidy
	@$(MAKE) --no-print-directory lint

.PHONY: lint
lint: check-rev
	golangci-lint config verify
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) ./...

.PHONY: test
test:
	go test -race -failfast ./...

.PHONY: cover
cover:
	go test -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: cover-html
cover-html: cover
	go tool cover -html=coverage.out -o coverage.html

.PHONY: check-tidy
check-tidy:
	go mod tidy -diff

# The full gate: every hook in .pre-commit-config.yaml -- check-tidy, lint and test
# included -- over every tracked file, i.e. what `git commit` runs but repo-wide.
# NEW_FROM_REV on the command line reaches the nested `make lint` via MAKEFLAGS;
# without it lint falls back to HEAD and reports nothing on CI's clean checkout.
.PHONY: check
check:
	prek run --all-files --show-diff-on-failure

.PHONY: release
release:
	goreleaser release --clean

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean
