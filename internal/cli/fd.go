package cli

import (
	"encoding/json"
	"fmt"
	"time"
)

type FDRunner struct {
	runner  *Runner
	tool    string
	hasJSON *bool
}

func NewFDRunner() *FDRunner {
	return &FDRunner{
		runner: NewRunner(30 * time.Second),
		tool:   "fd",
	}
}

func (f *FDRunner) IsAvailable() bool {
	return IsAvailable("fd")
}

func (f *FDRunner) HasJSONOutput() bool {
	if f.hasJSON != nil {
		return *f.hasJSON
	}

	supported := false
	r := NewRunner(5 * time.Second)

	_, err := r.ExecSimple("fd", []string{"--json", ".", "/tmp"})
	if err == nil {
		supported = true
	}

	f.hasJSON = &supported

	return supported
}

type FDMatch struct {
	Path        string `json:"path"`
	DisplayPath string `json:"display"`
	Name        string `json:"name"`
	Kind        string `json:"kind,omitempty"`
	IsDir       bool   `json:"is_dir,omitempty"`
	IsSymlink   bool   `json:"is_symlink,omitempty"`
	Size        int64  `json:"size,omitempty"`
	MTime       string `json:"mtime,omitempty"`
}

func (f *FDRunner) Search(term, root string) ([]FDMatch, error) {
	if !f.IsAvailable() {
		return nil, fmt.Errorf("fd not available")
	}

	if !f.HasJSONOutput() {
		return nil, fmt.Errorf("fd --json not supported")
	}

	args := []string{
		"--json",
		"--type", "f",
		"--type", "d",
		"--type", "l",
		"--search-term", term,
		root,
	}

	var matches []FDMatch
	err := f.runner.ExecStream("fd", args, func(line []byte) bool {
		var m FDMatch
		if err := json.Unmarshal(line, &m); err != nil {
			return true
		}

		if m.IsDir {
			m.Kind = "dir"
		} else if m.IsSymlink {
			m.Kind = "symlink"
		} else {
			m.Kind = "file"
		}

		matches = append(matches, m)
		return true
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (f *FDRunner) SearchDirs(term, root string) ([]FDMatch, error) {
	if !f.IsAvailable() {
		return nil, fmt.Errorf("fd not available")
	}

	if !f.HasJSONOutput() {
		return nil, fmt.Errorf("fd --json not supported")
	}

	args := []string{
		"--json",
		"--type", "d",
		"--search-term", term,
		root,
	}

	var matches []FDMatch
	err := f.runner.ExecStream("fd", args, func(line []byte) bool {
		var m FDMatch
		if err := json.Unmarshal(line, &m); err != nil {
			return true
		}
		m.Kind = "dir"
		matches = append(matches, m)
		return true
	})

	if err != nil {
		return nil, err
	}

	return matches, nil
}
