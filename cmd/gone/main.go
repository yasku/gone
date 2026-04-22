package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"gone/internal/tui"
)

func main() {
	initialSearch := strings.Join(os.Args[1:], " ")
	p := tea.NewProgram(tui.NewApp(initialSearch))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
