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

// Compare computes the drift between an old and new snapshot. The result is
// deterministic: changes are grouped by category and sorted by name.
func Compare(oldSnap, newSnap *Snapshot) *Diff {
	d := &Diff{
		OldDigest: oldSnap.Digest,
		NewDigest: newSnap.Digest,
		Identical: oldSnap.Digest == newSnap.Digest,
	}
	d.Changes = append(d.Changes, diffFiles(oldSnap.Files, newSnap.Files)...)
	d.Changes = append(d.Changes, diffTools(oldSnap.Tools, newSnap.Tools)...)
	d.Changes = append(d.Changes, diffEnv(oldSnap.Environment, newSnap.Environment)...)
	return d
}

func diffFiles(oldFiles, newFiles []FileRecord) []Change {
	oldByPath := make(map[string]FileRecord, len(oldFiles))
	for _, f := range oldFiles {
		oldByPath[f.Path] = f
	}
	newByPath := make(map[string]FileRecord, len(newFiles))
	for _, f := range newFiles {
		newByPath[f.Path] = f
	}

	var changes []Change
	for _, nf := range newFiles {
		of, ok := oldByPath[nf.Path]
		if !ok {
			changes = append(changes, Change{
				Category: FileCategory,
				Kind:     Added,
				Name:     nf.Path,
				Detail:   fmt.Sprintf("new file (%d bytes, digest %s)", nf.Size, short(nf.Digest)),
			})
			continue
		}
		if of.Digest != nf.Digest {
			changes = append(changes, Change{
				Category: FileCategory,
				Kind:     Modified,
				Name:     nf.Path,
				Detail:   fmt.Sprintf("content changed %s -> %s (%d -> %d bytes)", short(of.Digest), short(nf.Digest), of.Size, nf.Size),
			})
		} else if of.Mode != nf.Mode {
			changes = append(changes, Change{
				Category: FileCategory,
				Kind:     Modified,
				Name:     nf.Path,
				Detail:   fmt.Sprintf("mode changed %s -> %s", of.Mode, nf.Mode),
			})
		}
	}
	for _, of := range oldFiles {
		if _, ok := newByPath[of.Path]; !ok {
			changes = append(changes, Change{
				Category: FileCategory,
				Kind:     Removed,
				Name:     of.Path,
				Detail:   fmt.Sprintf("file removed (was %d bytes)", of.Size),
			})
		}
	}
	sortChanges(changes)
	return changes
}

func diffTools(oldTools, newTools []ToolRecord) []Change {
