package provenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Marshal serializes a snapshot to indented, deterministic JSON. Field order is
// fixed by the struct definition and all collections are pre-sorted during the
// scan, so identical inputs produce byte-identical output (modulo CreatedAt).
func Marshal(snap *Snapshot) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return nil, fmt.Errorf("marshal snapshot: %w", err)
	}
	return buf.Bytes(), nil
}

// Unmarshal parses a snapshot from JSON.
func Unmarshal(data []byte) (*Snapshot, error) {
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return &snap, nil
}

// Read parses a snapshot from an io.Reader.
func Read(r io.Reader) (*Snapshot, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	return Unmarshal(data)
}

// canonicalView is a projection of a snapshot that excludes volatile fields
// (CreatedAt and the Digest itself). Hashing this view yields the stable
// identity used as Snapshot.Digest.
type canonicalView struct {
	Schema      string        `json:"schema"`
	Tool        ToolInfo      `json:"tool"`
	Workspace   WorkspaceInfo `json:"workspace"`
	Files       []FileRecord  `json:"files"`
	Tools       []ToolRecord  `json:"tools"`
	Environment []EnvRecord   `json:"environment"`
}

// ComputeDigest returns the canonical SHA-256 identity of a snapshot. It is
// stable across runs with identical inputs because it excludes CreatedAt and
// the previously computed Digest.
func ComputeDigest(snap *Snapshot) string {
	view := canonicalView{
		Schema:      snap.Schema,
		Tool:        snap.Tool,
		Workspace:   snap.Workspace,
		Files:       snap.Files,
		Tools:       snap.Tools,
		Environment: snap.Environment,
	}
	// json.Marshal on a fixed struct with sorted slices is deterministic.
	data, err := json.Marshal(view)
	if err != nil {
		// The view contains only serializable primitives; marshaling cannot
		// realistically fail. Fall back to an empty digest input rather than
		// panicking.
		data = []byte{}
	}
	return hashBytes(data)
}

// Verify recomputes the digest of snap and reports whether it matches the
// stored Digest. A false result means the snapshot was tampered with or
// produced by an incompatible version.
func Verify(snap *Snapshot) bool {
	return ComputeDigest(snap) == snap.Digest
}

<!-- draft note 1471 -->
