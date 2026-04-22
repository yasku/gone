package tui

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
	cpuBar      progress.Model
	ramBar      progress.Model
	swapBar     progress.Model
	diskBar     progress.Model
	killPending bool
	killTarget  sysinfo.ProcInfo
	killErr     string
	filterInput textinput.Model
	filtering   bool
}

func newGaugeBar() progress.Model {
	return progress.New(
		progress.WithColors(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")),
		progress.WithoutPercentage(),
		progress.WithWidth(20),
	)
}

func NewMonitorModel() MonitorModel {
	fi := textinput.New()
	fi.Placeholder = "filter by name…"
	fi.CharLimit = 40

	return MonitorModel{
		styles:      DefaultStyles(),
		sortBy:      sortCPU,
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
		m.snapshot = sysinfo.TakeSnapshot(15)
		m.ready = true
		procs := m.sortedProcs()
		if m.cursor >= len(procs) && len(procs) > 0 {
			m.cursor = len(procs) - 1
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
	gauges := lipgloss.JoinHorizontal(lipgloss.Top,
		m.gaugeView("CPU", fmt.Sprintf("%.1f%%", s.CPUPercent), m.cpuBar),
		m.gaugeView("RAM", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.MemUsed), sysinfo.HumanBytes(s.MemTotal)), m.ramBar),
		m.gaugeView("Swap", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.SwapUsed), sysinfo.HumanBytes(s.SwapTotal)), m.swapBar),
		m.gaugeView("Disk", fmt.Sprintf("%s free / %s", sysinfo.HumanBytes(s.DiskFree), sysinfo.HumanBytes(s.DiskTotal)), m.diskBar),
	)
	b.WriteString(gauges)
	b.WriteString("\n\n")

	// Filter bar or sort/action hint
	if m.filtering {
		filterBar := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("245")).
			Padding(0, 1).
			Width(m.width - 8).
			Render(m.filterInput.View())
		b.WriteString("  " + filterBar + "\n\n")
	} else {
		hint := "Sort: [1]CPU [2]Mem [3]RSS [4]PID  ↑/↓ navigate  x kill  / filter"
		if m.killErr != "" {
			hint = lipgloss.NewStyle().Foreground(lipgloss.Color("167")).Render(m.killErr)
		}
		b.WriteString("  " + m.styles.DimText.Render(hint) + "\n\n")
	}

	// Sort procs
	procs := m.sortedProcs()

	if len(procs) == 0 && m.filtering {
		b.WriteString(m.styles.DimText.Render("  No processes match filter.") + "\n")
		return b.String()
	}

	// Process table
	b.WriteString(m.buildTable(procs))

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

func (m MonitorModel) buildTable(procs []sysinfo.ProcInfo) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#9B59B6")).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)

	oddRow := cellStyle.Foreground(lipgloss.Color("252"))
	evenRow := cellStyle.Foreground(lipgloss.Color("245"))

	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		Headers("PID", "Name", "CPU%", "MEM%", "RSS").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if row == m.cursor {
				return m.styles.CursorRow
			}
			if col == 2 && row >= 0 && row < len(procs) { // CPU% column
				pct := procs[row].CPU
				switch {
				case pct >= 70.0:
					return cellStyle.Foreground(lipgloss.Color("#FF6B6B"))
				case pct >= 30.0:
					return cellStyle.Foreground(lipgloss.Color("#FFDD57"))
				default:
					return cellStyle.Foreground(lipgloss.Color("#69FF94"))
				}
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

func truncateName(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
