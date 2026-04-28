package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"gone/internal/sysinfo"
)

type LogEntry struct {
	Timestamp  string `json:"ts"`
	Operation  string `json:"op"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	Kind       string `json:"kind"`
	SearchTerm string `json:"term,omitempty"`
}

type LogsModel struct {
	entries     []LogEntry
	filter      string
	filterInput textinput.Model
	viewport    viewport.Model
	filtering   bool
	styles      Styles
	width       int
	height      int
	ready       bool
	logPath     string
}

func NewLogsModel() LogsModel {
	fi := textinput.New()
	fi.Placeholder = "filter logs..."
	fi.CharLimit = 40

	vp := viewport.New()

	m := LogsModel{
		filterInput: fi,
		viewport:    vp,
		styles:      DefaultStyles(),
		logPath:     getLogPath(),
	}
	m.loadLogs()
	m.ready = true
	return m
}

func (m LogsModel) Init() tea.Cmd {
	return nil
}

func getLogPath() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	if home != "" {
		return filepath.Join(home, ".config", "gone", "operations.log")
	}
	return ""
}

func (m LogsModel) loadLogs() {
	if m.logPath == "" {
		return
	}

	file, err := os.Open(m.logPath)
	if err != nil {
		return
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	m.entries = entries
	if len(entries) > 0 {
		m.viewport.SetContent(m.formatLogs())
	}
}

func (m LogsModel) formatLogs() string {
	var b strings.Builder

	filtered := m.entries
	if m.filter != "" {
		filter := strings.ToLower(m.filter)
		var filteredEntries []LogEntry
		for _, entry := range m.entries {
			if strings.Contains(strings.ToLower(entry.Path), filter) ||
				strings.Contains(strings.ToLower(entry.Operation), filter) ||
				strings.Contains(strings.ToLower(entry.Kind), filter) {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		filtered = filteredEntries
	}

	for _, entry := range filtered {
		b.WriteString(m.formatEntry(entry))
	}

	return b.String()
}

func (m LogsModel) formatEntry(entry LogEntry) string {
	var opColor string
	switch entry.Operation {
	case "TRASH":
		opColor = "#FF6B6B"
	case "SCAN":
		opColor = "#69FF94"
	case "START":
		opColor = "#00BCD4"
	default:
		opColor = "#FFDD57"
	}

	opStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(opColor)).Bold(true)
	pathStyle := m.styles.DimText
	sizeStyle := m.styles.DimText

	ts := entry.Timestamp
	if ts == "" {
		ts = "unknown"
	}

	size := sysinfo.HumanBytes(uint64(entry.Size))
	if entry.Size == 0 {
		size = "-"
	}

	return fmt.Sprintf("%s %s %s %s %s\n",
		m.styles.DimText.Render(ts),
		opStyle.Render(fmt.Sprintf("%-5s", entry.Operation)),
		pathStyle.Render(truncateLogPath(entry.Path, 50)),
		sizeStyle.Render(size),
		m.styles.DimText.Render(entry.Kind),
	)
}

func truncateLogPath(path string, max int) string {
	if len(path) <= max {
		return path
	}
	return "…" + path[len(path)-max+1:]
}

func (m LogsModel) Update(msg tea.Msg) (LogsModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()

		if m.filtering {
			switch key {
			case "esc":
				m.filtering = false
				m.filterInput.Blur()
				m.filterInput.Reset()
				m.filter = ""
				m.viewport.SetContent(m.formatLogs())
				return m, nil
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
				m.filter = m.filterInput.Value()
				m.viewport.SetContent(m.formatLogs())
				return m, tea.Batch(cmds...)
			}
		}

		switch key {
		case "/":
			m.filtering = true
			cmd := m.filterInput.Focus()
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "r":
			m.loadLogs()
		case "c":
			m.filter = ""
			m.filterInput.Reset()
			m.viewport.SetContent(m.formatLogs())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m LogsModel) SetSize(w, h int) LogsModel {
	m.width = w
	m.height = h
	m.filterInput.SetWidth(w - 8)
	m.viewport.SetWidth(w - 4)
	m.viewport.SetHeight(h - 8)
	return m
}

func (m LogsModel) View() tea.View {
	if !m.ready {
		return tea.NewView("  Loading logs...")
	}

	var b strings.Builder

	if m.filtering {
		filterBar := m.styles.SearchBarActive.Width(m.width - 8).Render(m.filterInput.View())
		b.WriteString("  " + filterBar + "\n\n")
	} else {
		hint := m.styles.DimText.Render("  / filter  r refresh  c clear")
		b.WriteString(hint + "\n\n")
	}

	count := len(m.entries)
	if m.filter != "" {
		filtered := 0
		filter := strings.ToLower(m.filter)
		for _, entry := range m.entries {
			if strings.Contains(strings.ToLower(entry.Path), filter) ||
				strings.Contains(strings.ToLower(entry.Operation), filter) {
				filtered++
			}
		}
		b.WriteString(m.styles.DimText.Render(fmt.Sprintf("  %d / %d entries", filtered, count)) + "\n\n")
	} else {
		b.WriteString(m.styles.DimText.Render(fmt.Sprintf("  %d entries", count)) + "\n\n")
	}

	b.WriteString(m.viewport.View())

	return tea.NewView(b.String())
}
