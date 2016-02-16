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
