# Makefile for wallbit-go.
#
# Assumes GNU make and a POSIX shell. On Windows, run from Git Bash or WSL.
# Run `make` or `make help` for the list of available targets.

GO                     ?= go
GOLANGCI_LINT_VERSION  ?= v2.11.4
GOVULNCHECK_VERSION    ?= latest
COVERFILE              ?= coverage.out

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  test            Run unit tests (fast, no race detector)"
	@echo "  test-race       Run unit tests with the race detector (matches CI)"
	@echo "  vet             Run go vet"
	@echo "  fmt             Format code with gofmt"
	@echo "  lint            Run golangci-lint"
	@echo "  tidy            Run go mod tidy and verify"
	@echo "  vuln            Run govulncheck against all packages"
	@echo "  cover           Generate HTML coverage report (coverage.html)"
	@echo "  fuzz            Run fuzz targets briefly (10s each)"
	@echo "  check           Full pre-PR check (vet + lint + test-race + vuln)"
	@echo "  install-tools   Install pinned golangci-lint and govulncheck"
	@echo "  clean           Remove generated artifacts"
	@echo ""

.PHONY: test
test: ## Run unit tests (fast, no race detector)
	$(GO) test -count=1 ./...

.PHONY: test-race
test-race: ## Run unit tests with the race detector (matches CI)
	$(GO) test -race -count=1 ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: fmt
fmt: ## Format code with gofmt
	$(GO) fmt ./...

.PHONY: lint
lint: ## Run golangci-lint (requires `make install-tools` first)
	golangci-lint run ./...

.PHONY: tidy
tidy: ## Run go mod tidy and verify
	$(GO) mod tidy
	$(GO) mod verify

.PHONY: vuln
vuln: ## Run govulncheck against all packages
	$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

.PHONY: cover
cover: ## Generate HTML coverage report (coverage.html)
	$(GO) test -race -coverprofile=$(COVERFILE) ./...
	$(GO) tool cover -html=$(COVERFILE) -o coverage.html
	@echo "Coverage report written to coverage.html"

.PHONY: fuzz
fuzz: ## Run fuzz targets briefly (10s each)
	$(GO) test -run=^$$ -fuzz=FuzzErrorFromHTTP -fuzztime=10s ./wallbit/...

.PHONY: check
check: vet lint test-race vuln ## Full pre-PR check (mirrors CI)

.PHONY: install-tools
install-tools: ## Install pinned versions of golangci-lint and govulncheck
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

.PHONY: clean
clean: ## Remove generated artifacts
	@rm -f $(COVERFILE) coverage.html
