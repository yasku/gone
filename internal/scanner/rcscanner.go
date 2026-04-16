package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var RCFiles = []string{
	".zshrc", ".zshenv", ".zprofile",
	".bashrc", ".bash_profile", ".profile",
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
		path := filepath.Join(home, name)
		f, err := os.Open(path)
		if err != nil {
			continue
		}
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
		f.Close()
	}
	return matches
}
