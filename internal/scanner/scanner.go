// Package scanner walks the filesystem and shell RC files to find artifacts
// matching a given search term.
package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charlievieth/fastwalk"
)

// Match is a single filesystem or RC-file artifact found during a scan.
// Kind is one of "file", "dir", "symlink", or "rc-line".
type Match struct {
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
	Kind    string
}

// Search walks each root in paths and returns all entries whose name contains
// term (case-insensitive). Ignored directories (node_modules, .git, etc.) are
// skipped. Duplicate paths from overlapping roots are deduplicated.
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

// SearchStream returns a read-only channel that emits Match values as they are
// discovered. The channel is closed when the full scan (filesystem + RC files)
// is complete. Callers should drain the channel to avoid goroutine leaks.
func SearchStream(term string, paths []string) <-chan Match {
	ch := make(chan Match, 64)
	go func() {
		defer close(ch)
		var seen sync.Map
		lower := strings.ToLower(term)
		for _, root := range paths {
			if _, err := os.Lstat(root); err != nil {
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
				if _, loaded := seen.LoadOrStore(path, struct{}{}); loaded {
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
				ch <- Match{
					Path:    path,
					IsDir:   d.IsDir(),
					Size:    info.Size(),
					ModTime: info.ModTime(),
					Kind:    kind,
				}
				return nil
			})
		}
		for _, rc := range SearchRC(term) {
			m := Match{
				Path: fmt.Sprintf("%s:%d", rc.File, rc.LineNum),
				Kind: "rc-line",
				Size: int64(len(rc.Line)),
			}
			if _, loaded := seen.LoadOrStore(m.Path, struct{}{}); !loaded {
				ch <- m
			}
		}
	}()
	return ch
}

// DirSize returns the total byte size of all regular files under path.
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
