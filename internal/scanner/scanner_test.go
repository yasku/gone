package scanner_test

import (
	"os"
	"path/filepath"
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
