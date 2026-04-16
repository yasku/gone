package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"gone/internal/tui"
)

type rootModel struct {
	uninstall tui.UninstallModel
	width     int
	height    int
	ready     bool
}

func (m rootModel) Init() tea.Cmd {
	return m.uninstall.Init()
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.uninstall = m.uninstall.SetSize(msg.Width, msg.Height)
		m.ready = true
	}

	var cmd tea.Cmd
	m.uninstall, cmd = m.uninstall.Update(msg)
	return m, cmd
}

func (m rootModel) View() tea.View {
	if !m.ready {
		v := tea.NewView("loading...")
		v.AltScreen = true
		return v
	}
	v := tea.NewView(m.uninstall.View())
	v.AltScreen = true
	return v
}

func main() {
	m := rootModel{uninstall: tui.NewUninstallModel()}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
