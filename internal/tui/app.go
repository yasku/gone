package tui

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type activeTab int

const (
	tabUninstall activeTab = iota
	tabMonitor
)

// goneKeyMap defines all keybindings and satisfies the help.KeyMap interface.
// FullHelp is tab-aware: Monitor shows filter/kill/sort, Uninstall shows search/select/trash.
type goneKeyMap struct {
	active  activeTab
	Tab     key.Binding
	Help    key.Binding
	Quit    key.Binding
	Search  key.Binding
	Space   key.Binding
	Enter   key.Binding
	Escape  key.Binding
	Filter  key.Binding
	Kill    key.Binding
	NavUD   key.Binding
	SortNum key.Binding
}

func defaultKeyMap() goneKeyMap {
	return goneKeyMap{
		Tab:     key.NewBinding(key.WithKeys("tab"),                    key.WithHelp("tab",    "switch tabs")),
		Help:    key.NewBinding(key.WithKeys("?"),                       key.WithHelp("?",      "toggle help")),
		Quit:    key.NewBinding(key.WithKeys("ctrl+c"),                 key.WithHelp("ctrl+c", "quit")),
		Search:  key.NewBinding(key.WithKeys("enter"),                  key.WithHelp("enter",  "search")),
		Space:   key.NewBinding(key.WithKeys(" "),                      key.WithHelp("space",  "toggle selection")),
		Enter:   key.NewBinding(key.WithKeys("enter"),                  key.WithHelp("enter",  "trash selected")),
		Escape:  key.NewBinding(key.WithKeys("esc"),                    key.WithHelp("esc",    "back / quit")),
		Filter:  key.NewBinding(key.WithKeys("/"),                      key.WithHelp("/",      "filter processes")),
		Kill:    key.NewBinding(key.WithKeys("x"),                      key.WithHelp("x",      "kill process")),
		NavUD:   key.NewBinding(key.WithKeys("up", "k", "down", "j"),  key.WithHelp("↑/↓",   "navigate")),
		SortNum: key.NewBinding(key.WithKeys("1", "2", "3", "4"),      key.WithHelp("1-4",    "sort column")),
	}
}

func (k goneKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Help, k.Quit}
}

func (k goneKeyMap) FullHelp() [][]key.Binding {
	global := []key.Binding{k.Tab, k.Help, k.Quit}
	switch k.active {
	case tabUninstall:
		return [][]key.Binding{global, {k.Search, k.Space, k.Enter, k.Escape}}
	case tabMonitor:
		return [][]key.Binding{global, {k.Filter, k.Kill, k.NavUD, k.SortNum}}
	}
	return [][]key.Binding{global}
}

type AppModel struct {
	active     activeTab
	uninstall  UninstallModel
	monitor    MonitorModel
	splash     SplashModel
	styles     Styles
	keys       goneKeyMap
	helpView   help.Model
	width      int
	height     int
	ready      bool
	showSplash bool
	showHelp   bool
}

func NewApp(initialSearch string) AppModel {
	hv := help.New()
	hv.ShowAll = true

	return AppModel{
		active:     tabUninstall,
		uninstall:  NewUninstall(initialSearch),
		monitor:    NewMonitorModel(),
		splash:     NewSplashModel(),
		styles:     DefaultStyles(),
		keys:       defaultKeyMap(),
		helpView:   hv,
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
			m.keys.active = m.active
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := msg.Height - 4 // tab bar + padding
		m.uninstall = m.uninstall.SetSize(msg.Width, contentHeight)
		m.monitor = m.monitor.SetSize(msg.Width, contentHeight)
		m.helpView.SetWidth(m.width - 16)
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

	// Always route scan stream messages to uninstall (scan may run while on Monitor tab)
	switch msg.(type) {
	case scanItemMsg, scanDoneMsg:
		var cmd tea.Cmd
		m.uninstall, cmd = m.uninstall.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
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

		keybindingsText := m.helpView.View(m.keys)

		helpBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("245")).
			Padding(1, 3).
			Width(50).
			Render(
				ghostArt + "\n" +
					gradientText("g o n e") + "  — keybindings\n\n" +
					keybindingsText + "\n\n" +
					lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true).
						Render("  hunt. select. trash.") + "\n\n" +
					lipgloss.NewStyle().Foreground(lipgloss.Color("245")).
						Render("         x AI & DATA Labs."),
			)
		overlay := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpBox)
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
