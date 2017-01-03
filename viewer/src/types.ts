/**
 * Type definitions for a grovewatch provenance snapshot. These mirror the Go
 * model in internal/provenance/model.go exactly; keep them in sync when the
 * schema version changes.
 */

export interface ToolInfo {
  name: string;
  version: string;
}

export interface WorkspaceInfo {
  root: string;
  file_count: number;
  total_bytes: number;
  merkle_root: string;
}

export interface FileRecord {
  path: string;
  size: number;
  mode: string;
  digest: string;
}

export interface ToolRecord {
  name: string;
  path: string;
  version: string;
  found: boolean;
}

