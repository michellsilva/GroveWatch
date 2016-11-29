/**
 * Entry point for the grovewatch browser viewer.
 *
 * It loads a provenance snapshot and renders it. The report source is resolved
 * in this order:
 *   1. a `?report=<url>` query parameter, if present;
 *   2. a `<script id="gw-report" type="application/json">` inline block, which
 *      lets the viewer work from the local filesystem without a server;
 *   3. the default `./provenance.json` fetched relative to the page.
 */

import type { Snapshot } from "./types.js";
import { render, renderError } from "./render.js";

const DEFAULT_REPORT_URL = "./provenance.json";

/** Reads an inline JSON report embedded in the page, if present. */
function readInlineReport(): Snapshot | null {
  const tag = document.getElementById("gw-report");
  if (!tag || !tag.textContent || tag.textContent.trim().length === 0) {
    return null;
  }
  return JSON.parse(tag.textContent) as Snapshot;
}

/** Fetches a report from a URL. */
async function fetchReport(url: string): Promise<Snapshot> {
  const res = await fetch(url, { cache: "no-store" });
