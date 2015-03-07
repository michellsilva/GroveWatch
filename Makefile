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
