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
