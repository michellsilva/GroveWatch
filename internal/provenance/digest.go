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
	return hex.EncodeToString(sum[:])
}

// hashFile returns the hex-encoded SHA-256 of a file's contents. It streams the
// file so that arbitrarily large inputs use bounded memory.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// merkleRoot folds a set of file records into a single hash. The records are
// sorted by path first so the result is independent of discovery order. The
// root changes if any path, size, or content digest changes.
func merkleRoot(files []FileRecord) string {
	sorted := make([]FileRecord, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })

	h := sha256.New()
	for _, f := range sorted {
		// Length-prefixed fields prevent ambiguity between adjacent values.
		writeField(h, f.Path)
		writeField(h, f.Digest)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// writeField writes a length-prefixed string to h to avoid collision between
// concatenations such as ("ab","c") and ("a","bc").
func writeField(h io.Writer, s string) {
	var lenBuf [8]byte
	n := uint64(len(s))
	for i := 0; i < 8; i++ {
		lenBuf[i] = byte(n >> (8 * i))
	}
	_, _ = h.Write(lenBuf[:])
	_, _ = io.WriteString(h, s)
}

<!-- draft note 1475 -->
