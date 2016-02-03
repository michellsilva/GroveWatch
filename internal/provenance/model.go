// Package provenance defines the data model for software provenance snapshots.
//
// A Snapshot captures a deterministic, reproducible fingerprint of a workspace:
// the files it contains, the tools available on the host, and the relevant
// environment inputs. Snapshots are designed to be byte-for-byte reproducible
// given the same inputs so that two snapshots can be compared to explain drift.
package provenance

import "time"

// SchemaVersion identifies the on-disk provenance format. Bump on breaking
// changes to the JSON structure so viewers can adapt.
const SchemaVersion = "1.0.0"

// Snapshot is the top-level provenance record for a workspace.
type Snapshot struct {
	// Schema is the provenance schema version (see SchemaVersion).
	Schema string `json:"schema"`
	// Tool identifies the producer and its version.
	Tool ToolInfo `json:"tool"`
	// Workspace describes the scanned root and aggregate metrics.
	Workspace WorkspaceInfo `json:"workspace"`
	// Files are the recorded file inputs, sorted by path for determinism.
