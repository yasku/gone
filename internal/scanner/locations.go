package scanner

import (
	"os"
	"path/filepath"
)

// GetScanPaths returns the list of filesystem roots to search.
// Computed at call time so HOME is resolved correctly (including in tests).
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
