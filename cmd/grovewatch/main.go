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
