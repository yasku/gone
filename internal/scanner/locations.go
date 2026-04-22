package scanner

import (
	"os"
	"path/filepath"
)

// GetScanPaths returns the ordered list of filesystem roots that gone scans.
// The list is computed at call time so that HOME is resolved correctly,
// including inside tests that override the environment variable.
func GetScanPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return []string{
		home,
		filepath.Join(home, "Library"),
		filepath.Join(home, ".config"),
		filepath.Join(home, ".local"),
		"/usr/local",
		"/opt/homebrew",
		"/opt",
	}
}

// ScanPaths is kept for backward compatibility; prefer GetScanPaths().
var ScanPaths = GetScanPaths()

// SkipDirs is the set of directory names that are never descended into during
// a scan (e.g. node_modules, .git, Caches). Exposed so tests can verify the
// skip behaviour without walking the real filesystem.
var SkipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	".Trash":       true,
	"Caches":       true,
	"DerivedData":  true,
	"CachedData":   true,
	".npm":         true,
	"vendor":       true,
	"__pycache__":  true,
	"cache":        true,
	"Cache":        true,
}
