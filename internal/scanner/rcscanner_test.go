package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"gone/internal/scanner"
)

func TestSearchRCFindsMatchingLines(t *testing.T) {
	tmp := t.TempDir()
	rc := filepath.Join(tmp, ".zshrc")
	content := `export PATH="/usr/local/bin:$PATH"
export CLAUDE_API_KEY="sk-test"
alias ll="ls -la"
source ~/.claude/init.sh
`
	os.WriteFile(rc, []byte(content), 0o644)

	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{".zshrc"}
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer func() {
		scanner.RCFiles = origFiles
		os.Setenv("HOME", origHome)
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
