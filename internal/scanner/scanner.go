package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charlievieth/fastwalk"
)

type Match struct {
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
	Kind    string // "file", "dir", "symlink", "rc-line"
}

func Search(term string, paths []string) ([]Match, error) {
	lower := strings.ToLower(term)
	var mu sync.Mutex
	var results []Match
	seen := make(map[string]bool)

	for _, root := range paths {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		_ = fastwalk.Walk(nil, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			name := d.Name()
			if d.IsDir() && SkipDirs[name] {
				return filepath.SkipDir
			}
			if !strings.Contains(strings.ToLower(name), lower) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			kind := "file"
			if d.IsDir() {
				kind = "dir"
			}
			if d.Type()&fs.ModeSymlink != 0 {
				kind = "symlink"
			}
			mu.Lock()
			if !seen[path] {
				seen[path] = true
				results = append(results, Match{
					Path:    path,
					IsDir:   d.IsDir(),
					Size:    info.Size(),
					ModTime: info.ModTime(),
					Kind:    kind,
				})
			}
			mu.Unlock()
			return nil
		})
	}
	return results, nil
}

func DirSize(path string) int64 {
	var total int64
	filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}
