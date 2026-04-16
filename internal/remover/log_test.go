package remover_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gone/internal/remover"
)

func TestAppendLog(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	entry := remover.LogEntry{Path: "/test/file.txt", Size: 1024, Kind: "file", SearchTerm: "test"}
	if err := remover.AppendLog(entry); err != nil {
		t.Fatal(err)
	}

	logFile := filepath.Join(tmp, ".config", "gone", "operations.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.TrimSpace(string(data))
	var parsed remover.LogEntry
	if err := json.Unmarshal([]byte(lines), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Path != "/test/file.txt" {
		t.Errorf("expected /test/file.txt, got %s", parsed.Path)
	}
	if parsed.Op != "trash" {
		t.Errorf("expected op=trash, got %s", parsed.Op)
	}
}
