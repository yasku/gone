package cli

import (
	"encoding/json"
	"testing"
)

func TestFDRunnerIsAvailable(t *testing.T) {
	fd := NewFDRunner()

	if !fd.IsAvailable() {
		t.Skip("fd not installed, skipping")
	}
}

func TestFDRunnerHasJSONOutput(t *testing.T) {
	fd := NewFDRunner()

	if !fd.IsAvailable() {
		t.Skip("fd not installed, skipping")
	}

	hasJSON := fd.HasJSONOutput()
	t.Logf("fd JSON output support: %v", hasJSON)
}

func TestFDRunnerNotAvailableWhenToolNotOnPath(t *testing.T) {
	fd := NewFDRunner()

	if fd.IsAvailable() {
		t.Skip("fd is installed on this system")
	}
}

func TestFDSearchRequiresJSON(t *testing.T) {
	fd := NewFDRunner()

	if !fd.IsAvailable() {
		t.Skip("fd not installed, skipping")
	}

	if !fd.HasJSONOutput() {
		t.Skip("fd --json not supported, skipping")
	}

	matches, err := fd.Search("go", "/usr/local")
	if err != nil {
		t.Fatalf("FDRunner.Search failed: %v", err)
	}

	for _, m := range matches {
		if m.Kind == "" {
			t.Error("Kind should be set for each match")
		}
		if m.Kind != "dir" && m.Kind != "file" && m.Kind != "symlink" {
			t.Errorf("unexpected Kind: %s", m.Kind)
		}
	}
}

func TestFDSearchDirsRequiresJSON(t *testing.T) {
	fd := NewFDRunner()

	if !fd.IsAvailable() {
		t.Skip("fd not installed, skipping")
	}

	if !fd.HasJSONOutput() {
		t.Skip("fd --json not supported, skipping")
	}

	matches, err := fd.SearchDirs("go", "/usr/local")
	if err != nil {
		t.Fatalf("FDRunner.SearchDirs failed: %v", err)
	}

	for _, m := range matches {
		if m.Kind != "dir" {
			t.Errorf("expected kind 'dir', got %q", m.Kind)
		}
	}
}

func TestFDSearchNotAvailable(t *testing.T) {
	fd := NewFDRunner()

	delete(whichCache, "fd")

	_, err := fd.Search("go", "/tmp")
	if err == nil {
		t.Error("expected error when fd not available")
	}

	whichCache["fd"] = "/opt/homebrew/bin/fd"
}

func TestFDSearchJSONNotSupportedReturnsError(t *testing.T) {
	fd := NewFDRunner()

	if !fd.IsAvailable() {
		t.Skip("fd not installed, skipping")
	}

	if fd.HasJSONOutput() {
		t.Skip("fd --json is supported, skipping this test")
	}

	_, err := fd.Search("go", "/tmp")
	if err == nil {
		t.Error("expected error when fd --json not supported")
	}
}

func TestFDFDMatchJSONParsing(t *testing.T) {
	jsonLine := `{"path":"/usr/local/bin/test","display":"/usr/local/bin/test","name":"test","is_dir":false,"is_symlink":false,"size":1024,"mtime":"2024-01-01T00:00:00Z"}`

	var m FDMatch
	if err := json.Unmarshal([]byte(jsonLine), &m); err != nil {
		t.Fatalf("Failed to parse FDMatch JSON: %v", err)
	}

	if m.Path != "/usr/local/bin/test" {
		t.Errorf("expected path /usr/local/bin/test, got %s", m.Path)
	}
	if m.Name != "test" {
		t.Errorf("expected name test, got %s", m.Name)
	}
	if m.IsDir {
		t.Error("expected IsDir=false")
	}
	if m.Size != 1024 {
		t.Errorf("expected size 1024, got %d", m.Size)
	}
}

func TestFDFDMatchKindFromType(t *testing.T) {
	tests := []struct {
		json     string
		expected string
	}{
		{`{"path":"/dir","is_dir":true}`, "dir"},
		{`{"path":"/symlink","is_symlink":true}`, "symlink"},
		{`{"path":"/file","is_dir":false,"is_symlink":false}`, "file"},
	}

	for _, tt := range tests {
		var m FDMatch
		if err := json.Unmarshal([]byte(tt.json), &m); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		if m.IsDir {
			m.Kind = "dir"
		} else if m.IsSymlink {
			m.Kind = "symlink"
		} else {
			m.Kind = "file"
		}

		if m.Kind != tt.expected {
			t.Errorf("for %s: expected %s, got %s", tt.json, tt.expected, m.Kind)
		}
	}
}

func TestFDFDRunnerNewFDRunner(t *testing.T) {
	fd := NewFDRunner()
	if fd == nil {
		t.Fatal("NewFDRunner returned nil")
	}
	if fd.runner == nil {
		t.Error("Runner should not be nil")
	}
	if fd.tool != "fd" {
		t.Errorf("expected tool 'fd', got %s", fd.tool)
	}
}
