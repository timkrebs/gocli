TEST      ?= ./...
TESTARGS  ?=
GOFLAGS   ?=

# Default target
.DEFAULT_GOAL := help

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ── Build ─────────────────────────────────────────────────────────────────────

.PHONY: build
build: ## Compile all packages (catches import and syntax errors)
	go build $(GOFLAGS) $(TEST)

# ── Test ──────────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run the test suite
	go test -timeout=60s -parallel=10 $(TESTARGS) $(TEST)

.PHONY: testrace
testrace: ## Run the test suite with the race detector enabled
	go test -race -timeout=120s $(TESTARGS) $(TEST)

.PHONY: coverage
coverage: ## Generate an HTML coverage report (opens coverage.html)
	go test -coverprofile=coverage.out -covermode=atomic $(TEST)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

# ── Lint & Vet ────────────────────────────────────────────────────────────────

.PHONY: vet
vet: ## Run go vet
	go vet $(TEST)

.PHONY: lint
lint: ## Run golangci-lint (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m $(TEST)

.PHONY: lint-fix
lint-fix: ## Run golangci-lint and auto-fix issues where possible
	golangci-lint run --fix --timeout=5m $(TEST)

# ── Security ──────────────────────────────────────────────────────────────────

.PHONY: vuln
vuln: ## Run govulncheck (install: go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck $(TEST)

# ── Dependencies ──────────────────────────────────────────────────────────────

.PHONY: deps
deps: ## Download module dependencies
	go mod download

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum
	go mod tidy

.PHONY: verify
verify: ## Verify module dependencies against go.sum
	go mod verify

# ── Release ───────────────────────────────────────────────────────────────────

.PHONY: release-check
release-check: ## Validate the .goreleaser.yml configuration
	goreleaser check

.PHONY: release-snapshot
release-snapshot: ## Build a local snapshot release (no publish, no tag required)
	goreleaser release --snapshot --clean

# ── Clean ─────────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove generated artefacts (coverage files, goreleaser dist)
	rm -f coverage.out coverage.html
	rm -rf dist/

# ── CI shorthand ─────────────────────────────────────────────────────────────

.PHONY: ci
ci: verify build vet test testrace ## Run the full CI check suite locally
