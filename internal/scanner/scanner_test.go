package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"gone/internal/scanner"
)

func TestSearchFindsMatchingFiles(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, "foo-app", "config"), 0o755)
	os.WriteFile(filepath.Join(tmp, "foo-app", "foo-main.bin"), []byte("binary"), 0o755)
	os.WriteFile(filepath.Join(tmp, "unrelated.txt"), []byte("nope"), 0o644)

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
	os.MkdirAll(filepath.Join(tmp, "node_modules", "foo-pkg"), 0o755)
	os.WriteFile(filepath.Join(tmp, "foo-real.txt"), []byte("real"), 0o644)

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
