/**
 * Tests for the pure viewer helpers using Node's built-in test runner
 * (node:test) and assertions (node:assert). No third-party dependencies.
 *
 * Run with: node --test (after compiling with tsc), or via `make viewer-test`.
 * These import the compiled ES modules from ../dist.
 */

import test from "node:test";
import assert from "node:assert/strict";

import {
  buildTree,
  countFoundTools,
  countSetEnv,
  formatBytes,
  formatTimestamp,
  shortDigest,
} from "../dist/format.js";

test("formatBytes scales units", () => {
  assert.equal(formatBytes(0), "0 B");
  assert.equal(formatBytes(512), "512 B");
  assert.equal(formatBytes(1024), "1.0 KB");
  assert.equal(formatBytes(1536), "1.5 KB");
  assert.equal(formatBytes(1048576), "1.0 MB");
  assert.equal(formatBytes(-1), "—");
});

test("shortDigest truncates", () => {
  assert.equal(shortDigest("abcdef0123456789", 8), "abcdef01");
  assert.equal(shortDigest("abc", 8), "abc");
  assert.equal(shortDigest(""), "—");
});

test("formatTimestamp handles bad input", () => {
  assert.equal(formatTimestamp("not-a-date"), "not-a-date");
  const out = formatTimestamp("2026-01-02T03:04:05.678Z");
  assert.match(out, /2026-01-02 03:04:05/);
});

test("buildTree nests paths and sorts dirs first", () => {
  const tree = buildTree([
    { path: "src/main.go", size: 100, mode: "0644", digest: "d1" },
    { path: "README.md", size: 50, mode: "0644", digest: "d2" },
    { path: "src/util/helper.go", size: 25, mode: "0644", digest: "d3" },
  ]);
