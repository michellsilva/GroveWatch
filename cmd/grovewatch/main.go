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
		ToolVersion: version,
		Ignore:      splitList(*ignore),
		Tools:       splitList(*tools),
		EnvKeys:     splitList(*env),
	})
	if err != nil {
		return err
	}

	data, err := provenance.Marshal(snap)
	if err != nil {
		return err
	}
	if *out == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", *out, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d files, digest %s)\n", *out, snap.Workspace.FileCount, snap.Digest[:12])
	return nil
}

// runDiff implements the "diff" subcommand.
func runDiff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "emit the diff as JSON instead of text")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("diff requires <old.json> <new.json>")
	}

	oldSnap, err := loadSnapshot(fs.Arg(0))
	if err != nil {
		return err
	}
	newSnap, err := loadSnapshot(fs.Arg(1))
	if err != nil {
		return err
	}

	d := provenance.Compare(oldSnap, newSnap)
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(d)
	}
	fmt.Print(d.Summary())
	if !d.Identical {
		// Non-zero exit signals drift, useful in CI gates.
		os.Exit(3)
	}
	return nil
}

// runVerify implements the "verify" subcommand.
func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("verify requires <snapshot.json>")
	}
	snap, err := loadSnapshot(fs.Arg(0))
	if err != nil {
		return err
	}
	if provenance.Verify(snap) {
		fmt.Printf("OK: digest %s matches content\n", snap.Digest[:12])
		return nil
	}
	recomputed := provenance.ComputeDigest(snap)
	fmt.Fprintf(os.Stderr, "MISMATCH: stored %s but content hashes to %s\n", snap.Digest[:12], recomputed[:12])
	os.Exit(4)
	return nil
}

// loadSnapshot reads and parses a snapshot file.
func loadSnapshot(path string) (*provenance.Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	snap, err := provenance.Read(f)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return snap, nil
}

// splitList splits a comma-separated flag value into trimmed, non-empty parts.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

<!-- draft note 1458 -->
