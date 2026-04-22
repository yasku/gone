package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var RCFiles = []string{
	".zshrc", ".zshenv", ".zprofile", ".zlogout",
	".bashrc", ".bash_profile", ".profile", ".bash_logout",
	".config/fish/config.fish",
}

// completionDirs are directories whose files are scanned for RC-style lines.
var completionDirs = []string{
	".oh-my-zsh/completions",
	".oh-my-zsh/custom/plugins",
	".bash_completion.d",
	".local/share/bash-completion/completions",
	".config/fish/completions",
	".config/fish/conf.d",
}

type RCMatch struct {
	File    string
	LineNum int
	Line    string
}

func SearchRC(term string) []RCMatch {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	lower := strings.ToLower(term)
	var matches []RCMatch

	for _, name := range RCFiles {
		matches = appendRCMatches(matches, filepath.Join(home, name), lower)
	}
	matches = append(matches, scanCompletionDirs(home, lower)...)
	return matches
}

func scanCompletionDirs(home, lower string) []RCMatch {
	var matches []RCMatch
	for _, dir := range completionDirs {
		full := filepath.Join(home, dir)
		entries, err := os.ReadDir(full)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			matches = appendRCMatches(matches, filepath.Join(full, e.Name()), lower)
		}
	}
	return matches
}

func appendRCMatches(matches []RCMatch, path, lower string) []RCMatch {
	f, err := os.Open(path)
	if err != nil {
		return matches
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		if strings.Contains(strings.ToLower(sc.Text()), lower) {
			matches = append(matches, RCMatch{
				File:    path,
				LineNum: lineNum,
				Line:    sc.Text(),
			})
		}
	}
	return matches
}
