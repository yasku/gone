package scanner

import (
	"os"
	"path/filepath"
)

var home = os.Getenv("HOME")

var ScanPaths = []string{
	home,
	filepath.Join(home, "Library"),
	filepath.Join(home, ".config"),
	filepath.Join(home, ".local"),
	"/usr/local",
	"/opt/homebrew",
	"/opt",
}

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
