package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"gone/internal/sysinfo"
)

type refreshMsg time.Time

func doRefresh() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

// sortCol tracks which column the process list is sorted by.
type sortCol int

const (
	sortCPU sortCol = iota
	sortMem
	sortRSS
	sortPID
)

// MonitorModel is the system monitor tab model.
type MonitorModel struct {
	snapshot    sysinfo.Snapshot
	styles      Styles
	width       int
	height      int
	ready       bool
	cursor      int // selected row in process table
	sortBy      sortCol
	filterInput textinput.Model
	filtering   bool
}

func NewMonitorModel() MonitorModel {
	fi := textinput.New()
	fi.Placeholder = "filter by name…"
	fi.CharLimit = 40

	return MonitorModel{
		styles:      DefaultStyles(),
		sortBy:      sortCPU,
		filterInput: fi,
	}
}

func (m MonitorModel) Init() tea.Cmd {
	return doRefresh()
}

func (m MonitorModel) Update(msg tea.Msg) (MonitorModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case refreshMsg:
		m.snapshot = sysinfo.TakeSnapshot(15)
		m.ready = true
		procs := m.sortedProcs()
		if m.cursor >= len(procs) && len(procs) > 0 {
			m.cursor = len(procs) - 1
		}
		return m, doRefresh()

	case tea.KeyPressMsg:
		key := msg.String()

		if m.filtering {
			switch key {
			case "esc":
				m.filtering = false
				m.filterInput.Blur()
				m.filterInput.Reset()
				m.cursor = 0
				return m, nil
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
				// clamp cursor after filter text changes
				procs := m.sortedProcs()
				if m.cursor >= len(procs) && len(procs) > 0 {
					m.cursor = len(procs) - 1
				}
				return m, tea.Batch(cmds...)
			}
		}

		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sortedProcs())-1 {
				m.cursor++
			}
		case "1":
			m.sortBy = sortCPU
		case "2":
			m.sortBy = sortMem
		case "3":
			m.sortBy = sortRSS
		case "4":
			m.sortBy = sortPID
		case "/":
			m.filtering = true
			cmd := m.filterInput.Focus()
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m MonitorModel) SetSize(w, h int) MonitorModel {
	m.width = w
	m.height = h
	m.filterInput.SetWidth(w - 8)
	return m
}

func (m MonitorModel) View() string {
	if !m.ready {
		return "  Loading system info…"
	}

	s := m.snapshot
	var b strings.Builder

	// System gauges
	gaugeWidth := m.width/4 - 4
	if gaugeWidth < 12 {
		gaugeWidth = 12
	}
	gauges := lipgloss.JoinHorizontal(lipgloss.Top,
		m.gauge("CPU", fmt.Sprintf("%.1f%%", s.CPUPercent), gaugeWidth),
		m.gauge("RAM", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.MemUsed), sysinfo.HumanBytes(s.MemTotal)), gaugeWidth),
		m.gauge("Swap", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.SwapUsed), sysinfo.HumanBytes(s.SwapTotal)), gaugeWidth),
		m.gauge("Disk", fmt.Sprintf("%s free / %s", sysinfo.HumanBytes(s.DiskFree), sysinfo.HumanBytes(s.DiskTotal)), gaugeWidth),
	)
	b.WriteString(gauges)
	b.WriteString("\n\n")

	// Filter bar (shown when filtering is active)
	if m.filtering {
		filterBar := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("245")).
			Padding(0, 1).
			Width(m.width - 8).
			Render(m.filterInput.View())
		b.WriteString("  " + filterBar + "\n\n")
	} else {
		sortHint := m.styles.DimText.Render("Sort: [1]CPU [2]Mem [3]RSS [4]PID  ↑/↓ navigate  / filter")
		b.WriteString("  " + sortHint + "\n\n")
	}

	// Process table header
	header := fmt.Sprintf("  %-8s %-25s %8s %8s %12s", "PID", "Name", "CPU%", "MEM%", "RSS")
	b.WriteString(m.styles.DimText.Render(header) + "\n")
	b.WriteString(m.styles.DimText.Render("  "+strings.Repeat("─", m.width-6)) + "\n")

	// Sort procs
	procs := m.sortedProcs()

	if len(procs) == 0 && m.filtering {
		b.WriteString(m.styles.DimText.Render("  No processes match filter.") + "\n")
		return b.String()
	}

	// Process rows
	for i, p := range procs {
		line := fmt.Sprintf("  %-8d %-25s %8.1f %8.1f %12s",
			p.PID,
			truncateName(p.Name, 25),
			p.CPU,
			p.Mem,
			sysinfo.HumanBytes(p.RSS),
		)
		if i == m.cursor {
			b.WriteString(m.styles.Cursor.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

func (m MonitorModel) sortedProcs() []sysinfo.ProcInfo {
	procs := make([]sysinfo.ProcInfo, len(m.snapshot.Procs))
	copy(procs, m.snapshot.Procs)

	if m.filtering && m.filterInput.Value() != "" {
		lower := strings.ToLower(m.filterInput.Value())
		filtered := procs[:0]
		for _, p := range procs {
			if strings.Contains(strings.ToLower(p.Name), lower) {
				filtered = append(filtered, p)
			}
		}
		procs = filtered
	}

	switch m.sortBy {
	case sortCPU:
		sort.Slice(procs, func(i, j int) bool { return procs[i].CPU > procs[j].CPU })
	case sortMem:
		sort.Slice(procs, func(i, j int) bool { return procs[i].Mem > procs[j].Mem })
	case sortRSS:
		sort.Slice(procs, func(i, j int) bool { return procs[i].RSS > procs[j].RSS })
	case sortPID:
		sort.Slice(procs, func(i, j int) bool { return procs[i].PID < procs[j].PID })
	}
	return procs
}

func (m MonitorModel) gauge(label, value string, width int) string {
	title := m.styles.TabActive.Render(label)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(value)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Width(width).
		Render(title + "\n" + val)
}

func truncateName(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
