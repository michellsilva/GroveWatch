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
