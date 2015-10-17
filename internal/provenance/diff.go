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
	oldByName := make(map[string]ToolRecord, len(oldTools))
	for _, t := range oldTools {
		oldByName[t.Name] = t
	}
	newByName := make(map[string]ToolRecord, len(newTools))
	for _, t := range newTools {
		newByName[t.Name] = t
	}

	var changes []Change
	for _, nt := range newTools {
		ot, ok := oldByName[nt.Name]
		if !ok {
			changes = append(changes, Change{
				Category: ToolCategory,
				Kind:     Added,
				Name:     nt.Name,
				Detail:   fmt.Sprintf("tool now tracked (found=%v version=%q)", nt.Found, nt.Version),
			})
			continue
		}
		switch {
		case ot.Found && !nt.Found:
			changes = append(changes, Change{
				Category: ToolCategory, Kind: Removed, Name: nt.Name,
				Detail: "tool no longer available on PATH",
			})
		case !ot.Found && nt.Found:
			changes = append(changes, Change{
				Category: ToolCategory, Kind: Added, Name: nt.Name,
				Detail: fmt.Sprintf("tool became available (version %q)", nt.Version),
			})
		case ot.Version != nt.Version:
			changes = append(changes, Change{
				Category: ToolCategory, Kind: Modified, Name: nt.Name,
				Detail: fmt.Sprintf("version changed %q -> %q", ot.Version, nt.Version),
			})
		case ot.Path != nt.Path:
			changes = append(changes, Change{
				Category: ToolCategory, Kind: Modified, Name: nt.Name,
				Detail: fmt.Sprintf("path changed %q -> %q", ot.Path, nt.Path),
			})
		}
	}
	for _, ot := range oldTools {
		if _, ok := newByName[ot.Name]; !ok {
			changes = append(changes, Change{
				Category: ToolCategory, Kind: Removed, Name: ot.Name,
				Detail: "tool no longer tracked",
			})
		}
	}
	sortChanges(changes)
	return changes
}

func diffEnv(oldEnv, newEnv []EnvRecord) []Change {
	oldByKey := make(map[string]EnvRecord, len(oldEnv))
	for _, e := range oldEnv {
		oldByKey[e.Key] = e
	}
	newByKey := make(map[string]EnvRecord, len(newEnv))
	for _, e := range newEnv {
		newByKey[e.Key] = e
	}

	var changes []Change
	for _, ne := range newEnv {
		oe, ok := oldByKey[ne.Key]
		if !ok {
			changes = append(changes, Change{
				Category: EnvCategory, Kind: Added, Name: ne.Key,
				Detail: fmt.Sprintf("env now tracked (set=%v)", ne.Set),
			})
			continue
		}
		switch {
		case oe.Set && !ne.Set:
			changes = append(changes, Change{
				Category: EnvCategory, Kind: Removed, Name: ne.Key,
				Detail: "env variable unset",
			})
		case !oe.Set && ne.Set:
			changes = append(changes, Change{
				Category: EnvCategory, Kind: Added, Name: ne.Key,
				Detail: "env variable set",
			})
		case oe.ValueDigest != ne.ValueDigest:
			changes = append(changes, Change{
				Category: EnvCategory, Kind: Modified, Name: ne.Key,
				Detail: fmt.Sprintf("value changed %s -> %s", short(oe.ValueDigest), short(ne.ValueDigest)),
			})
		}
	}
	for _, oe := range oldEnv {
		if _, ok := newByKey[oe.Key]; !ok {
			changes = append(changes, Change{
				Category: EnvCategory, Kind: Removed, Name: oe.Key,
				Detail: "env no longer tracked",
			})
		}
	}
	sortChanges(changes)
	return changes
}

// sortChanges orders changes by name then kind for a stable, readable report.
func sortChanges(c []Change) {
	sort.SliceStable(c, func(i, j int) bool {
		if c[i].Name != c[j].Name {
			return c[i].Name < c[j].Name
		}
		return c[i].Kind < c[j].Kind
	})
}
