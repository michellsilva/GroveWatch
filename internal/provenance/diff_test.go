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
	if len(d.Changes) != 0 {
		t.Fatalf("expected no changes, got %d", len(d.Changes))
	}
}

func TestCompareFileDrift(t *testing.T) {
	oldSnap := mkSnap([]FileRecord{
		{Path: "keep.txt", Size: 1, Mode: "0644", Digest: "k1"},
		{Path: "changed.txt", Size: 2, Mode: "0644", Digest: "c1"},
		{Path: "gone.txt", Size: 3, Mode: "0644", Digest: "g1"},
	}, nil, nil)
	newSnap := mkSnap([]FileRecord{
		{Path: "keep.txt", Size: 1, Mode: "0644", Digest: "k1"},
		{Path: "changed.txt", Size: 5, Mode: "0644", Digest: "c2"},
