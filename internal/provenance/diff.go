package provenance

import (
	"fmt"
	"sort"
	"strings"
)

// ChangeKind classifies a single drift entry.
type ChangeKind string

const (
	// Added means the item exists in the new snapshot but not the old.
	Added ChangeKind = "added"
	// Removed means the item existed in the old snapshot but not the new.
	Removed ChangeKind = "removed"
	// Modified means the item exists in both but its content differs.
	Modified ChangeKind = "modified"
)

// Category groups changes by the kind of input that drifted.
type Category string

const (
	FileCategory Category = "file"
	ToolCategory Category = "tool"
	EnvCategory  Category = "env"
)

// Change is a single explained drift entry.
type Change struct {
	Category Category   `json:"category"`
	Kind     ChangeKind `json:"kind"`
	// Name is the identifier of the changed item (file path, tool name, env key).
	Name string `json:"name"`
	// Detail is a human-readable explanation of what changed.
	Detail string `json:"detail"`
}

// Diff is the result of comparing two snapshots.
type Diff struct {
	// OldDigest and NewDigest are the canonical identities being compared.
	OldDigest string `json:"old_digest"`
	NewDigest string `json:"new_digest"`
	// Identical is true when the two snapshots share the same canonical digest.
	Identical bool `json:"identical"`
	// Changes is the ordered list of drift entries (files, then tools, then env).
	Changes []Change `json:"changes"`
}
