package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"sort"
)

// hashBytes returns the hex-encoded SHA-256 of b.
func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
