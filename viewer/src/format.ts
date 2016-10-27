/**
 * Pure, dependency-free helpers for the viewer. These contain no DOM access so
 * they are trivially unit-testable and reusable.
 */

import type { FileRecord, Snapshot, TreeNode } from "./types.js";

/** Formats a byte count as a human-readable string (e.g. "1.5 KB"). */
export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return "—";
  if (bytes < 1024) return `${bytes} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let value = bytes / 1024;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return `${value.toFixed(1)} ${units[unit]}`;
}

/** Returns the first `n` characters of a hex digest for compact display. */
export function shortDigest(digest: string, n = 12): string {
  if (!digest) return "—";
  return digest.length <= n ? digest : digest.slice(0, n);
}

/** Formats an ISO timestamp into a locale string, tolerating bad input. */
export function formatTimestamp(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso || "—";
  return d.toISOString().replace("T", " ").replace(/\.\d+Z$/, " UTC");
}

/**
 * Builds a nested tree from the flat, slash-separated file paths in a snapshot.
 * Directories are synthesized as needed and children are returned sorted with
 * directories first, then files, both alphabetically.
 */
export function buildTree(files: FileRecord[]): TreeNode {
  const root: TreeNode = {
    name: "",
    path: "",
    isDir: true,
    size: 0,
    children: [],
  };

  for (const file of files) {
    const parts = file.path.split("/").filter((p) => p.length > 0);
    let node = root;
    let accum = "";
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      accum = accum ? `${accum}/${part}` : part;
      const isLeaf = i === parts.length - 1;
      let child = node.children.find((c) => c.name === part);
      if (!child) {
        child = {
          name: part,
          path: accum,
          isDir: !isLeaf,
          size: isLeaf ? file.size : 0,
          digest: isLeaf ? file.digest : undefined,
          children: [],
        };
        node.children.push(child);
      }
      node = child;
    }
  }
