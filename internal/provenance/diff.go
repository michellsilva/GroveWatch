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
