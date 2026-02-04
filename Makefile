# grovewatch — build and quality targets.
#
# Go targets use only the standard toolchain. Viewer targets use the TypeScript
# compiler and Node's built-in test runner (no third-party runtime deps).

BIN        := bin/grovewatch
PKG        := ./...
SAMPLE_WS  := examples/sample-workspace
REPORT     := examples/report/provenance.json
VERSION    := 0.1.0

.PHONY: all build test vet fmt clean \
        viewer viewer-build viewer-test viewer-typecheck \
        report verify demo ci help

all: build viewer-build ## Build the CLI and the viewer

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

## ---- Go ----

build: ## Compile the grovewatch CLI to bin/
	@mkdir -p bin
	go build -ldflags "-X main.version=$(VERSION)" -o $(BIN) ./cmd/grovewatch

test: ## Run Go tests
	go test $(PKG)

vet: ## Run go vet
	go vet $(PKG)

fmt: ## Format Go sources
	gofmt -w $(shell git ls-files '*.go' 2>/dev/null || echo .)

## ---- Provenance data ----

report: build ## Regenerate the sample provenance report
	$(BIN) scan -out $(REPORT) $(SAMPLE_WS)

verify: build ## Verify the sample report's digest integrity
	$(BIN) verify $(REPORT)

demo: build ## Run a full scan/verify/diff walkthrough
	@echo "== scan ==" && $(BIN) scan $(SAMPLE_WS) | head -n 20
	@echo "== verify ==" && $(BIN) verify $(REPORT)

## ---- Viewer (TypeScript) ----

viewer: viewer-build viewer-test ## Build and test the viewer

viewer-build: ## Compile the TypeScript viewer to viewer/dist
	cd viewer && npx tsc -p tsconfig.json

viewer-typecheck: ## Type-check the viewer without emitting
	cd viewer && npx tsc -p tsconfig.json --noEmit

viewer-test: viewer-build ## Run viewer unit tests (node:test)
	cd viewer && node --test

## ---- Aggregate ----

ci: vet test viewer-build viewer-test verify ## Everything CI runs

clean: ## Remove build artifacts
	rm -rf bin viewer/dist

<!-- draft note 1437 -->
