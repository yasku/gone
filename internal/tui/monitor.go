package tui

import (
	"fmt"
	"sort"
	"strings"
	"syscall"
	"time"

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

type killDoneMsg struct {
	pid int32
	err error
}

func killProc(pid int32) tea.Cmd {
	return func() tea.Msg {
		err := syscall.Kill(int(pid), syscall.SIGTERM)
		return killDoneMsg{pid: pid, err: err}
	}
}

// MonitorModel is the system monitor tab model.
type MonitorModel struct {
	snapshot    sysinfo.Snapshot
	styles      Styles
	width       int
	height      int
	ready       bool
	cursor      int // selected row in process table
	sortBy      sortCol
	killPending bool
	killTarget  sysinfo.ProcInfo
	killErr     string
}

func NewMonitorModel() MonitorModel {
	return MonitorModel{
		styles: DefaultStyles(),
		sortBy: sortCPU,
	}
}

func (m MonitorModel) Init() tea.Cmd {
	return doRefresh()
}

func (m MonitorModel) Update(msg tea.Msg) (MonitorModel, tea.Cmd) {
	if m.killPending {
		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "enter":
				m.killPending = false
				return m, killProc(m.killTarget.PID)
			case "esc":
				m.killPending = false
				return m, nil
			}
		}
		return m, nil // swallow all input while confirm is active
	}

	switch msg := msg.(type) {
	case refreshMsg:
		m.snapshot = sysinfo.TakeSnapshot(15)
		m.ready = true
		// clamp cursor
		if m.cursor >= len(m.snapshot.Procs) && len(m.snapshot.Procs) > 0 {
			m.cursor = len(m.snapshot.Procs) - 1
		}
		return m, doRefresh()

	case killDoneMsg:
		if msg.err != nil {
			m.killErr = fmt.Sprintf("kill %d: %v", msg.pid, msg.err)
		} else {
			m.killErr = ""
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.snapshot.Procs)-1 {
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
		case "x":
			procs := m.sortedProcs()
			if len(procs) > 0 && m.cursor < len(procs) {
				m.killTarget = procs[m.cursor]
				m.killPending = true
			}
		}
	}
	return m, nil
}

func (m MonitorModel) SetSize(w, h int) MonitorModel {
	m.width = w
	m.height = h
	return m
}

func (m MonitorModel) View() string {
	if !m.ready {
		return "  Loading system info…"
	}

	if m.killPending {
		return m.killConfirmView()
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

	// Sort + action hint
	hint := "Sort: [1]CPU [2]Mem [3]RSS [4]PID  ↑/↓ navigate  x kill"
	if m.killErr != "" {
		hint = lipgloss.NewStyle().Foreground(lipgloss.Color("167")).Render(m.killErr)
	}
	b.WriteString("  " + m.styles.DimText.Render(hint) + "\n\n")

	// Process table header
	header := fmt.Sprintf("  %-8s %-25s %8s %8s %12s", "PID", "Name", "CPU%", "MEM%", "RSS")
	b.WriteString(m.styles.DimText.Render(header) + "\n")
	b.WriteString(m.styles.DimText.Render("  "+strings.Repeat("─", m.width-6)) + "\n")

	// Sort procs
	procs := m.sortedProcs()

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

func (m MonitorModel) killConfirmView() string {
	msg := fmt.Sprintf(
		"  Send SIGTERM to %q (PID %d)?\n\n  [Enter] Confirm    [Esc] Cancel",
		m.killTarget.Name, m.killTarget.PID,
	)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("245")).
		Padding(1, 3).
		Width(50).
		Render(msg)
	return lipgloss.Place(m.width, max(0, m.height-4), lipgloss.Center, lipgloss.Center, box)
}

func (m MonitorModel) sortedProcs() []sysinfo.ProcInfo {
	procs := make([]sysinfo.ProcInfo, len(m.snapshot.Procs))
	copy(procs, m.snapshot.Procs)
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
