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

# golangci-lint only *warns* and exits 0 when --new-from-rev does not resolve,
# so a typo'd or unfetched rev would silently lint nothing and report success.
# Both --new-from-rev callers below depend on this guard.
.PHONY: check-rev
check-rev:
	@git rev-parse --verify --quiet '$(NEW_FROM_REV)^{commit}' >/dev/null || \
		{ echo "NEW_FROM_REV=$(NEW_FROM_REV) is not a valid revision" >&2; exit 1; }

.PHONY: fix
fix: check-rev
	go mod tidy
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) --fix ./...

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

# The full gate: every hook in .pre-commit-config.yaml -- check-tidy, lint and
# test included -- over every tracked file, i.e. the same suite `git commit`
# runs, but repo-wide instead of staged. NEW_FROM_REV is set on the command
# line, so make exports it and prek passes it through to the nested make.
.PHONY: check
check:
	prek run --all-files --show-diff-on-failure

.PHONY: release
release:
	goreleaser release --clean

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean
