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

/** Creates an element with optional class and text content. */
function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  className?: string,
  text?: string,
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}
