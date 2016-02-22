package provenance

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ScanConfig controls how a workspace is scanned.
type ScanConfig struct {
	// Root is the directory to scan.
	Root string
	// ToolVersion is the grovewatch version stamped into the snapshot.
	ToolVersion string
	// Ignore holds path fragments; any file whose relative path contains one
	// of these fragments is skipped. Matching is done on slash-separated
	// segments (e.g. ".git", "node_modules", "dist").
	Ignore []string
	// Tools lists external tool names to probe on PATH (e.g. "go", "node").
	Tools []string
	// EnvKeys lists environment variable names whose presence and value hash
	// are recorded.
	EnvKeys []string
	// Now returns the snapshot timestamp; injectable for testing. If nil,
	// time.Now is used.
	Now func() time.Time
}

// versionRe extracts the first dotted version number from a tool's output.
var versionRe = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// Scan walks the configured workspace and returns a deterministic snapshot.
// The returned snapshot's Digest depends only on inputs, not on wall-clock
// time or filesystem traversal order.
func Scan(cfg ScanConfig) (*Snapshot, error) {
	if cfg.Root == "" {
		return nil, fmt.Errorf("scan: root must not be empty")
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("scan: resolve root: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("scan: stat root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scan: root %q is not a directory", root)
	}

	files, err := scanFiles(root, cfg.Ignore)
	if err != nil {
		return nil, err
	}
	tools := scanTools(cfg.Tools)
	env := scanEnv(cfg.EnvKeys)

	var totalBytes int64
	for _, f := range files {
		totalBytes += f.Size
	}

