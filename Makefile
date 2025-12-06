# VIIPER Makefile
# Cross-platform build automation for VIIPER

############################################################
# Variables
# These are defined in a cross-platform way. We branch early
# so that later variable definitions do not need per-OS logic.
############################################################

BINARY_NAME := viiper
MAIN_PKG := ./cmd/viiper
SRC_DIR := .
DIST_DIR := dist

# OS-specific helpers
ifeq ($(OS),Windows_NT)
	NULL_DEVICE := nul
	DATE_CMD := powershell -NoProfile -NonInteractive -Command "Get-Date -Format 'yyyy-MM-dd_HH:mm:ss'"
	EXE_EXT := .exe
	RM_DIR := rmdir /S /Q
	RM_FILE := del /Q
	COVERAGE_OUT := $(SRC_DIR)\coverage.out
	COVERAGE_HTML := $(SRC_DIR)\coverage.html
	export CGO_ENABLED=0
else
	NULL_DEVICE := /dev/null
	DATE_CMD := date -u +"%Y-%m-%d_%H:%M:%S"
	EXE_EXT :=
	RM_DIR := rm -rf
	RM_FILE := rm -f
	COVERAGE_OUT := $(SRC_DIR)/coverage.out
	COVERAGE_HTML := $(SRC_DIR)/coverage.html
	export CGO_ENABLED=0
endif

# Git-derived metadata (robust to missing git by redirecting errors)
VERSION ?= $(shell git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --always 2>$(NULL_DEVICE) || echo v0.0.0-dev)
COMMIT := $(shell git rev-parse --short HEAD 2>$(NULL_DEVICE) || echo unknown)
BUILD_TIME := $(shell $(DATE_CMD))

# Go build flags
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(BUILD_TIME) -X github.com/Alia5/VIIPER/internal/codegen/common.Version=$(VERSION)
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

# Windows resource embedding
VERSIONINFO_JSON := versioninfo.json
RESOURCE_SYSO := cmd/viiper/resource.syso

.PHONY: all
all: test build

.PHONY: help
help: ## Show this help message
	@echo VIIPER Makefile
	@echo.
	@echo Usage: make [target]
	@echo.
	@echo Build Targets:
	@echo   build                Build VIIPER for current platform
	@echo   clean                Remove build artifacts
	@echo   test                 Run tests
	@echo   test-coverage        Run tests with coverage
	@echo.
	@echo SDK Code Generation:
	@echo   codegen-all          Generate all SDK client libraries
	@echo   codegen-c            Generate C SDK
	@echo   codegen-cpp          Generate C++ SDK
	@echo   codegen-csharp       Generate C# SDK
	@echo   codegen-rust         Generate Rust SDK
	@echo   codegen-typescript   Generate TypeScript SDK
	@echo.
	@echo SDK Building:
	@echo   build-sdks           Build all SDK client libraries
	@echo   build-sdk-c          Build C SDK
	@echo   build-sdk-cpp        Build C++ SDK
	@echo   build-sdk-csharp     Build C# SDK
	@echo   build-sdk-rust       Build Rust SDK
	@echo   build-sdk-typescript Build TypeScript SDK
	@echo.
	@echo Example Building:
	@echo   build-examples       Build all examples for all SDKs
	@echo   build-examples-c     Build C examples
	@echo   build-examples-cpp   Build C++ examples
	@echo   build-examples-csharp Build C# examples
	@echo   build-examples-rust  Build Rust examples
	@echo   build-examples-typescript Build TypeScript examples
	@echo.
	@echo Cleaning:
	@echo   clean-sdks           Clean SDK build artifacts
	@echo   clean-examples       Clean example build artifacts
	@echo.
	@echo Complete Rebuild:
	@echo   rebuild-all          Clean, regenerate, and build all SDKs and examples
	@echo.
	@echo Other Targets:
	@echo   help                 Show this help message
	@echo   deps                 Download Go dependencies
	@echo   tidy                 Tidy Go dependencies
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

