package provenance

import (
	"testing"
)

func mkSnap(files []FileRecord, tools []ToolRecord, env []EnvRecord) *Snapshot {
	s := &Snapshot{
		Schema:      SchemaVersion,
		Tool:        ToolInfo{Name: "grovewatch", Version: "test"},
		Files:       files,
		Tools:       tools,
		Environment: env,
	}
	s.Workspace.MerkleRoot = merkleRoot(files)
	s.Digest = ComputeDigest(s)
	return s
}

func TestCompareIdentical(t *testing.T) {
	files := []FileRecord{{Path: "a.txt", Size: 1, Mode: "0644", Digest: "aa"}}
	a := mkSnap(files, nil, nil)
	b := mkSnap(files, nil, nil)
	d := Compare(a, b)
	if !d.Identical {
		t.Fatalf("expected identical, got changes: %+v", d.Changes)
	}
