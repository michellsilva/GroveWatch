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
	}
	sub := os.Args[1]
	args := os.Args[2:]

	var err error
	switch sub {
	case "scan":
		err = runScan(args)
	case "diff":
		err = runDiff(args)
	case "verify":
		err = runVerify(args)
	case "version", "--version", "-version":
		fmt.Printf("grovewatch %s\n", version)
	case "help", "-h", "--help":
		usage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "grovewatch: unknown command %q\n\n", sub)
		usage(os.Stderr)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "grovewatch: %v\n", err)
		os.Exit(1)
	}
}

func usage(w *os.File) {
	fmt.Fprint(w, `grovewatch — software provenance sentinel

Usage:
  grovewatch scan   [flags] <workspace>   Scan a workspace, emit provenance JSON
  grovewatch diff   <old.json> <new.json> Compare two snapshots, explain drift
  grovewatch verify <snapshot.json>       Verify a snapshot's digest integrity
  grovewatch version                      Print the version

Scan flags:
  -out string      Write snapshot to file instead of stdout
  -ignore string   Comma-separated path segments to skip
                   (default ".git,node_modules,dist,build,.grovewatch")
  -tools string    Comma-separated tool names to probe on PATH
                   (default "go,node,python,git")
  -env string      Comma-separated environment variables to record
                   (default "CI,GOOS,GOARCH,NODE_ENV")
`)
}

// runScan implements the "scan" subcommand.
func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	out := fs.String("out", "", "write snapshot to file instead of stdout")
	ignore := fs.String("ignore", ".git,node_modules,dist,build,.grovewatch", "comma-separated path segments to skip")
	tools := fs.String("tools", "go,node,python,git", "comma-separated tool names to probe")
	env := fs.String("env", "CI,GOOS,GOARCH,NODE_ENV", "comma-separated environment variables to record")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("scan requires exactly one <workspace> argument")
	}

	snap, err := provenance.Scan(provenance.ScanConfig{
		Root:        fs.Arg(0),
