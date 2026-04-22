package remover_test

import (
	"encoding/json"
	"fmt"
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

func TestAppendLogSetsTimestamp(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	entry := remover.LogEntry{Path: "/foo/bar", Kind: "file", SearchTerm: "bar"}
	if err := remover.AppendLog(entry); err != nil {
		t.Fatal(err)
	}

	logFile := filepath.Join(tmp, ".config", "gone", "operations.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	var parsed remover.LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
	if parsed.Op != "trash" {
		t.Errorf("expected op=trash, got %s", parsed.Op)
	}
}

func TestAppendLogMultipleEntries(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	for i := 0; i < 3; i++ {
		entry := remover.LogEntry{Path: filepath.Join("/tmp", fmt.Sprintf("file%d", i)), Kind: "file", SearchTerm: "test"}
		if err := remover.AppendLog(entry); err != nil {
			t.Fatal(err)
		}
	}

	logFile := filepath.Join(tmp, ".config", "gone", "operations.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(lines))
	}
}

func TestAppendLogCreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", os.Getenv("HOME"))

	// Ensure the config dir does not pre-exist
	configDir := filepath.Join(tmp, ".config", "gone")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("expected config dir to not exist yet")
	}

	entry := remover.LogEntry{Path: "/x", Kind: "file", SearchTerm: "x"}
	if err := remover.AppendLog(entry); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(configDir); err != nil {
		t.Errorf("expected config dir to be created, got: %v", err)
	}
}
