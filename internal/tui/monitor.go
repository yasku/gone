package tui

// MonitorModel implements the system monitor tab. It polls gopsutil every 2 s,
// renders CPU/RAM/Swap/Disk gauge bars, and displays a sortable, filterable
// process table with an inline kill-confirmation flow.

import (
	"fmt"
	"sort"
	"strings"
	"syscall"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/sahilm/fuzzy"
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

type MonitorModel struct {
	snapshot    sysinfo.Snapshot
	styles      Styles
	width       int
	height      int
	ready       bool
	cursor      int
	sortBy      sortCol
	cpuBar      progress.Model
	ramBar      progress.Model
	swapBar     progress.Model
	diskBar     progress.Model
	killPending bool
	killTarget  sysinfo.ProcInfo
	killErr     string
	filterInput textinput.Model
	filtering   bool
	paused      bool
}

func newGaugeBar() progress.Model {
	return progress.New(
		progress.WithColors(lipgloss.Color("#FF6B35"), lipgloss.Color("240")),
		progress.WithoutPercentage(),
		progress.WithWidth(24),
	)
}

// NewMonitorModel creates a MonitorModel with default gauge bars and an empty
// filter input ready for use.
func NewMonitorModel() MonitorModel {
	fi := textinput.New()
	fi.Placeholder = "filter by name…"
	fi.CharLimit = 40

	return MonitorModel{
		styles:      DefaultStyles(),
		sortBy:      sortMem,
		paused:      true,
		cpuBar:      newGaugeBar(),
		ramBar:      newGaugeBar(),
		swapBar:     newGaugeBar(),
		diskBar:     newGaugeBar(),
		filterInput: fi,
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

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case refreshMsg:
		if !m.paused {
			m.snapshot = sysinfo.TakeSnapshot(15)
			s := m.snapshot
			cpuCmd := m.cpuBar.SetPercent(s.CPUPercent / 100.0)
			var ramPct float64
			if s.MemTotal > 0 {
				ramPct = float64(s.MemUsed) / float64(s.MemTotal)
			}
			ramCmd := m.ramBar.SetPercent(ramPct)
			var swapPct float64
			if s.SwapTotal > 0 {
				swapPct = float64(s.SwapUsed) / float64(s.SwapTotal)
			}
			swapCmd := m.swapBar.SetPercent(swapPct)
			var diskPct float64
			if s.DiskTotal > 0 {
				diskPct = float64(s.DiskUsed) / float64(s.DiskTotal)
			}
			diskCmd := m.diskBar.SetPercent(diskPct)
			procs := m.sortedProcs()
			if m.cursor >= len(procs) && len(procs) > 0 {
				m.cursor = len(procs) - 1
			}
			return m, tea.Batch(doRefresh(), cpuCmd, ramCmd, swapCmd, diskCmd)
		}
		m.ready = true
		return m, doRefresh()

	case progress.FrameMsg:
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

	case killDoneMsg:
		if msg.err != nil {
			m.killErr = fmt.Sprintf("kill %d: %v", msg.pid, msg.err)
		} else {
			m.killErr = ""
		}
		return m, nil

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
		case "x":
			procs := m.sortedProcs()
			if len(procs) > 0 && m.cursor < len(procs) {
				m.killTarget = procs[m.cursor]
				m.killPending = true
			}
		case "/":
			m.filtering = true
			cmd := m.filterInput.Focus()
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case " ":
			m.paused = !m.paused
		case "r":
			m.snapshot = sysinfo.TakeSnapshot(15)
			m.ready = true
			s := m.snapshot
			cpuCmd := m.cpuBar.SetPercent(s.CPUPercent / 100.0)
			var ramPct float64
			if s.MemTotal > 0 {
				ramPct = float64(s.MemUsed) / float64(s.MemTotal)
			}
			ramCmd := m.ramBar.SetPercent(ramPct)
			var swapPct float64
			if s.SwapTotal > 0 {
				swapPct = float64(s.SwapUsed) / float64(s.SwapTotal)
			}
			swapCmd := m.swapBar.SetPercent(swapPct)
			var diskPct float64
			if s.DiskTotal > 0 {
				diskPct = float64(s.DiskUsed) / float64(s.DiskTotal)
			}
			diskCmd := m.diskBar.SetPercent(diskPct)
			return m, tea.Batch(cpuCmd, ramCmd, swapCmd, diskCmd)
		}
	}

	return m, tea.Batch(cmds...)
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
	m.filterInput.SetWidth(w - 8)
	return m
}

