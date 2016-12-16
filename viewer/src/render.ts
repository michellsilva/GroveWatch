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

/** Renders the whole snapshot into the given container, replacing its content. */
export function render(container: HTMLElement, snap: Snapshot): void {
  container.replaceChildren(
    renderHeader(snap),
    renderStats(snap),
    renderSection("Files", renderFileTree(snap)),
    renderSection("Toolchain", renderTools(snap)),
    renderSection("Environment", renderEnv(snap)),
  );
}

function renderHeader(snap: Snapshot): HTMLElement {
  const header = el("header", "gw-header");
  header.append(
    el("h1", "gw-title", "grovewatch provenance"),
    el("p", "gw-sub", `${snap.tool.name} v${snap.tool.version} · schema ${snap.schema}`),
  );
  const digest = el("div", "gw-digest");
  digest.append(
    el("span", "gw-label", "digest "),
    el("code", "gw-mono", snap.digest),
  );
  header.append(digest);
  header.append(el("p", "gw-time", `captured ${formatTimestamp(snap.created_at)}`));
  return header;
}

function renderStats(snap: Snapshot): HTMLElement {
  const grid = el("div", "gw-stats");
  const stats: Array<[string, string]> = [
    ["Files", String(snap.workspace.file_count)],
    ["Total size", formatBytes(snap.workspace.total_bytes)],
    ["Tools found", `${countFoundTools(snap)} / ${snap.tools.length}`],
    ["Env set", `${countSetEnv(snap)} / ${snap.environment.length}`],
    ["Merkle root", shortDigest(snap.workspace.merkle_root)],
  ];
  for (const [label, value] of stats) {
    const card = el("div", "gw-stat");
    card.append(el("div", "gw-stat-value", value), el("div", "gw-stat-label", label));
    grid.append(card);
  }
  return grid;
}

function renderSection(title: string, body: HTMLElement): HTMLElement {
  const section = el("section", "gw-section");
  section.append(el("h2", "gw-section-title", title), body);
  return section;
}

function renderFileTree(snap: Snapshot): HTMLElement {
