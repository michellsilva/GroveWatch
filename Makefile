# grovewatch — build and quality targets.
#
# Go targets use only the standard toolchain. Viewer targets use the TypeScript
# compiler and Node's built-in test runner (no third-party runtime deps).

BIN        := bin/grovewatch
PKG        := ./...
SAMPLE_WS  := examples/sample-workspace
REPORT     := examples/report/provenance.json
VERSION    := 0.1.0