func (m MonitorModel) View() tea.View {
	if !m.ready {
		return tea.NewView("  Loading system info…")
	}

	if m.killPending {
		return tea.NewView(m.killConfirmView())
	}

	s := m.snapshot
	var b strings.Builder

// System gauges — 4 boxes centered as a block
	totalW := m.width - 4
	gaugeW := totalW / 4
	gauges := lipgloss.JoinHorizontal(lipgloss.Center,
		m.gaugeView("CPU", fmt.Sprintf("%.1f%%", s.CPUPercent), m.cpuBar, gaugeW),
		m.gaugeView("RAM", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.MemUsed), sysinfo.HumanBytes(s.MemTotal)), m.ramBar, gaugeW),
		m.gaugeView("Swap", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.SwapUsed), sysinfo.HumanBytes(s.SwapTotal)), m.swapBar, gaugeW),
		m.gaugeView("Disk", fmt.Sprintf("%s free / %s", sysinfo.HumanBytes(s.DiskFree), sysinfo.HumanBytes(s.DiskTotal)), m.diskBar, gaugeW),
	)
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Left, gauges))
	b.WriteString("\n\n")

	// Resolve procs early so count is available for the hint bar
	procs := m.sortedProcs()
	total := len(m.snapshot.Procs)

	// Filter bar or sort/action hint
	if m.filtering {
		filterBar := m.styles.SearchBarActive.Width(m.width - 8).Render(m.filterInput.View())
		b.WriteString("  " + filterBar + "\n\n")
	} else {
		pauseLabel := "▶"
		if m.paused {
			pauseLabel = "⏸"
		}
		countPart := fmt.Sprintf("%d procs", total)
		hint := fmt.Sprintf("Sort: [1]CPU [2]Mem [3]RSS [4]PID  ↑/↓  x kill  / filter  [space] %s  [r] refresh  ·  %s", pauseLabel, countPart)
		b.WriteString("  " + m.styles.DimText.Render(hint) + "\n\n")
	}

	if m.killErr != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("167")).Render("  " + m.killErr) + "\n")
	}

	if len(procs) == 0 && m.filtering {
		b.WriteString(m.styles.DimText.Render("  No processes match filter.") + "\n")
		return tea.NewView(b.String())
	}

	// Process table centered like the gauges
	b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Left, m.buildTable(procs)))

	return tea.NewView(b.String())
}

func (m MonitorModel) gaugeView(label, value string, bar progress.Model, width int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B35")).Render(label)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(value)
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width).
		Render(title + "\n" + bar.View() + "\n" + val)
}

// ProcCount returns the number of processes captured in the most recent
// snapshot, or 0 if no snapshot has been taken yet.
func (m MonitorModel) ProcCount() int {
	return len(m.snapshot.Procs)
}

func (m MonitorModel) buildTable(procs []sysinfo.ProcInfo) string {
	pidHdr, nameHdr, cpuHdr, memHdr, rssHdr := "PID", "Name", "CPU%", "MEM%", "RSS"
	switch m.sortBy {
	case sortCPU:
		cpuHdr = "CPU% ▼"
	case sortMem:
		memHdr = "MEM% ▼"
	case sortRSS:
		rssHdr = "RSS ▼"
	case sortPID:
		pidHdr = "PID ▼"
	}

	orange := lipgloss.Color("#FF6B35")
	gray := lipgloss.Color("240")
	lightGray := lipgloss.Color("252")

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(orange).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	oddRow := cellStyle.Foreground(gray)
	evenRow := cellStyle.Foreground(lightGray)

	borderStyle := lipgloss.NewStyle().Foreground(gray)

	t := table.New().
		Width(m.width - 4).
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		Headers(pidHdr, nameHdr, cpuHdr, memHdr, rssHdr).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if row == m.cursor {
				return m.styles.CursorRow
			}
			if row%2 == 0 {
				return evenRow
			}
			return oddRow
		})

	for _, p := range procs {
		t.Row(
			fmt.Sprintf("%d", p.PID),
			truncateName(p.Name, 25),
			fmt.Sprintf("%.1f", p.CPU),
			fmt.Sprintf("%.1f", p.Mem),
			sysinfo.HumanBytes(p.RSS),
		)
	}

	return t.Render()
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

	if m.filtering && m.filterInput.Value() != "" {
		names := make([]string, len(procs))
		for i, p := range procs {
			names[i] = p.Name
		}
		matches := fuzzy.Find(m.filterInput.Value(), names)
		filtered := make([]sysinfo.ProcInfo, 0, len(matches))
		for _, match := range matches {
			filtered = append(filtered, procs[match.Index])
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

func truncateName(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
