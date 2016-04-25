package provenance

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTree creates a small deterministic workspace for testing and returns its
// root. It includes a nested directory and an ignored directory.
func writeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"README.md":           "# demo\n",
		"src/main.go":         "package main\nfunc main() {}\n",
		"src/util/helper.go":  "package util\n",
		"node_modules/dep.js": "// should be ignored\n",
	}
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return root
}

func fixedNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

func TestScanRecordsFilesAndIgnores(t *testing.T) {
	root := writeTree(t)
	snap, err := Scan(ScanConfig{
		Root:        root,
		ToolVersion: "test",
		Ignore:      []string{"node_modules"},
		Now:         fixedNow,
	})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if snap.Workspace.FileCount != 3 {
		t.Fatalf("expected 3 files, got %d: %+v", snap.Workspace.FileCount, snap.Files)
	}
	// Files must be sorted by path.
	want := []string{"README.md", "src/main.go", "src/util/helper.go"}
	for i, w := range want {
		if snap.Files[i].Path != w {
			t.Errorf("file[%d] = %q, want %q", i, snap.Files[i].Path, w)
		}
	}
	// Ignored file must not appear.
	for _, f := range snap.Files {
		if f.Path == "node_modules/dep.js" {
			t.Errorf("ignored file was recorded: %s", f.Path)
		}
	}
}

func TestScanDeterministicDigest(t *testing.T) {
	root := writeTree(t)
	cfg := ScanConfig{Root: root, ToolVersion: "test", Ignore: []string{"node_modules"}, Now: fixedNow}

	a, err := Scan(cfg)
	if err != nil {
		t.Fatalf("scan a: %v", err)
	}
	b, err := Scan(cfg)
	if err != nil {
		t.Fatalf("scan b: %v", err)
	}
	if a.Digest != b.Digest {
		t.Fatalf("digest not stable: %s != %s", a.Digest, b.Digest)
	}
	if a.Workspace.MerkleRoot != b.Workspace.MerkleRoot {
		t.Fatalf("merkle root not stable")
	}
	if !Verify(a) {
		t.Fatalf("scanned snapshot failed self-verification")
	}
}

