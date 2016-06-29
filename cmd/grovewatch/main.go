// Command grovewatch is a software provenance sentinel.
//
// It scans a workspace and emits a deterministic provenance snapshot capturing
// file, tool, and environment inputs. Two snapshots can be compared to explain
// exactly what drifted, and a snapshot's integrity can be verified against its
// self-describing digest.
//
// Usage:
//
//	grovewatch scan   [flags] <workspace>
//	grovewatch diff   <old.json> <new.json>
//	grovewatch verify <snapshot.json>
//	grovewatch version
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/michellsilva/GroveWatch/internal/provenance"
)

// version is the tool version stamped into snapshots. Overridable at build
// time via -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
