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
	active     activeTab
	uninstall  UninstallModel
	monitor    MonitorModel
	splash     SplashModel
	styles     Styles
	width      int
	height     int
	ready      bool
	showSplash bool
	showHelp   bool
}

func NewApp(initialSearch string) AppModel {
	return AppModel{
		active:     tabUninstall,
		uninstall:  NewUninstall(initialSearch),
		monitor:    NewMonitorModel(),
		splash:     NewSplashModel(),
		styles:     DefaultStyles(),
		showSplash: true,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.splash.Init(),
		m.uninstall.Init(),
		m.monitor.Init(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case splashDoneMsg:
		m.showSplash = false
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			return m, nil // swallow all keys while help overlay is active
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

	// Forward all messages to splash while it is active
	if m.showSplash {
		var cmd tea.Cmd
		m.splash, cmd = m.splash.Update(msg)
		cmds = append(cmds, cmd)
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

	if m.showSplash {
		v := tea.NewView(m.splash.View())
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
					"  Esc       Back to search (from list)\n" +
					"  Esc       Quit (from search bar)\n" +
					"  x         Kill process (Monitor)\n" +
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

	// Header bar: tabs (left) + GONE branding (right)
	var uninstallTab, monitorTab string
	activeTabSt := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BCD4")).Padding(0, 1)
	inactiveTabSt := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)
	if m.active == tabUninstall {
		uninstallTab = activeTabSt.Render("◉ Uninstall")
		monitorTab = inactiveTabSt.Render("○ Monitor")
	} else {
		uninstallTab = inactiveTabSt.Render("○ Uninstall")
		monitorTab = activeTabSt.Render("◉ Monitor")
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Center, uninstallTab, "  ", monitorTab)

	goneLogo := gradientText("G O N E")
	tagline := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true).Render("hunt. select. trash.")
	brand := goneLogo + "  " + tagline

	// Calculate spacer to push brand to the right
	tabsW := lipgloss.Width(tabs)
	brandW := lipgloss.Width(brand)
	contentW := m.width - 4 // 2 padding + 2 margin
	spacerW := contentW - tabsW - brandW
	if spacerW < 1 {
		spacerW = 1
	}
	headerContent := tabs + strings.Repeat(" ", spacerW) + brand

	header := lipgloss.NewStyle().
		Width(m.width - 2).
		PaddingLeft(1).
		PaddingRight(1).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
		Render(headerContent)

	b.WriteString(header)
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
