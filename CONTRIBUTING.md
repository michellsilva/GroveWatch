# Contributing

Thanks for your interest in improving this project.

## Ground rules

1. One logical change per pull request.
2. Run the test suite before pushing (see the README / Makefile for the exact commands).
3. Keep changes small and reviewable; describe what and why.

## Repository layout

- internal/provenance/ - the Go scanner, digests, diff and serializer
- cmd/grovewatch/ - the CLI entry point
- iewer/ - the TypeScript terminal viewer
- xamples/ - a sample workspace and a generated report
- docs/ - provenance format notes