.PHONY: generate-versioninfo
generate-versioninfo: ## Generate Windows version info resource
ifeq ($(OS),Windows_NT)
	@echo Generating Windows version info...
	@go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
	@powershell -NoProfile -NonInteractive -File scripts/inject-version.ps1 "$(VERSION)" "$(VERSIONINFO_JSON)" "versioninfo.tmp.json"
	@cd $(SRC_DIR) && goversioninfo -64 -o $(RESOURCE_SYSO) versioninfo.tmp.json
	@del versioninfo.tmp.json
else
	@echo Skipping versioninfo generation on non-Windows platform
endif

.PHONY: clean-versioninfo
clean-versioninfo: ## Remove generated Windows version info resource
	-@$(RM_FILE) $(RESOURCE_SYSO) 2>$(NULL_DEVICE)
	-@$(RM_FILE) versioninfo.tmp.json 2>$(NULL_DEVICE)

.PHONY: build
build: ## Build for current platform
ifeq ($(OS),Windows_NT)
	@$(MAKE) generate-versioninfo
endif
	cd $(SRC_DIR) && go build $(BUILD_FLAGS) -o $(DIST_DIR)/$(BINARY_NAME)$(EXE_EXT) $(MAIN_PKG)

.PHONY: clean
clean: clean-versioninfo ## Remove build artifacts
	-@$(RM_DIR) $(DIST_DIR) 2>$(NULL_DEVICE)
	-@$(RM_FILE) $(COVERAGE_OUT) 2>$(NULL_DEVICE)
	-@$(RM_FILE) $(COVERAGE_HTML) 2>$(NULL_DEVICE)

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

############################################################
# SDK Code Generation
############################################################

CLIENTS_DIR := clients

.PHONY: codegen-all
codegen-all: ## Generate all SDK client libraries (C, C++, C#, Rust, TypeScript)
	@echo Generating all SDK clients...
	cd $(SRC_DIR) && go run $(MAIN_PKG) codegen --lang all --output $(CLIENTS_DIR)

.PHONY: codegen-c
codegen-c: ## Generate C SDK client library
	cd $(SRC_DIR) && go run $(MAIN_PKG) codegen --lang c --output $(CLIENTS_DIR)

.PHONY: codegen-cpp
codegen-cpp: ## Generate C++ SDK client library
	cd $(SRC_DIR) && go run $(MAIN_PKG) codegen --lang cpp --output $(CLIENTS_DIR)

.PHONY: codegen-csharp
codegen-csharp: ## Generate C# SDK client library
	cd $(SRC_DIR) && go run $(MAIN_PKG) codegen --lang csharp --output $(CLIENTS_DIR)

.PHONY: codegen-rust
codegen-rust: ## Generate Rust SDK client library
	cd $(SRC_DIR) && go run $(MAIN_PKG) codegen --lang rust --output $(CLIENTS_DIR)

.PHONY: codegen-typescript
codegen-typescript: ## Generate TypeScript SDK client library
	cd $(SRC_DIR) && go run $(MAIN_PKG) codegen --lang typescript --output $(CLIENTS_DIR)

############################################################
# SDK Building
############################################################

.PHONY: build-sdks
build-sdks: build-sdk-c build-sdk-cpp build-sdk-csharp build-sdk-rust build-sdk-typescript ## Build all SDK client libraries

.PHONY: build-sdk-c
build-sdk-c: ## Build C SDK
	@echo Building C SDK...
	@if exist $(CLIENTS_DIR)\c (cd $(CLIENTS_DIR)\c && cmake -B build -S . -DCMAKE_BUILD_TYPE=Release && cmake --build build --config Release) else (echo C SDK not generated yet. Run 'make codegen-c' first.)

.PHONY: build-sdk-cpp
build-sdk-cpp: ## Build C++ SDK (header-only, no build needed)
	@echo C++ SDK is header-only - no build needed.

