# Changelog

All notable changes to grovewatch are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-08-31

### Added

- Go CLI (`grovewatch`) with four subcommands:
  - `scan` — walk a workspace and emit a deterministic provenance snapshot
    (files, toolchain, environment) as JSON, with per-file SHA-256 digests,
    a Merkle root over all files, and a canonical snapshot digest.
  - `diff` — compare two snapshots and explain drift per input category,
    exiting non-zero when drift is present (CI-gate friendly).
  - `verify` — recompute a snapshot's canonical digest and detect tampering.
  - `version` — print the tool version.
- Deterministic serialization: identical inputs yield an identical canonical
  digest, independent of filesystem traversal order or capture time.
- Environment values are hashed (never stored verbatim) to avoid leaking
  secrets while still detecting change.
- Zero-dependency TypeScript browser viewer that renders a bundled report:
  stats, a nested file tree with size roll-ups, a toolchain table, and an
  environment table. XSS-safe (textContent only), works from `file://` via an
  inline report block.
- Sample mixed-language workspace and a curated, verified sample report.
- Documentation: `README.md` with Mermaid diagrams and a demo, and
  `docs/PROVENANCE.md` describing the model and digest algorithm.
- Tooling: `Makefile`, GitHub Actions CI (Go + viewer), `LICENSE` (MIT),
  `.gitignore`, and focused Go and TypeScript tests.

[0.1.0]: https://github.com/michellsilva/GroveWatch/releases/tag/v0.1.0

<!-- draft note 1111 -->
