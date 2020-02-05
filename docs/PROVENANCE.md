# Provenance Model & Digest Algorithm

This document specifies the grovewatch provenance format, the guarantees it
provides, and exactly how digests are computed. It is the reference for anyone
implementing a compatible producer or consumer.

## Goals

1. **Deterministic.** Two scans of identical inputs produce an identical
   canonical digest, regardless of filesystem traversal order or capture time.
2. **Tamper-evident.** A snapshot carries a self-describing digest so any later
   modification is detectable with `grovewatch verify`.
3. **Secret-safe.** Environment values are hashed, never stored verbatim.
4. **Explainable.** Two snapshots can be diffed into a per-input list of
   added / modified / removed changes.

## Schema

The on-disk format is JSON. `schema` carries a SemVer string (currently
`1.0.0`); consumers should reject or adapt to unknown major versions.

```
Snapshot
├─ schema        string            format version
├─ tool          { name, version } producer identity
├─ workspace
│  ├─ root         string          scanned path as supplied
│  ├─ file_count   int
│  ├─ total_bytes  int64
│  └─ merkle_root  hex             hash over all file (path, digest) pairs
├─ files[]       { path, size, mode, digest }   sorted by path
├─ tools[]       { name, path, version, found } sorted by name
├─ environment[] { key, value_digest, set }     sorted by key
├─ digest        hex               canonical snapshot identity
└─ created_at    RFC3339 UTC       volatile; excluded from digest
```

### Files

Each regular file under the workspace root (minus ignored segments) becomes one
record:

- `path` — slash-separated, relative to the root. Always uses `/` even on
  Windows so snapshots are portable across operating systems.
- `size` — byte length.
- `mode` — octal permission bits, e.g. `0644`.
- `digest` — hex SHA-256 of the file contents, streamed so large files use
  bounded memory.

Symlinks, devices, and other non-regular files are skipped because their
content is not portable. Directories are not recorded directly; they are
implied by file paths and reconstructed by the viewer.

### Tools

For each requested tool name, grovewatch resolves it on `PATH`. If found, it
runs `<tool> --version` (falling back to `version` / `-version`) and extracts
the first dotted version number. Only the discovered path and version are
recorded — never full command output.

### Environment

For each requested variable, grovewatch records whether it is `set` and, if so,
the SHA-256 of its value. The raw value is never persisted, so a snapshot can be
shared without leaking tokens or credentials while still detecting when a value
changes.

## Digest algorithm
