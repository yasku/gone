package cli

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
	whichCache = make(map[string]string)
	whichOnce  sync.Once
)

func Which(tool string) (string, error) {
	whichOnce.Do(func() {
		// Pre-populate common tools
		tools := []string{"fd", "osqueryi", "glances", "bmon", "mtr", "nmap"}
		for _, t := range tools {
			if path, err := exec.LookPath(t); err == nil {
				whichCache[t] = path
			}
		}
	})

	if path, ok := whichCache[tool]; ok {
		return path, nil
	}

	path, err := exec.LookPath(tool)
	if err != nil {
		return "", fmt.Errorf("which %s: not found", tool)
	}

	whichCache[tool] = path
	return path, nil
}

func IsAvailable(tool string) bool {
	_, err := Which(tool)
	return err == nil
}

func AvailableTools() []string {
	tools := []string{"fd", "osqueryi", "glances", "bmon", "mtr", "nmap"}
	var available []string
	for _, t := range tools {
		if IsAvailable(t) {
			available = append(available, t)
		}
	}
	return available
}
