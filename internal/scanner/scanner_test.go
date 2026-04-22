package scanner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gone/internal/scanner"
)

func TestSearchFindsMatchingFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "foo-app", "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "foo-app", "foo-main.bin"), []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "unrelated.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := scanner.Search("foo", []string{tmp})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 matches (dir + file), got %d", len(results))
	}
	found := map[string]bool{}
	for _, r := range results {
		found[filepath.Base(r.Path)] = true
	}
	if !found["foo-app"] {
		t.Error("expected to find foo-app dir")
	}
}

func TestSearchSkipsIgnoredDirs(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "node_modules", "foo-pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "foo-real.txt"), []byte("real"), 0o644); err != nil {
		t.Fatal(err)
	}

	results, err := scanner.Search("foo", []string{tmp})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if filepath.Base(r.Path) == "foo-pkg" {
			t.Error("should not match inside node_modules")
		}
	}
}

func TestSearchConcurrentSafety(t *testing.T) {
	tmp := t.TempDir()
	// Create enough nested dirs to trigger fastwalk's parallel walking
	for i := 0; i < 50; i++ {
		dir := filepath.Join(tmp, "dir"+string(rune('a'+i%26)), "sub"+string(rune('0'+i%10)))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 20; j++ {
			if err := os.WriteFile(filepath.Join(dir, "match-"+string(rune('a'+j%26))+".txt"), []byte("data"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	for k := 0; k < 5; k++ {
		results, err := scanner.Search("match", []string{tmp})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) == 0 {
			t.Error("expected results")
		}
	}
}

func TestSearchDeduplicatesOverlappingPaths(t *testing.T) {
	tmp := t.TempDir()
	sub := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "target.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Search with overlapping paths: tmp includes sub
	results, err := scanner.Search("target", []string{tmp, sub})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, r := range results {
		if filepath.Base(r.Path) == "target.txt" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 match for target.txt (deduplication), got %d", count)
	}
}

func TestDirSizeCountsAllFiles(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "a.txt"), make([]byte, 1024), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), make([]byte, 512), 0o644); err != nil {
		t.Fatal(err)
	}

	got := scanner.DirSize(tmp)
	if got < 1536 {
		t.Errorf("DirSize = %d, want >= 1536", got)
	}
}

func TestDirSizeEmptyDir(t *testing.T) {
	tmp := t.TempDir()
	if got := scanner.DirSize(tmp); got != 0 {
		t.Errorf("DirSize(empty) = %d, want 0", got)
	}
}

func TestSearchStreamDrainsAndCloses(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"myapp-bin", "myapp-config", "other.txt"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{}
	defer func() { scanner.RCFiles = origFiles }()

	ch := scanner.SearchStream("myapp", []string{tmp})
	var matches []scanner.Match
	for m := range ch {
		matches = append(matches, m)
	}
	if len(matches) < 2 {
		t.Errorf("expected ≥2 matches, got %d", len(matches))
	}
	for _, m := range matches {
		if !strings.Contains(strings.ToLower(filepath.Base(m.Path)), "myapp") {
			t.Errorf("unexpected match path: %s", m.Path)
		}
	}
}

func TestSearchStreamEmitsRCLines(t *testing.T) {
	tmp := t.TempDir()
	rc := filepath.Join(tmp, ".zshrc")
	if err := os.WriteFile(rc, []byte("export PATH=/opt/myapp/bin:$PATH\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	origFiles := scanner.RCFiles
	scanner.RCFiles = []string{".zshrc"}
	t.Setenv("HOME", tmp)
	defer func() { scanner.RCFiles = origFiles }()

	ch := scanner.SearchStream("myapp", []string{})
	var rcMatches int
	for m := range ch {
		if m.Kind == "rc-line" {
			rcMatches++
		}
	}
	if rcMatches == 0 {
		t.Error("expected at least one rc-line match")
	}
}

func TestGetScanPathsContainsHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	paths := scanner.GetScanPaths()
	found := false
	for _, p := range paths {
		if p == tmp {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetScanPaths() does not contain HOME=%s; got %v", tmp, paths)
	}
}
