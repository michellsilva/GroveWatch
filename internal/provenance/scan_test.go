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
