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

