# gen-cobra-flags Makefile
#
# This repository is composed of three Go modules:
#   .         the generator + CLI  (github.com/gocloud9/gen-cobra-flags)
#   ./sdk     the runtime library imported by generated code
#   ./example a read-only fixture exercising the generator
#
# golangci-lint v2 (built with the same Go toolchain as the target) is required;
# older v1 releases cannot analyze a go1.26 module. Run `make tools` to install
# a compatible version into $(go env GOPATH)/bin.

GO            ?= go
GOBIN         := $(shell $(GO) env GOPATH)/bin
GOLANGCI_LINT := $(GOBIN)/golangci-lint
GOLANGCI_VERSION ?= v2.12.2
CONFIG        := $(CURDIR)/.golangci.yml

MODULES       := . sdk example
CMD           := ./cmd/gen-cobra-flags
BINARY        := bin/gen-cobra-flags

# Parameters used to regenerate the example fixture's output.
EXAMPLE_DIR    := $(CURDIR)/example
EXAMPLE_OUT    := $(EXAMPLE_DIR)/generated
EXAMPLE_PKG    := generated
EXAMPLE_IMPORT := github.com/gocloud9/gen-cobra-flags/example

.DEFAULT_GOAL := build

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the gen-cobra-flags CLI binary into ./bin.
	$(GO) build -o $(BINARY) $(CMD)

.PHONY: install
install: ## Install the CLI into $(GOBIN).
	$(GO) install $(CMD)

.PHONY: test
test: ## Run tests across all modules.
	@set -e; for m in $(MODULES); do \
		echo "==> test $$m"; \
		(cd $$m && $(GO) test ./...); \
	done

.PHONY: vet
vet: ## Run go vet across all modules.
	@set -e; for m in $(MODULES); do \
		echo "==> vet $$m"; \
		(cd $$m && $(GO) vet ./...); \
	done

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint all modules with golangci-lint v2.
	@set -e; for m in $(MODULES); do \
		echo "==> lint $$m"; \
		(cd $$m && $(GOLANGCI_LINT) run --config $(CONFIG) ./...); \
	done

.PHONY: fmt
fmt: $(GOLANGCI_LINT) ## Format code (gofmt + import grouping) across all modules.
	@set -e; for m in $(MODULES); do \
		echo "==> fmt $$m"; \
		$(GO) -C $$m fmt ./...; \
		(cd $$m && $(GOLANGCI_LINT) run --config $(CONFIG) --fix ./... || true); \
	done

.PHONY: generate
generate: build ## Regenerate the example fixture's output with the freshly built CLI.
	$(BINARY) \
		-input $(EXAMPLE_DIR) \
		-output $(EXAMPLE_OUT) \
		-package $(EXAMPLE_PKG) \
		-source-import $(EXAMPLE_IMPORT)

.PHONY: tidy
tidy: ## Run go mod tidy across all modules.
	@set -e; for m in $(MODULES); do \
		echo "==> tidy $$m"; \
		(cd $$m && $(GO) mod tidy); \
	done

.PHONY: check
check: vet lint test ## Run vet, lint, and tests across all modules.

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf bin

.PHONY: tools
tools: ## Install pinned developer tools (golangci-lint v2) into $(GOBIN).
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

$(GOLANGCI_LINT):
	@echo "golangci-lint not found at $(GOLANGCI_LINT); installing $(GOLANGCI_VERSION)"
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
