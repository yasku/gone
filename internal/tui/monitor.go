package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/progress"
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
	snapshot sysinfo.Snapshot
	styles   Styles
	width    int
	height   int
	ready    bool
	cursor   int // selected row in process table
	sortBy   sortCol
	cpuBar   progress.Model
	ramBar   progress.Model
	swapBar  progress.Model
	diskBar  progress.Model
}

func newGaugeBar() progress.Model {
	return progress.New(
		progress.WithColors(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")),
		progress.WithoutPercentage(),
		progress.WithWidth(20),
	)
}

func NewMonitorModel() MonitorModel {
	return MonitorModel{
		styles:  DefaultStyles(),
		sortBy:  sortCPU,
		cpuBar:  newGaugeBar(),
		ramBar:  newGaugeBar(),
		swapBar: newGaugeBar(),
		diskBar: newGaugeBar(),
	}
}

func (m MonitorModel) Init() tea.Cmd {
	return doRefresh()
}

func (m MonitorModel) Update(msg tea.Msg) (MonitorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case refreshMsg:
		m.snapshot = sysinfo.TakeSnapshot(15)
		m.ready = true
		// clamp cursor
		if m.cursor >= len(m.snapshot.Procs) && len(m.snapshot.Procs) > 0 {
			m.cursor = len(m.snapshot.Procs) - 1
		}
		s := m.snapshot
		// CPU: 0–100%
		cpuCmd := m.cpuBar.SetPercent(s.CPUPercent / 100.0)
		// RAM: used/total
		var ramPct float64
		if s.MemTotal > 0 {
			ramPct = float64(s.MemUsed) / float64(s.MemTotal)
		}
		ramCmd := m.ramBar.SetPercent(ramPct)
		// Swap: used/total
		var swapPct float64
		if s.SwapTotal > 0 {
			swapPct = float64(s.SwapUsed) / float64(s.SwapTotal)
		}
		swapCmd := m.swapBar.SetPercent(swapPct)
		// Disk: used/total
		var diskPct float64
		if s.DiskTotal > 0 {
			diskPct = float64(s.DiskUsed) / float64(s.DiskTotal)
		}
		diskCmd := m.diskBar.SetPercent(diskPct)
		return m, tea.Batch(doRefresh(), cpuCmd, ramCmd, swapCmd, diskCmd)

	case progress.FrameMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd
		m.cpuBar, cmd = m.cpuBar.Update(msg)
		cmds = append(cmds, cmd)
		m.ramBar, cmd = m.ramBar.Update(msg)
		cmds = append(cmds, cmd)
		m.swapBar, cmd = m.swapBar.Update(msg)
		cmds = append(cmds, cmd)
		m.diskBar, cmd = m.diskBar.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

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
		}
	}
	return m, nil
}

func (m MonitorModel) SetSize(w, h int) MonitorModel {
	m.width = w
	m.height = h
	barW := w/4 - 8
	if barW < 12 {
		barW = 12
	}
	m.cpuBar.SetWidth(barW)
	m.ramBar.SetWidth(barW)
	m.swapBar.SetWidth(barW)
	m.diskBar.SetWidth(barW)
	return m
}

func (m MonitorModel) View() string {
	if !m.ready {
		return "  Loading system info…"
	}

	s := m.snapshot
	var b strings.Builder

	// System gauges
	gauges := lipgloss.JoinHorizontal(lipgloss.Top,
		m.gaugeView("CPU", fmt.Sprintf("%.1f%%", s.CPUPercent), m.cpuBar),
		m.gaugeView("RAM", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.MemUsed), sysinfo.HumanBytes(s.MemTotal)), m.ramBar),
		m.gaugeView("Swap", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.SwapUsed), sysinfo.HumanBytes(s.SwapTotal)), m.swapBar),
		m.gaugeView("Disk", fmt.Sprintf("%s free / %s", sysinfo.HumanBytes(s.DiskFree), sysinfo.HumanBytes(s.DiskTotal)), m.diskBar),
	)
	b.WriteString(gauges)
	b.WriteString("\n\n")

	// Sort hint
	sortHint := m.styles.DimText.Render("Sort: [1]CPU [2]Mem [3]RSS [4]PID  ↑/↓ navigate")
	b.WriteString("  " + sortHint + "\n\n")

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

func (m MonitorModel) gaugeView(label, value string, bar progress.Model) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BCD4")).Render(label)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(value)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
		Padding(0, 2).
		Width(m.width/4 - 4).
		Render(title + "\n" + bar.View() + "\n" + val)
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

func truncateName(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
