package remover

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func MoveToTrash(absPath string) error {
	escaped := strings.ReplaceAll(absPath, `"`, `\"`)
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, escaped)
	cmd := exec.Command("/usr/bin/osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("trash %s: %w: %s", absPath, err, stderr.String())
	}
	return nil
}
