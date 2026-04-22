// Package tui implements the gone terminal UI using Bubble Tea v2, Lip Gloss
// v2, and Bubbles v2. AppModel is the root model that hosts the Uninstall and
// Monitor tabs, the splash screen, and the help overlay.
package tui

import (
	"fmt"
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

// goneKeyMap defines all keybindings for gone and satisfies the help.KeyMap
// interface. FullHelp is tab-aware: the Monitor tab shows filter/kill/sort
// bindings while the Uninstall tab shows search/select/trash bindings.
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
		Tab:     key.NewBinding(key.WithKeys("tab"),                   key.WithHelp("tab",    "switch tabs")),
		Help:    key.NewBinding(key.WithKeys("?"),                      key.WithHelp("?",      "toggle help")),
		Quit:    key.NewBinding(key.WithKeys("ctrl+c"),                key.WithHelp("ctrl+c", "quit")),
		Search:  key.NewBinding(key.WithKeys("enter"),                 key.WithHelp("enter",  "search")),
		Space:   key.NewBinding(key.WithKeys(" "),                     key.WithHelp("space",  "toggle selection")),
		Enter:   key.NewBinding(key.WithKeys("enter"),                 key.WithHelp("enter",  "trash selected")),
		Escape:  key.NewBinding(key.WithKeys("esc"),                   key.WithHelp("esc",    "back / quit")),
		Filter:  key.NewBinding(key.WithKeys("/"),                     key.WithHelp("/",      "filter processes")),
		Kill:    key.NewBinding(key.WithKeys("x"),                     key.WithHelp("x",      "kill process")),
		NavUD:   key.NewBinding(key.WithKeys("up", "k", "down", "j"), key.WithHelp("↑/↓",   "navigate")),
		SortNum: key.NewBinding(key.WithKeys("1", "2", "3", "4"),     key.WithHelp("1-4",    "sort column")),
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

// AppModel is the root Bubble Tea model. It owns the splash screen, tab
// navigation, help overlay, and delegates messages to the active child model.
type AppModel struct {
	active      activeTab
	uninstall   UninstallModel
	monitor     MonitorModel
	splash      SplashModel
	styles      Styles
	keys        goneKeyMap
	helpView    help.Model // full help overlay (ShowAll=true)
	footerHelp  help.Model // footer key hints (ShowAll=false)
	width       int
	height      int
	ready       bool
	showSplash  bool
	showHelp    bool
}

// NewApp constructs the root AppModel. If initialSearch is non-empty the
// Uninstall tab starts scanning for that term immediately.
func NewApp(initialSearch string) AppModel {
	// Full-screen help overlay
	hv := help.New()
	hv.ShowAll = true
	hv.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4")).Bold(true)
	hv.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hv.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	// Footer short-help bar
	fh := help.New()
	fh.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4")).Bold(true)
	fh.Styles.ShortDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	fh.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	return AppModel{
		active:     tabUninstall,
		uninstall:  NewUninstall(initialSearch),
		monitor:    NewMonitorModel(),
		splash:     NewSplashModel(),
		styles:     DefaultStyles(),
		keys:       defaultKeyMap(),
		helpView:   hv,
		footerHelp: fh,
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
		contentHeight := msg.Height - 5 // header(2) + gap(1) + status(1) + footer(1)
		m.uninstall = m.uninstall.SetSize(msg.Width, contentHeight)
		m.monitor = m.monitor.SetSize(msg.Width, contentHeight)
		m.helpView.SetWidth(m.width - 16)
		m.footerHelp.SetWidth(m.width - 6)
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
		if _, ok := msg.(refreshMsg); !ok {
			var cmd tea.Cmd
			m.uninstall, cmd = m.uninstall.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tabMonitor:
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

	// ── Header bar ──────────────────────────────────────────────────────────
	activeTabSt := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BCD4")).Padding(0, 1)
	inactiveTabSt := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)

	uninstallLabel := "Uninstall"
	monitorLabel := "Monitor"
	if c := m.uninstall.ItemCount(); c > 0 {
		uninstallLabel = fmt.Sprintf("Uninstall  %s", m.styles.TabBadge.Render(fmt.Sprintf("%d", c)))
	}
	if c := m.monitor.ProcCount(); c > 0 {
		monitorLabel = fmt.Sprintf("Monitor  %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fmt.Sprintf("(%d)", c)))
	}

	var uninstallTab, monitorTab string
	if m.active == tabUninstall {
		uninstallTab = activeTabSt.Render("◉ " + uninstallLabel)
		monitorTab = inactiveTabSt.Render("○ Monitor")
	} else {
		uninstallTab = inactiveTabSt.Render("○ Uninstall")
		monitorTab = activeTabSt.Render("◉ " + monitorLabel)
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Center, uninstallTab, "  ", monitorTab)

	goneLogo := gradientText("G O N E")
	tagline := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true).Render("hunt. select. trash.")
	brand := goneLogo + "  " + tagline

	tabsW := lipgloss.Width(tabs)
	brandW := lipgloss.Width(brand)
	contentW := m.width - 4
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

	// ── Tab content ─────────────────────────────────────────────────────────
	switch m.active {
	case tabUninstall:
		b.WriteString(m.uninstall.View())
	case tabMonitor:
		b.WriteString(m.monitor.View())
	}

	// ── Footer key hints ─────────────────────────────────────────────────────
	footerStr := m.footerHelp.View(m.keys)
	b.WriteString("\n")
	b.WriteString(m.styles.FooterBar.Width(m.width).Render(footerStr))

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}
