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
		{Path: "added.txt", Size: 9, Mode: "0644", Digest: "a1"},
	}, nil, nil)

	d := Compare(oldSnap, newSnap)
	if d.Identical {
		t.Fatal("expected drift")
	}

	got := map[string]ChangeKind{}
	for _, c := range d.Changes {
		if c.Category != FileCategory {
			t.Errorf("unexpected category %s", c.Category)
		}
		got[c.Name] = c.Kind
	}
	checks := map[string]ChangeKind{
		"added.txt":   Added,
		"changed.txt": Modified,
		"gone.txt":    Removed,
	}
	for name, kind := range checks {
		if got[name] != kind {
			t.Errorf("%s: got %q, want %q", name, got[name], kind)
		}
	}
	if _, ok := got["keep.txt"]; ok {
		t.Error("unchanged file should not appear in diff")
	}
}

func TestCompareToolAndEnvDrift(t *testing.T) {
	oldSnap := mkSnap(nil,
		[]ToolRecord{{Name: "go", Found: true, Version: "1.24"}},
		[]EnvRecord{{Key: "CI", Set: true, ValueDigest: "x1"}},
	)
	newSnap := mkSnap(nil,
		[]ToolRecord{{Name: "go", Found: true, Version: "1.25"}},
		[]EnvRecord{{Key: "CI", Set: true, ValueDigest: "x2"}},
	)
	d := Compare(oldSnap, newSnap)

	var sawTool, sawEnv bool
	for _, c := range d.Changes {
		if c.Category == ToolCategory && c.Name == "go" && c.Kind == Modified {
			sawTool = true
		}
		if c.Category == EnvCategory && c.Name == "CI" && c.Kind == Modified {
			sawEnv = true
		}
	}
	if !sawTool {
		t.Error("expected go version drift")
	}
	if !sawEnv {
		t.Error("expected CI env drift")
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	snap := mkSnap(
		[]FileRecord{{Path: "a.txt", Size: 1, Mode: "0644", Digest: "aa"}},
		[]ToolRecord{{Name: "go", Found: true, Version: "1.24"}},
		[]EnvRecord{{Key: "CI", Set: true, ValueDigest: "x1"}},
	)
	data, err := Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := Unmarshal(data)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !Verify(got) {
		t.Fatal("round-tripped snapshot failed verification")
	}
	if got.Digest != snap.Digest {
		t.Fatalf("digest changed on round trip: %s != %s", got.Digest, snap.Digest)
	}
}

func TestVerifyDetectsTampering(t *testing.T) {
	snap := mkSnap([]FileRecord{{Path: "a.txt", Size: 1, Mode: "0644", Digest: "aa"}}, nil, nil)
	// Tamper with content after digest was computed.
	snap.Files[0].Digest = "bb"
	if Verify(snap) {
		t.Fatal("verify should fail after tampering")
	}
}

<!-- draft note 1484 -->
