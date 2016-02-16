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
