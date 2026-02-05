# Makefile for the Tournabyte webapi

CMD_DIR  ?= $(shell pwd)/app/start
BUILD_DIR ?= $(shell pwd)/bin
APP_NAME ?= tbyte-webapi

GO  ?= go


GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_TAG    := $(shell git describe --tags --dirty --always 2>/dev/null)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")


.PHONY: help build test test-race test-cov vet fmt deps clean

help: ## Display this message and exit
	@echo "Makefile for the Tournabyte API webserver. Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the go application
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

test: ## Run the unit test suite
	$(GO) test ./...

test-race: ## Run the unit test suite with race condition detection
	$(GO) test -race ./...

test-cov: ## Run the unit test suite with code coverage analysis
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

fmt: ## Format go modules
	$(GO) fmt ./...

vet: ## Vet go modules
	$(GO) vet ./...

deps: ## Manage go dependencies
	$(GO) mod tidy
	$(GO) mod download

clean: ## Clean up build artifacts
	rm -rf $(BUILD_DIR) coverage.out
