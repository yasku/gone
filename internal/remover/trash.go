package remover

// MoveToTrash moves the file or directory at absPath to the macOS Trash using
// AppleScript via osascript. The path is passed as an argv argument rather than
// interpolated into the script string, so filenames containing double-quote
// characters are handled correctly.

import (
	"bytes"
	"fmt"
	"os/exec"
)

func MoveToTrash(absPath string) error {
	// Pass the path as an argv argument so no string interpolation into
	// AppleScript is needed.  AppleScript has no \"-style escape sequence;
	// embedding an arbitrary path directly in a string literal breaks on
	// filenames that contain a double-quote character (legal on APFS/HFS+).
	script := `on run argv
	tell application "Finder" to delete (POSIX file (item 1 of argv))
end run`
	cmd := exec.Command("/usr/bin/osascript", "-e", script, "--", absPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("trash %s: %w: %s", absPath, err, stderr.String())
	}
	return nil
}
