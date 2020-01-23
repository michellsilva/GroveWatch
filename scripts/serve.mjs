#!/usr/bin/env node
/**
 * A tiny, zero-dependency static file server for previewing the grovewatch
 * viewer locally. Uses only Node's standard library.
 *
 * Usage: node scripts/serve.mjs [port]
 * Then open http://localhost:<port>/viewer/public/index.html
 */

import { createServer } from "node:http";
import { readFile, stat } from "node:fs/promises";
import { extname, join, normalize, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = resolve(fileURLToPath(new URL("..", import.meta.url)));
const PORT = Number(process.argv[2] || process.env.PORT || 8080);

const MIME = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
