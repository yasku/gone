package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"gone/internal/cli"
)

type auditRefreshMsg time.Time

func doAuditRefresh() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg { return auditRefreshMsg(t) })
}

type AuditCategory struct {
	Name    string
	Count   int
	Status  string
	Details []map[string]string
}

type AuditModel struct {
	available  bool
	categories []AuditCategory
	loading    bool
	spinner    spinner.Model
	selected   int
	styles     Styles
	width      int
	height     int
	ready      bool
}

func NewAuditModel() AuditModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := AuditModel{
		spinner: sp,
		styles:  DefaultStyles(),
	}
	m.available = cli.IsAvailable("osqueryi")
	m.loadAudit()
	m.ready = true
	return m
}

func (m AuditModel) Init() tea.Cmd {
	return doAuditRefresh()
}

func (m AuditModel) loadAudit() {
	if !m.available {
		return
	}

	m.loading = true

	o := cli.NewOsqueryRunner()
	if !o.IsAvailable() {
		m.available = false
		m.loading = false
		return
	}

	var categories []AuditCategory

	apps, _ := o.GetApps()
	categories = append(categories, AuditCategory{
		Name:    "Applications",
		Count:   len(apps),
		Status:  "ok",
		Details: apps,
	})

	startup, _ := o.GetStartupItems()
	status := "ok"
	if len(startup) > 10 {
		status = "warn"
	}
	categories = append(categories, AuditCategory{
		Name:    "Startup Items",
		Count:   len(startup),
		Status:  status,
		Details: startup,
	})

	plugins, _ := o.GetBrowserPlugins()
	categories = append(categories, AuditCategory{
		Name:    "Browser Plugins",
		Count:   len(plugins),
		Status:  "ok",
		Details: plugins,
	})

	connections, _ := o.GetNetworkConnections()
	status = "ok"
	if len(connections) > 50 {
		status = "warn"
	}
	categories = append(categories, AuditCategory{
		Name:    "Network Connections",
		Count:   len(connections),
		Status:  status,
		Details: connections,
	})

	ports, _ := o.GetOpenPorts()
	status = "ok"
	if len(ports) > 20 {
		status = "warn"
	}
	categories = append(categories, AuditCategory{
		Name:    "Open Ports",
		Count:   len(ports),
		Status:  status,
		Details: ports,
	})

	m.categories = categories
	m.loading = false
}

func (m AuditModel) Update(msg tea.Msg) (AuditModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case auditRefreshMsg:
		m.loadAudit()
		return m, doAuditRefresh()

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyPressMsg:
		key := msg.String()

		switch key {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.categories)-1 {
				m.selected++
			}
		case "r":
			m.loadAudit()
		case "1", "2", "3", "4", "5":
			idx := int(key[0] - '1')
			if idx >= 0 && idx < len(m.categories) {
				m.selected = idx
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m AuditModel) SetSize(w, h int) AuditModel {
	m.width = w
	m.height = h
	return m
}

func (m AuditModel) View() tea.View {
	if !m.ready {
		return tea.NewView("  Loading audit info...")
	}

	var b strings.Builder

	b.WriteString(m.styles.DimText.Render("  Security Audit via osquery  ·  r refresh  ↑/↓ navigate\n\n"))

	if !m.available {
		b.WriteString(m.notAvailableView())
		return tea.NewView(b.String())
	}

	if m.loading {
		b.WriteString("  " + m.spinner.View() + " Scanning...\n")
		return tea.NewView(b.String())
	}

	b.WriteString(m.categoriesView())

	return tea.NewView(b.String())
}

func (m AuditModel) notAvailableView() string {
	notAvail := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 3).
		Width(50).
		Render(
			"⚠ osquery not installed\n\n" +
				"Install osquery to enable security auditing:\n\n" +
				"  brew install osquery\n\n" +
				"then run: gone\n\n" +
				"or use without audit:\n  Tab to other features",
		)
	return lipgloss.Place(m.width, m.height/2, lipgloss.Center, lipgloss.Center, notAvail)
}

func (m AuditModel) categoriesView() string {
	var b strings.Builder

	for i, cat := range m.categories {
		statusIcon := "✓"
		statusColor := "#69FF94"
		if cat.Status == "warn" {
			statusIcon = "⚠"
			statusColor = "#FFDD57"
		} else if cat.Status == "error" {
			statusIcon = "✗"
			statusColor = "#FF6B6B"
		}

		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true)
		cursor := "  "
		if i == m.selected {
			cursor = "◉ "
		}

		line := fmt.Sprintf("%s%s %s %s (%d)",
			cursor,
			statusStyle.Render(statusIcon),
			lipgloss.NewStyle().Bold(i == m.selected).Render(cat.Name),
			m.styles.DimText.Render(fmt.Sprintf("[%d]", cat.Count)),
			cat.Count,
		)

		if i == m.selected {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("#1A1A2E")).
				Foreground(lipgloss.Color("#00BCD4")).
				Render(line)
		}

		b.WriteString(line + "\n")

		if i == m.selected && len(cat.Details) > 0 {
			b.WriteString(m.styles.DimText.Render("    ") + m.formatDetails(cat.Details) + "\n")
		}
	}

	return b.String()
}

func (m AuditModel) formatDetails(details []map[string]string) string {
	if len(details) == 0 {
		return ""
	}

	var lines []string
	for i, d := range details {
		if i >= 3 {
			lines = append(lines, m.styles.DimText.Render(fmt.Sprintf("  ... and %d more", len(details)-3)))
			break
		}

		var parts []string
		for k, v := range d {
			if v != "" && k != "" {
				parts = append(parts, fmt.Sprintf("%s=%s", k, v))
			}
		}
		if len(parts) > 0 {
			lines = append(lines, strings.Join(parts, " "))
		}
	}

	return strings.Join(lines, " | ")
}
