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
