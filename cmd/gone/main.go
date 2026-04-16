package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"gone/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.NewApp())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
