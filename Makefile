# VIIPER Makefile
# Cross-platform build automation for VIIPER

# Variables
BINARY_NAME := viiper
MAIN_PKG := ./cmd/viiper
SRC_DIR := viiper
DIST_DIR := dist
VERSION ?= $(shell git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --always 2>nul || echo v0.0.0-dev)
COMMIT := $(shell git rev-parse --short HEAD 2>nul || echo unknown)
BUILD_TIME := $(shell powershell -NoProfile -NonInteractive -Command "Get-Date -Format 'yyyy-MM-dd_HH:mm:ss'")

# Go build flags
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

# Windows detection and environment setup
ifeq ($(OS),Windows_NT)
    EXE_EXT := .exe
    export CGO_ENABLED=0
else
    EXE_EXT :=
    export CGO_ENABLED=0
endif

.PHONY: all
all: test build

.PHONY: help
help: ## Show this help message
	@echo VIIPER Makefile
	@echo.
	@echo Usage: make [target]
	@echo.
	@echo Targets:
	@echo   help                 Show this help message
	@echo   deps                 Download Go dependencies
	@echo   tidy                 Tidy Go dependencies
	@echo   build                Build for current platform
	@echo   test                 Run tests
	@echo   test-coverage        Run tests with coverage
	@echo   clean                Remove build artifacts
	@echo   fmt                  Format Go code
	@echo   vet                  Run go vet
	@echo   lint                 Run golangci-lint
	@echo   run                  Build and run VIIPER
	@echo   run-server           Build and run VIIPER server
	@echo   docs-serve           Serve MkDocs documentation locally
	@echo   docs-build           Build MkDocs documentation
	@echo   docs-deploy          Deploy documentation to GitHub Pages
	@echo   version              Show version information

.PHONY: deps
deps: ## Download Go dependencies
	cd $(SRC_DIR) && go mod download

.PHONY: tidy
tidy: ## Tidy Go dependencies
	cd $(SRC_DIR) && go mod tidy

.PHONY: vet
vet: ## Run go vet
	cd $(SRC_DIR) && go vet ./...

.PHONY: test
test: ## Run tests
	cd $(SRC_DIR) && go test -count=1 -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	cd $(SRC_DIR) && go test -count=1 -coverprofile=coverage.out ./...
	cd $(SRC_DIR) && go tool cover -html=coverage.out -o coverage.html

.PHONY: build
build: ## Build for current platform
	cd $(SRC_DIR) && go build $(BUILD_FLAGS) -o ../$(DIST_DIR)/$(BINARY_NAME)$(EXE_EXT) $(MAIN_PKG)

.PHONY: clean
clean: ## Remove build artifacts
	-@if exist $(DIST_DIR) rmdir /S /Q $(DIST_DIR) 2>nul
	-@if exist $(SRC_DIR)\coverage.out del /Q $(SRC_DIR)\coverage.out 2>nul
	-@if exist $(SRC_DIR)\coverage.html del /Q $(SRC_DIR)\coverage.html 2>nul

.PHONY: fmt
fmt: ## Format Go code
	cd $(SRC_DIR) && go fmt ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint installed)
	cd $(SRC_DIR) && golangci-lint run

.PHONY: run
run: ## Build and run VIIPER
	cd $(SRC_DIR) && go run $(MAIN_PKG)

.PHONY: run-server
run-server: ## Build and run VIIPER server with default settings
	cd $(SRC_DIR) && go run $(MAIN_PKG) server

.PHONY: docs-serve
docs-serve: ## Serve MkDocs documentation locally (latest dev version)
	mike serve

.PHONY: docs-build
docs-build: ## Build MkDocs documentation
	mkdocs build

.PHONY: docs-deploy-dev
docs-deploy-dev: ## Deploy dev documentation version to GitHub Pages
	mike deploy --push --update-aliases dev latest

.PHONY: version
version: ## Show version information
	@echo Version: $(VERSION)
	@echo Commit:  $(COMMIT)
	@echo Built:   $(BUILD_TIME)
