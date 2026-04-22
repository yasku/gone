package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"gone/internal/scanner"
)

func TestSearchRCCompletionDirs(t *testing.T) {
	tmp := t.TempDir()

	// Create a completion dir with a file containing a match
	compDir := filepath.Join(tmp, ".oh-my-zsh", "completions")
	if err := os.MkdirAll(compDir, 0o755); err != nil {
		t.Fatal(err)
	}
	compFile := filepath.Join(compDir, "_myapp")
	if err := os.WriteFile(compFile, []byte("#compdef myapp\n# completion for myapp\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{}
	t.Setenv("HOME", tmp)
	defer func() { scanner.RCFiles = origFiles }()

	matches := scanner.SearchRC("myapp")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches from completion dir, got %d: %+v", len(matches), matches)
	}
}

func TestSearchRCCaseInsensitive(t *testing.T) {
	tmp := t.TempDir()
	rc := filepath.Join(tmp, ".zshrc")
	if err := os.WriteFile(rc, []byte("export PATH=/opt/MyApp/bin\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{".zshrc"}
	t.Setenv("HOME", tmp)
	defer func() { scanner.RCFiles = origFiles }()

	matches := scanner.SearchRC("myapp") // lowercase term should match uppercase content
	if len(matches) != 1 {
		t.Fatalf("expected 1 case-insensitive match, got %d", len(matches))
	}
}

func TestSearchRCMultipleFiles(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{".zshrc", ".bashrc"} {
		path := filepath.Join(tmp, name)
		if err := os.WriteFile(path, []byte("source /opt/myapp/init.sh\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{".zshrc", ".bashrc"}
	t.Setenv("HOME", tmp)
	defer func() { scanner.RCFiles = origFiles }()

	matches := scanner.SearchRC("myapp")
	if len(matches) != 2 {
		t.Fatalf("expected 1 match per file (2 total), got %d", len(matches))
	}
}

func TestSearchRCNoMatchReturnsEmptySlice(t *testing.T) {
	tmp := t.TempDir()
	rc := filepath.Join(tmp, ".zshrc")
	if err := os.WriteFile(rc, []byte("export PATH=/usr/bin\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{".zshrc"}
	t.Setenv("HOME", tmp)
	defer func() { scanner.RCFiles = origFiles }()

	matches := scanner.SearchRC("nonexistent-tool-xyz")
	if len(matches) != 0 {
		t.Errorf("expected empty slice, got %d matches", len(matches))
	}
}

func TestSearchRCFindsMatchingLines(t *testing.T) {
	tmp := t.TempDir()
	rc := filepath.Join(tmp, ".zshrc")
	content := `export PATH="/usr/local/bin:$PATH"
export CLAUDE_API_KEY="sk-test"
alias ll="ls -la"
source ~/.claude/init.sh
`
	if err := os.WriteFile(rc, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{".zshrc"}
	t.Setenv("HOME", tmp)
	defer func() {
		scanner.RCFiles = origFiles
	}()

	matches := scanner.SearchRC("claude")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d: %+v", len(matches), matches)
	}
	if matches[0].LineNum != 2 {
		t.Errorf("expected line 2, got %d", matches[0].LineNum)
	}
	if matches[1].LineNum != 4 {
		t.Errorf("expected line 4, got %d", matches[1].LineNum)
	}
}
