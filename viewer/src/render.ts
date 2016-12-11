/**
 * DOM rendering for the grovewatch viewer. Everything is built with the native
 * DOM API — no frameworks, no dependencies. All user-controlled strings go
 * through textContent, never innerHTML, so the viewer is XSS-safe by default.
 */

import type { Snapshot, TreeNode } from "./types.js";
import {
  buildTree,
  countFoundTools,
  countSetEnv,
  formatBytes,
  formatTimestamp,
  shortDigest,
} from "./format.js";

