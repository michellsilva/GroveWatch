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
  const tree = buildTree(snap.files);
  const list = el("ul", "gw-tree");
  for (const child of tree.children) {
    list.append(renderTreeNode(child));
  }
  if (tree.children.length === 0) {
    list.append(el("li", "gw-empty", "no files recorded"));
  }
  return list;
}

function renderTreeNode(node: TreeNode): HTMLElement {
  const li = el("li", node.isDir ? "gw-node gw-dir" : "gw-node gw-file");
  const row = el("div", "gw-row");
  row.append(el("span", "gw-icon", node.isDir ? "▸" : "·"));
  row.append(el("span", "gw-name", node.name));
  row.append(el("span", "gw-size", formatBytes(node.size)));
  if (!node.isDir && node.digest) {
    const code = el("code", "gw-mono gw-hash", shortDigest(node.digest));
    code.title = node.digest;
    row.append(code);
  }
  li.append(row);
  if (node.isDir && node.children.length > 0) {
    const sub = el("ul", "gw-tree");
    for (const child of node.children) sub.append(renderTreeNode(child));
    li.append(sub);
  }
  return li;
}

function renderTools(snap: Snapshot): HTMLElement {
  const table = el("table", "gw-table");
  const head = el("tr");
  for (const h of ["Tool", "Status", "Version", "Path"]) {
    head.append(el("th", undefined, h));
  }
  table.append(el("thead").appendChild(head).parentElement!);

  const body = el("tbody");
  for (const t of snap.tools) {
    const row = el("tr");
    row.append(el("td", "gw-mono", t.name));
    const status = el("td");
    status.append(
      el("span", t.found ? "gw-badge gw-ok" : "gw-badge gw-miss", t.found ? "found" : "missing"),
    );
    row.append(status);
    row.append(el("td", "gw-mono", t.version || "—"));
    row.append(el("td", "gw-path", t.path || "—"));
    body.append(row);
  }
  table.append(body);
