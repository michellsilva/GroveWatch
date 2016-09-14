/**
 * Pure, dependency-free helpers for the viewer. These contain no DOM access so
 * they are trivially unit-testable and reusable.
 */

import type { FileRecord, Snapshot, TreeNode } from "./types.js";

/** Formats a byte count as a human-readable string (e.g. "1.5 KB"). */
export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return "—";
