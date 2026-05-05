<!-- grovewatch — a forest ranger's provenance field guide -->

<p align="center">
  <img src="docs/assets/provenance-tree.svg" alt="grovewatch provenance tree: roots are inputs, trunk is the build, canopy is the snapshot" width="100%">
</p>

<h1 align="center">grovewatch</h1>

<p align="center">
  <em>A software-provenance sentinel that fingerprints a workspace into a
  deterministic, tamper-evident snapshot — and tells you exactly what drifted.</em>
</p>

<p align="center">
  <strong>Go CLI, stdlib only</strong> ·
  <strong>zero-dependency TypeScript viewer</strong> ·
  <strong>secret-safe</strong> · <strong>MIT</strong>
</p>

---

## Field notes from the grove

Every build stands on a tree of inputs. In the **roots** are the things you feed
it: workspace files, the toolchain on the host, and the environment variables
that colour the run. Up the **trunk** runs one deterministic computation — a
canonical digest and a Merkle root over the files. In the **canopy** hangs the
fruit: a single JSON snapshot naming every input, sealed under a self-describing
hash. grovewatch is the ranger that walks this tree: it records the roots, folds
them through the trunk, tags the canopy, and comes back later to tell you which
leaves changed colour. That is *provenance* — given the same inputs, do we get
the same thing, and if not, **what changed?**

---

## Trail map