.PHONY: build-sdk-csharp
build-sdk-csharp: ## Build C# SDK
	@echo Building C# SDK...
	@if exist $(CLIENTS_DIR)\csharp (cd $(CLIENTS_DIR)\csharp\Viiper.Client && dotnet build) else (echo C# SDK not generated yet. Run 'make codegen-csharp' first.)

.PHONY: build-sdk-rust
build-sdk-rust: ## Build Rust SDK
	@echo Building Rust SDK...
	@if exist $(CLIENTS_DIR)\rust (cd $(CLIENTS_DIR)\rust && cargo build) else (echo Rust SDK not generated yet. Run 'make codegen-rust' first.)

.PHONY: build-sdk-typescript
build-sdk-typescript: ## Build TypeScript SDK
	@echo Building TypeScript SDK...
	cd $(CLIENTS_DIR)\typescript && npm install && npm run build

############################################################
# Example Building
############################################################

.PHONY: build-examples
build-examples: build-examples-c build-examples-cpp build-examples-csharp build-examples-rust build-examples-typescript ## Build all examples for all SDKs

.PHONY: build-examples-c
build-examples-c: ## Build C examples
	@echo Building C examples...
	cd examples\c && cmake -B build -S . -DCMAKE_BUILD_TYPE=Release && cmake --build build --config Release

.PHONY: build-examples-cpp
build-examples-cpp: ## Build C++ examples
	@echo Building C++ examples...
	cd examples\cpp && cmake -B build -S . -DCMAKE_BUILD_TYPE=Release && cmake --build build --config Release

.PHONY: build-examples-csharp
build-examples-csharp: ## Build C# examples
	@echo Building C# examples...
	cd examples\csharp\virtual_keyboard && dotnet build
	cd examples\csharp\virtual_mouse && dotnet build
	cd examples\csharp\virtual_x360_pad && dotnet build

.PHONY: build-examples-rust
build-examples-rust: ## Build Rust examples
	@echo Building Rust examples...
	cd examples\rust && cargo build --workspace

.PHONY: build-examples-typescript
build-examples-typescript: build-sdk-typescript ## Build TypeScript examples
	@echo Building TypeScript examples...
	cd examples\typescript && npm install && npm run build

############################################################
# Cleaning
############################################################

.PHONY: clean-sdks
clean-sdks: ## Clean all SDK build artifacts
	@echo Cleaning SDK build artifacts...
	-@$(RM_DIR) $(CLIENTS_DIR)\c\build 2>$(NULL_DEVICE)
	-@$(RM_DIR) $(CLIENTS_DIR)\csharp\Viiper.Client\bin 2>$(NULL_DEVICE)
	-@$(RM_DIR) $(CLIENTS_DIR)\csharp\Viiper.Client\obj 2>$(NULL_DEVICE)
	-@$(RM_DIR) $(CLIENTS_DIR)\rust\target 2>$(NULL_DEVICE)
	-@$(RM_DIR) $(CLIENTS_DIR)\typescript\node_modules 2>$(NULL_DEVICE)
	-@$(RM_DIR) $(CLIENTS_DIR)\typescript\dist 2>$(NULL_DEVICE)

.PHONY: clean-examples
clean-examples: ## Clean all example build artifacts
	@echo Cleaning example build artifacts...
	-@$(RM_DIR) examples\c\build 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\cpp\build 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\csharp\virtual_keyboard\bin 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\csharp\virtual_keyboard\obj 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\csharp\virtual_mouse\bin 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\csharp\virtual_mouse\obj 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\csharp\virtual_x360_pad\bin 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\csharp\virtual_x360_pad\obj 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\rust\target 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\typescript\node_modules 2>$(NULL_DEVICE)
	-@$(RM_DIR) examples\typescript\dist 2>$(NULL_DEVICE)

############################################################
# Complete Rebuild
############################################################

.PHONY: rebuild-all
rebuild-all: clean-sdks clean-examples codegen-all build-sdks build-examples ## Complete rebuild: clean, regenerate all SDKs, build all SDKs and examples
	@echo.
	@echo ============================================================
	@echo REBUILD COMPLETE
	@echo ============================================================
	@echo All SDKs have been regenerated and built.
	@echo All examples have been built.
	@echo ============================================================
