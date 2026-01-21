BINARY := pingheat
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-s -w -X github.com/pbv7/pingheat/pkg/version.Version=$(VERSION) -X github.com/pbv7/pingheat/pkg/version.Commit=$(COMMIT) -X github.com/pbv7/pingheat/pkg/version.BuildTime=$(BUILD_TIME)"

.PHONY: all build clean clean-dist clean-all test test-cover cover-summary lint lint-md lint-workflows lint-all run install release release-snapshot release-check

all: build

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/pingheat

run: build
	./bin/$(BINARY) $(ARGS)

install:
	go install $(LDFLAGS) ./cmd/pingheat

test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "\nCoverage report generated: coverage.html"
	@echo "Open with: open coverage.html"

cover-summary:
	go test ./... -coverprofile=coverage.out
	@echo "\n=== Coverage Summary ==="
	@go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run ./...

lint-md:
	npx --yes markdownlint-cli2 "*.md" "**/*.md" "!node_modules" "!dist"

lint-workflows:
	@command -v actionlint >/dev/null 2>&1 || { echo "actionlint not found. Install with: brew install actionlint"; exit 1; }
	actionlint .github/workflows/*.yml

lint-all: lint lint-md lint-workflows

clean:
	rm -rf bin/ coverage.out coverage.html

clean-dist:
	rm -rf dist/

clean-all: clean clean-dist

deps:
	go mod download
	go mod tidy

# Cross-compilation targets
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/pingheat

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 ./cmd/pingheat
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/pingheat

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe ./cmd/pingheat

build-all: build-linux build-darwin build-windows

# GoReleaser targets
release:
	goreleaser release --clean

release-snapshot:
	goreleaser release --snapshot --clean

release-check:
	goreleaser check
