package provenance

import (
	"testing"
)

func mkSnap(files []FileRecord, tools []ToolRecord, env []EnvRecord) *Snapshot {
	s := &Snapshot{
		Schema:      SchemaVersion,
		Tool:        ToolInfo{Name: "grovewatch", Version: "test"},
		Files:       files,
		Tools:       tools,
