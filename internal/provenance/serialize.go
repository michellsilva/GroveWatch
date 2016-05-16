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
