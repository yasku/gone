package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type activeTab int

const (
	tabUninstall activeTab = iota
	tabMonitor
)

type AppModel struct {
	active    activeTab
	uninstall UninstallModel
	monitor   MonitorModel
	styles    Styles
	width     int
	height    int
	ready     bool
	showHelp  bool
}

func NewApp() AppModel {
	return AppModel{
		active:    tabUninstall,
		uninstall: NewUninstallModel(),
		monitor:   NewMonitorModel(),
		styles:    DefaultStyles(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.uninstall.Init(),
		m.monitor.Init(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if msg.String() == "tab" {
			if m.active == tabUninstall {
				m.active = tabMonitor
			} else {
				m.active = tabUninstall
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := msg.Height - 4 // tab bar + padding
		m.uninstall = m.uninstall.SetSize(msg.Width, contentHeight)
		m.monitor = m.monitor.SetSize(msg.Width, contentHeight)
		m.ready = true
	}

	// Always route refresh ticks to monitor (prevents freeze)
	if _, ok := msg.(refreshMsg); ok {
		var cmd tea.Cmd
		m.monitor, cmd = m.monitor.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Route other messages to active tab
	switch m.active {
	case tabUninstall:
		// Don't re-route refreshMsg
		if _, ok := msg.(refreshMsg); !ok {
			var cmd tea.Cmd
			m.uninstall, cmd = m.uninstall.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tabMonitor:
		// Don't re-route refreshMsg (already handled above)
		if _, ok := msg.(refreshMsg); !ok {
			var cmd tea.Cmd
			m.monitor, cmd = m.monitor.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() tea.View {
	if !m.ready {
		v := tea.NewView("loading...")
		v.AltScreen = true
		return v
	}

	if m.showHelp {
		ghost := "" +
			"      ▄▄████████▄▄\n" +
			"    ██              ██\n" +
			"    ██  ██      ██  ██\n" +
			"    ██  ██      ██  ██\n" +
			"    ██              ██\n" +
			"    ██              ██\n" +
			"    ▀█▄██▄██▄██▄██▄█▀\n" +
			"      ▀  ▀  ▀  ▀  ▀\n"
		ghostArt := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Render(ghost)
		help := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("245")).
			Padding(1, 3).
			Width(50).
			Render(
				ghostArt + "\n" +
					"  g o n e — keybindings\n\n" +
					"  Tab       Switch tabs\n" +
					"  /         Filter list (Uninstall) / processes (Monitor)\n" +
					"  Esc       Exit filter / back to search\n" +
					"  Space     Toggle selection\n" +
					"  Enter     Search (input) / Trash (list)\n" +
					"  ?         Toggle help\n" +
					"  Ctrl+C    Quit\n\n" +
					"  hunt. select. trash.\n\n" +
					"         x AI & DATA Labs.",
			)
		overlay := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, help)
		v := tea.NewView(overlay)
		v.AltScreen = true
		return v
	}

	var b strings.Builder

	// Tab bar
	uninstallTab := " Uninstall "
	monitorTab := " Monitor "
	if m.active == tabUninstall {
		uninstallTab = m.styles.TabActive.Render(uninstallTab)
		monitorTab = m.styles.TabInactive.Render(monitorTab)
	} else {
		uninstallTab = m.styles.TabInactive.Render(uninstallTab)
		monitorTab = m.styles.TabActive.Render(monitorTab)
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, uninstallTab, monitorTab)
	tabLine := lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width - 4).
		Render(tabBar)
	b.WriteString(tabLine)
	b.WriteString("\n")

	// Content
	switch m.active {
	case tabUninstall:
		b.WriteString(m.uninstall.View())
	case tabMonitor:
		b.WriteString(m.monitor.View())
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
