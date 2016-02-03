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