- [What grovewatch is (and is not)](#what-grovewatch-is-and-is-not)
- [The three roots: what gets captured](#the-three-roots-what-gets-captured)
- [The trunk: digests, Merkle root, determinism](#the-trunk-digests-merkle-root-determinism)
- [Install & build](#install--build)
- [Three journeys: scan → verify → diff](#three-journeys-scan--verify--diff)
- [CLI reference](#cli-reference)
- [The snapshot schema, field by field](#the-snapshot-schema-field-by-field)
- [Reading a drift report](#reading-a-drift-report)
- [The viewer](#the-viewer)
- [Use cases](#use-cases)
- [Recipes](#recipes)
- [Trust boundary & threat model](#trust-boundary--threat-model)
- [Troubleshooting](#troubleshooting)
- [Repository layout](#repository-layout)
- [Testing](#testing)
- [Limitations](#limitations)
- [Roadmap](#roadmap)
- [Further reading](#further-reading)

---

## What grovewatch is (and is not)

**It is** a small, auditable tool that answers one question well: a `scan`
produces a portable JSON record of *inputs*, a `verify` proves that record was
not edited since, a `diff` explains the difference between two records in plain,
category-grouped language, and a browser `viewer` renders any record with no
build step or network call.

**It is not** a runtime tracer. grovewatch does **not** hook syscalls, use eBPF,
or observe processes as they execute. It reads the filesystem, resolves tool
names on `PATH`, runs `<tool> --version`, and reads a fixed list of
environment-variable *keys* — everything it knows, it learned by looking, never
by intercepting. If you need process/syscall provenance, this is the wrong tool;
if you need reproducible-input fingerprints you can diff and verify, it is the
right size.

---

## The three roots: what gets captured

| Root | Recorded | Not recorded |
|------|----------|--------------|
| **Files** | Relative slash-path, size, octal mode, streamed SHA-256 of contents (sorted by path) | Symlinks, devices, sockets (non-portable), and ignored segments |
| **Tools** | Logical name, resolved `PATH` location, first dotted version number, a `found` flag | Full command output — only the version string is kept |
| **Environment** | Key, a `set` flag, and the SHA-256 of the value when set | The raw value — **never** stored, so snapshots are safe to share |

Directories are not stored as records; they are *implied* by file paths and
reconstructed by the viewer into a tree. Defaults are deliberately small and
overridable per scan: files are everything under root minus
`-ignore .git,node_modules,dist,build,.grovewatch`; tools default to
`-tools go,node,python,git`; environment keys default to
`-env CI,GOOS,GOARCH,NODE_ENV`.

---

## The trunk: digests, Merkle root, determinism

Three SHA-256 hashes (hex-encoded) hold the tree together.

**1 · Per-file digest** — each regular file is *streamed* through SHA-256, so a
multi-gigabyte input still uses bounded memory: `file.digest = hex(SHA-256(contents))`.

**2 · Merkle root** — files are sorted by path, then folded into one hash. Each
field is *length-prefixed* (8-byte little-endian length, then bytes) so adjacent
fields can never be confused (`"ab"+"c"` ≠ `"a"+"bc"`):

```
h = SHA-256()
for f in sort_by_path(files):
    h.update( len(f.path) || f.path );  h.update( len(f.digest) || f.digest )
merkle_root = hex(h)
```

It changes if *any* file is added, removed, renamed, or edited — a fast "did the
file set move at all?" check.

**3 · Canonical snapshot digest** — the snapshot's identity is a SHA-256 over a
*canonical view* that deliberately excludes the volatile `created_at` **and the
digest field itself**:

```
view   = { schema, tool, workspace, files, tools, environment }
digest = hex( SHA-256( json_marshal(view) ) )
```

Every collection is pre-sorted (files by path, tools by name, env by key) and
Go's `encoding/json` emits struct fields in declaration order, so the serialized
view is byte-stable. **Identical inputs therefore yield an identical digest —
across machines and across time.** That is the point of the trunk: strip the
noise (when, on what host) and hash only what feeds the build.

> Full specification: **[docs/PROVENANCE.md](docs/PROVENANCE.md)**.

---

## Install & build

Requirements: **Go ≥ 1.24** for the CLI; **Node ≥ 18** + **TypeScript ≥ 5** for
the viewer (dev only — the shipped viewer has no runtime deps).

```sh
make build      # compile the CLI to bin/grovewatch
make viewer     # type-check, build, and test the viewer
make ci         # everything CI runs: vet + tests + viewer + verify sample
```

Or directly: `go build -o bin/grovewatch ./cmd/grovewatch`, and in `viewer/`,
`npx tsc -p tsconfig.json && node --test`. Run `make help` to list all targets.

---

## Three journeys: scan → verify → diff

### Journey 1 — scan the grove

Walk the bundled sample workspace and write a snapshot. The summary line goes to
stderr; the snapshot goes to the file (or stdout without `-out`):

```console
$ grovewatch scan -out examples/report/provenance.json examples/sample-workspace
wrote examples/report/provenance.json (4 files, digest b8afd4eb68c2)
```

### Journey 2 — verify the seal

`verify` recomputes the canonical digest and compares it to the one stored in the
file, so a single edited byte breaks the seal (exit `4`):

```console
$ grovewatch verify examples/report/provenance.json
OK: digest b8afd4eb68c2 matches content

$ grovewatch verify tampered.json
MISMATCH: stored b8afd4eb68c2 but content hashes to 41c9a7de0b52
```

### Journey 3 — diff the drift

Change some inputs, scan again, and compare against the baseline. `diff` groups
changes by category (files → tools → env) and sorts them for a stable report:

```console
$ echo "// new feature" >> workspace/main.go
$ rm workspace/config.json
$ echo "console.log('new');" > workspace/web/new.js
$ grovewatch scan -out new.json workspace
$ grovewatch diff examples/report/provenance.json new.json
Drift detected: 1 added, 1 modified, 1 removed.
  old b8afd4eb68c2 -> new 2990f2bef8e0
  [file/removed] config.json: file removed (was 131 bytes)
  [file/modified] main.go: content changed 2362f7c57b9e -> 96f65db9dc8d (218 -> 236 bytes)
  [file/added] web/new.js: new file (21 bytes, digest 5828a22c8b1b)
# exit status 3 — drift present, ideal for CI gates
```

<p align="center">
  <img src="docs/assets/drift-terminal.svg" alt="Animated terminal panorama: scan writes a snapshot, verify confirms the digest, diff prints an added/modified/removed drift report" width="100%">
</p>

> The transcripts above mirror a real run against the bundled sample workspace on
> this repository — the digest `b8afd4eb68c2` matches
> [`examples/report/provenance.json`](examples/report/provenance.json).

---

## CLI reference

```
grovewatch scan   [flags] <workspace>   Scan a workspace, emit provenance JSON
grovewatch diff   [-json] <old> <new>   Compare two snapshots, explain drift
grovewatch verify <snapshot.json>       Verify a snapshot's digest integrity
grovewatch version                      Print the version
```

**`scan` flags**

| Flag | Default | Purpose |
|------|---------|---------|
| `-out` | *(stdout)* | Write the snapshot to a file; a summary goes to stderr |
| `-ignore` | `.git,node_modules,dist,build,.grovewatch` | Path segments to skip (segment-boundary match) |
| `-tools` | `go,node,python,git` | Tool names to resolve on `PATH` |
| `-env` | `CI,GOOS,GOARCH,NODE_ENV` | Env keys to record (values hashed) |

**`diff` flag** — `-json` emits the diff as structured JSON instead of text.

**Exit codes** — `0` success / no drift · `1` I/O or parse error · `2` usage
error · `3` `diff` found drift · `4` `verify` digest mismatch. Codes `3` and `4`
are the ones you wire into CI gates.

---

## The snapshot schema, field by field

A snapshot is a single JSON object. Abridged from the bundled sample:

```json
{
  "schema": "1.0.0",
  "tool": { "name": "grovewatch", "version": "0.1.0" },
  "workspace": { "root": "examples/sample-workspace", "file_count": 4,
                 "total_bytes": 825, "merkle_root": "5be5329bef2c…247e" },
  "files":       [ { "path": "README.md", "size": 298, "mode": "0666", "digest": "45fef498…0f95" } ],
  "tools":       [ { "name": "go", "path": "…/go.exe", "version": "1.24.4", "found": true } ],
  "environment": [ { "key": "CI", "value_digest": "", "set": false } ],
  "digest": "b8afd4eb68c2…689d",
  "created_at": "2026-08-31T16:17:09Z"
}
```

| Field | Type | Notes |
|-------|------|-------|
| `schema` | SemVer | Currently `1.0.0`; gate on the major version |
| `tool` | `{name, version}` | Producer identity |
| `workspace` | object | `root`, `file_count`, `total_bytes`, `merkle_root` (hex) |
| `files[]` | `{path, size, mode, digest}` | Sorted by path; `path` always uses `/` |
| `tools[]` | `{name, path, version, found}` | Sorted by name |
| `environment[]` | `{key, value_digest, set}` | Sorted by key; value never stored |
| `digest` | hex | Canonical identity (excludes `created_at` and itself) |
| `created_at` | RFC 3339 UTC | **Volatile** — excluded from the digest |

Paths are normalised to forward slashes even on Windows, so a snapshot from one
OS is comparable to one from another.

---

## Reading a drift report

Every text-mode change line has one shape — `[<category>/<kind>] <name>: <detail>`
— where **category** is `file`/`tool`/`env`, **kind** is `added` (new only),
`removed` (old only), or `modified` (in both, content differs), and **detail** is
a human sentence (byte-size deltas for files, `version changed …` for tools,
`value changed …` for env).

A **file** counts as modified when its content digest differs (or, same content,
its octal mode differs); a **tool** when its version differs, availability flips,
or resolved path differs; an **env** var when its value's SHA-256 differs or the
`set` flag flips. For machines, `grovewatch diff -json` emits the same data as a
structured object (`old_digest`, `new_digest`, `identical`, `changes[]`), still
exiting `3` on drift.

---

## The viewer

The viewer is plain TypeScript compiled to ES modules — **no framework, no
bundler, no runtime dependency.** It resolves its report in order: (1) a
`?report=<url>` query parameter; (2) an inline
`<script id="gw-report" type="application/json">` block (how it works straight
from `file://`, no server needed); (3) otherwise `./provenance.json` relative to
the page.

It then renders summary stat cards, a nested **file tree** with directory size
roll-ups, a **toolchain** table, and an **environment** table. Every
user-controlled string is written through `textContent` — **never** `innerHTML`
— so a hostile snapshot cannot inject markup. The pure logic (byte formatting,
digest truncation, timestamp handling, tree building) lives in
`viewer/src/format.ts` and is unit-tested with Node's built-in `node:test`.

```sh
make viewer-build         # compile viewer/src → viewer/dist
node scripts/serve.mjs    # tiny stdlib static server on :8080
# then open http://localhost:8080/viewer/public/index.html
```

The shipped `viewer/public/index.html` also opens straight from disk — it carries
an inline copy of the sample report.

---

## Use cases

- **Reproducible-build gate** — baseline a known-good environment; fail CI when
  inputs drift (exit `3`).
- **Supply-chain audit** — keep signed-off snapshots of what an artifact was
  built from, and diff releases to explain which inputs moved.
- **Toolchain drift detection** — catch a runner silently upgrading `go` or
  `node` under you.
- **Config integrity** — detect an unexpected edit or removal of a config file
  without diffing whole trees by hand.
- **Tamper alarm** — ship the snapshot with a release so anyone can `verify` the
  record was not altered.

---

## Recipes

**Fail a pipeline on any input drift** — `diff` exits `3` on drift, so CI fails:

```sh
grovewatch scan -out current.json .
grovewatch diff baseline.json current.json
```

**Track a specific toolchain and secret set** — secrets are stored only as
`set` + a value hash, never verbatim:

```sh
grovewatch scan -tools go,node,rustc,docker \
  -env CI,NODE_ENV,DATABASE_URL,AWS_REGION -out snap.json .
```

Also: **ignore generated dirs** with `-ignore .git,node_modules,dist,build,target,coverage`;
**refresh & re-verify** the sample with `make report` then `make verify`;
**machine-readable drift** via `grovewatch diff -json baseline.json current.json > drift.json`.

---

## Trust boundary & threat model

grovewatch is honest about what its guarantees mean.

- **The digest proves the *snapshot* is intact, not the *world*.** `verify`
  detects edits to the JSON after it was written; it cannot tell that the
  workspace was already compromised at scan time — garbage in, sealed garbage out.
- **Env values are hashed, not encrypted.** A value's SHA-256 detects change and
  confirms equality without revealing the secret, but it does not hide *which
  keys* you track and a guessed value can be confirmed. Change-detector, not vault.
- **Tools are trusted to report their own version.** grovewatch records whatever
  `<tool> --version` prints; a malicious binary on `PATH` can lie.
- **No network, no execution beyond version probes**, no telemetry — nothing
  leaves the host.
- **Integrity, not authenticity.** The digest is not a signature; to prove *who*
  produced a snapshot, sign the file with your own key on top of grovewatch.

---

## Troubleshooting

| Symptom | Likely cause & fix |
|---------|--------------------|
| `verify` prints `MISMATCH` | JSON edited after scanning, or an incompatible schema major. Re-scan to regenerate. |
| Tool shows `found: false` | Not on the `PATH` grovewatch inherited. Check the shell/CI env, or drop it from `-tools`. |
| `version` empty but `found: true` | The binary prints no dotted version to `--version`/`version`/`-version`. Path is still recorded. |
| Digest differs across machines for "identical" inputs | Something in the roots really differs — often file **mode** bits or a tool version. Run `diff` to see. |
| `diff` exits non-zero in CI | By design: exit `3` means drift was found — a gate signal, not a crash. |
| Viewer shows "Failed to load report" | Could not resolve a report from `?report=`, the inline block, or `./provenance.json`. Serve the folder or use the inline build. |

---

## Repository layout

```
grovewatch/
├─ cmd/grovewatch/main.go     CLI entry point & subcommands
├─ internal/provenance/       core library (stdlib only)
│  ├─ model.go                snapshot data model + schema version
│  ├─ scan.go                 workspace / tool / env scanner
│  ├─ digest.go               SHA-256 helpers + Merkle root
│  ├─ serialize.go            deterministic JSON, ComputeDigest, Verify
│  ├─ diff.go                 drift comparison engine + Summary()
│  └─ *_test.go               focused tests
├─ viewer/                    zero-dep TypeScript viewer (src · test · public)
├─ examples/                  sample-workspace/ + report/provenance.json
├─ docs/                      PROVENANCE.md spec + assets/ (this guide's SVGs)
├─ scripts/serve.mjs          tiny stdlib static server
└─ Makefile · .github/workflows/ci.yml · LICENSE · CHANGELOG.md
```

---

## Testing

```sh
make test          # Go tests: scan determinism, diff classification, verify
make viewer-test   # viewer tests: byte formatting, digest truncation, tree build
make ci            # vet + Go tests + viewer build/test + verify sample
```

The Go tests assert that identical inputs yield an identical digest, that the
digest is independent of capture time, that drift is classified into the right
category and kind, and that tampering is detected. The viewer tests cover byte
formatting, digest truncation, timestamp handling, and tree construction.---

## Limitations

- **Inputs, not process** — grovewatch fingerprints files, tools, and env keys;
  it does not trace what a build *does* (no syscall/eBPF instrumentation).
- **Regular files only** — symlinks, devices, and sockets are skipped; empty
  directories leave no trace, since directories are implied by file paths.
- **Version parsing is heuristic** — the first dotted number from `--version` is
  not always the semantic tool version.
- **Whole-file granularity** — `diff` reports that a file changed and its size
  delta, not a line-level content diff.
- **Digest ≠ signature** — integrity is guaranteed, authenticity is not; sign the
  snapshot separately if you need it.

---

## Roadmap

Directional, not committed (current release `0.1.0`): optional detached signing
(minisign / cosign) for authenticity atop integrity; a `-format` flag for compact
single-line JSON; glob-based ignore patterns; a browser side-by-side diff of two
snapshots; optional file-timestamp recording behind an explicit flag.

---

## Further reading

- **[docs/PROVENANCE.md](docs/PROVENANCE.md)** — the data model and digest algorithm in full
- **[CHANGELOG.md](CHANGELOG.md)** — release history · **[LICENSE](LICENSE)** — MIT
- **[examples/report/provenance.json](examples/report/provenance.json)** — the verified sample snapshot

---

<p align="center">
  <em>Built with the Go standard library and dependency-free TypeScript.<br>
  Roots to canopy, nothing leaves the grove.</em>
</p>

<!-- draft note 1477 -->
