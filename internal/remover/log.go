package remover

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LogEntry struct {
	Timestamp  string `json:"ts"`
	Op         string `json:"op"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Kind       string `json:"kind"`
	SearchTerm string `json:"term"`
}

func logPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	if home == "" {
		return "", fmt.Errorf("cannot resolve home directory: os.UserHomeDir failed and $HOME is unset")
	}
	return filepath.Join(home, ".config", "gone", "operations.log"), nil
}

func AppendLog(entry LogEntry) error {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	entry.Op = "trash"

	p, err := logPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}
