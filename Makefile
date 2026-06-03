VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)
COMMIT ?= $(shell git rev-parse --short HEAD)
LDFLAGS ?= -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)
NEW_FROM_REV ?= HEAD

.PHONY: build
build:
	mkdir -p ./bin/
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o ./bin/ ./cmd/mint

.PHONY: fix
fix:
	go mod tidy
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) -v --fix ./...

.PHONY: lint-verify
lint-verify:
	golangci-lint config verify

.PHONY: lint
lint: lint-verify
	golangci-lint run --new-from-rev=$(NEW_FROM_REV) -v ./...

.PHONY: test
test:
	go test -race -failfast -v ./...

.PHONY: check-tidy
check-tidy:
	go mod tidy -diff

.PHONY: check
check: check-tidy lint test
