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
