package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"gone/internal/remover"
	"gone/internal/scanner"
)

// --- Item type for the list ---

type fileItem struct {
	path     string
	size     int64
	modTime  string
	kind     string
	selected bool
}

func (f fileItem) FilterValue() string { return f.path }
func (f fileItem) Title() string       { return f.path }
func (f fileItem) Description() string {
	return fmt.Sprintf("%s  %s  %s", f.kind, HumanSize(f.size), f.modTime)
}

// --- Custom delegate ---

type fileDelegate struct {
	styles  Styles
	maxSize int64
}

func (d fileDelegate) Height() int                             { return 1 }
func (d fileDelegate) Spacing() int                            { return 0 }
func (d fileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d fileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	f, ok := item.(fileItem)
	if !ok {
		return
	}
	check := "[ ]"
	if f.selected {
		check = "[x]"
	}
	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}

	sizeStr := HumanSize(f.size)
	// Color by size
	switch {
	case f.size < 1024*1024:
		sizeStr = d.styles.SizeSmall.Render(sizeStr)
	case f.size < 100*1024*1024:
		sizeStr = d.styles.SizeMedium.Render(sizeStr)
	default:
		sizeStr = d.styles.SizeLarge.Render(sizeStr)
	}

	// Mini progress bar (static, no animation)
	var barStr string
	if d.maxSize > 0 {
		pct := float64(f.size) / float64(d.maxSize)
		p := progress.New(
			progress.WithColors(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")),
			progress.WithoutPercentage(),
			progress.WithWidth(10),
		)
		barStr = p.ViewAs(pct)
	} else {
		barStr = strings.Repeat("░", 10)
	}

	path := truncate(f.path, 45)
	kind := d.styles.DimText.Render(f.kind)
	line := fmt.Sprintf("%s%s %-45s %s %s %8s  %s", cursor, check, path, kind, barStr, sizeStr, f.modTime)

	if index == m.Index() {
		fmt.Fprint(w, d.styles.Cursor.Render(line))
	} else if f.selected {
		fmt.Fprint(w, d.styles.Selected.Render(line))
	} else {
		fmt.Fprint(w, line)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "…" + s[len(s)-max+1:]
}

// --- Scan result message ---

type scanResultMsg struct {
	items []fileItem
}

func runFullScan(term string) tea.Cmd {
	return func() tea.Msg {
		var items []fileItem

		// File scan
		matches, err := scanner.Search(term, scanner.GetScanPaths())
		if err != nil {
			return scanResultMsg{items: items}
		}
		for _, m := range matches {
			size := m.Size
			if m.IsDir {
				size = scanner.DirSize(m.Path)
			}
			items = append(items, fileItem{
				path:    m.Path,
				size:    size,
				modTime: m.ModTime.Format("2006-01-02"),
				kind:    m.Kind,
			})
		}

		// RC scan
		rcMatches := scanner.SearchRC(term)
		for _, rc := range rcMatches {
			items = append(items, fileItem{
				path:    fmt.Sprintf("%s:%d", rc.File, rc.LineNum),
				size:    int64(len(rc.Line)),
				modTime: "",
				kind:    "rc-line",
			})
		}

		return scanResultMsg{items: items}
	}
}

// --- Trash result message ---

type trashDoneMsg struct {
	count  int
	freed  int64
	errors []string
}

func trashSelected(items []fileItem, term string) tea.Cmd {
	return func() tea.Msg {
		var count int
		var freed int64
		var errs []string
		for _, item := range items {
			if item.kind == "rc-line" {
				continue // skip rc lines for now
			}
			if err := remover.MoveToTrash(item.path); err != nil {
				errs = append(errs, err.Error())
				continue
			}
			remover.AppendLog(remover.LogEntry{
				Path: item.path, Size: item.size, Kind: item.kind, SearchTerm: term,
			})
			count++
			freed += item.size
		}
		return trashDoneMsg{count: count, freed: freed, errors: errs}
	}
}

// --- Focus state ---

type focus int

const (
	focusSearch focus = iota
	focusList
)

// --- Uninstall model ---

type UninstallModel struct {
	textinput   textinput.Model
	spinner     spinner.Model
	list        list.Model
	viewport    viewport.Model
	styles      Styles
	focus       focus
	scanning    bool
	showPreview bool
	width       int
	height      int
	term        string
	status      string
}

func NewUninstallModel() UninstallModel {
	ti := textinput.New()
	ti.Placeholder = "Type a name to search (e.g. claude, nvm, rustup)..."
	ti.Focus()
	ti.CharLimit = 128

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	delegate := fileDelegate{styles: DefaultStyles()}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	vp := viewport.New()

	return UninstallModel{
		textinput:   ti,
		spinner:     sp,
		list:        l,
		viewport:    vp,
		styles:      DefaultStyles(),
		focus:       focusSearch,
		showPreview: true,
	}
}

func (m UninstallModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m UninstallModel) Update(msg tea.Msg) (UninstallModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		key := msg.String()

		switch m.focus {
		case focusSearch:
			switch key {
			case "enter":
				term := strings.TrimSpace(m.textinput.Value())
				if term == "" {
					return m, nil
				}
				m.term = term
				m.scanning = true
				m.status = fmt.Sprintf("Scanning for %q…", term)
				return m, tea.Batch(m.spinner.Tick, runFullScan(term))
			case "esc":
				return m, tea.Quit
			}

		case focusList:
			switch key {
			case " ", "space":
				idx := m.list.Index()
				items := m.list.Items()
				if idx >= 0 && idx < len(items) {
					if f, ok := items[idx].(fileItem); ok {
						f.selected = !f.selected
						items[idx] = f
						m.list.SetItem(idx, f)
					}
				}
				return m, nil
			case "esc":
				m.focus = focusSearch
				cmd := m.textinput.Focus()
				return m, cmd
			case "enter":
				sel := m.SelectedItems()
				if len(sel) == 0 {
					return m, nil
				}
				m.status = fmt.Sprintf("Trashing %d items…", len(sel))
				m.scanning = true
				return m, trashSelected(sel, m.term)
			}
		}

	case scanResultMsg:
		m.scanning = false
		items := make([]list.Item, len(msg.items))
		for i, it := range msg.items {
			items[i] = it
		}
		m.list.SetItems(items)
		// Compute maxSize for relative progress bars
		var maxSz int64
		for _, it := range msg.items {
			if it.size > maxSz {
				maxSz = it.size
			}
		}
		m.list.SetDelegate(fileDelegate{styles: m.styles, maxSize: maxSz})
		m.focus = focusList
		m.textinput.Blur()
		m.status = fmt.Sprintf("Found %d matches for %q", len(msg.items), m.term)
		// Set initial preview content for the first item
		if len(msg.items) > 0 {
			m.viewport.SetContent(previewContent(msg.items[0]))
		}
		return m, nil

	case trashDoneMsg:
		m.scanning = false
		m.status = fmt.Sprintf("Trashed %d items, freed %s", msg.count, HumanSize(msg.freed))
		if len(msg.errors) > 0 {
			m.status += fmt.Sprintf(" (%d errors)", len(msg.errors))
		}
		// Remove trashed items from list
		var remaining []list.Item
		for _, item := range m.list.Items() {
			if f, ok := item.(fileItem); ok && !f.selected {
				remaining = append(remaining, f)
			}
		}
		m.list.SetItems(remaining)
		return m, nil

	case spinner.TickMsg:
		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Route to focused component
	switch m.focus {
	case focusSearch:
		var cmd tea.Cmd
		m.textinput, cmd = m.textinput.Update(msg)
		cmds = append(cmds, cmd)
	case focusList:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
		// Update preview after list navigation
		if item, ok := m.list.SelectedItem().(fileItem); ok {
			m.viewport.SetContent(previewContent(item))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m UninstallModel) SetSize(w, h int) UninstallModel {
	m.width = w
	m.height = h
	m.textinput.SetWidth(w - 6)
	total := w - 2
	listW := total * 3 / 5
	previewW := total - listW - 4
	m.list.SetSize(listW, h-6)
	m.viewport.SetWidth(previewW)
	m.viewport.SetHeight(h - 6)
	m.showPreview = w > 80
	return m
}

// previewContent builds the text content shown in the preview pane for a given file item.
func previewContent(item fileItem) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Path: %s\n", item.path))
	b.WriteString(fmt.Sprintf("Type: %s\n", item.kind))
	b.WriteString(fmt.Sprintf("Size: %s\n", HumanSize(item.size)))
	if item.modTime != "" {
		b.WriteString(fmt.Sprintf("Modified: %s\n", item.modTime))
	}
	if item.kind == "dir" {
		entries, err := os.ReadDir(item.path)
		if err != nil {
			b.WriteString(fmt.Sprintf("Error reading dir: %v\n", err))
		} else {
			b.WriteString(fmt.Sprintf("Contains: %d entries\n", len(entries)))
			for i, e := range entries {
				if i >= 20 {
					b.WriteString(fmt.Sprintf("  ... and %d more\n", len(entries)-20))
					break
				}
				b.WriteString(fmt.Sprintf("  %s\n", e.Name()))
			}
		}
	}
	if item.kind == "rc-line" {
		parts := strings.SplitN(item.path, ":", 2)
		if len(parts) == 2 {
			b.WriteString(fmt.Sprintf("\nFile: %s\nLine: %s\n", parts[0], parts[1]))
		}
	}
	return b.String()
}

func (m UninstallModel) View() string {
	var b strings.Builder

	// Search bar (full width)
	searchStyle := m.styles.SearchBar
	if m.focus == focusSearch {
		searchStyle = m.styles.SearchBarActive
	}
	b.WriteString(searchStyle.Width(m.width - 6).Render(m.textinput.View()))
	b.WriteString("\n")

	if m.scanning {
		b.WriteString("\n  " + m.spinner.View() + " " + m.status + "\n")
	} else if len(m.list.Items()) > 0 {
		if m.showPreview {
			total := m.width - 2
			listW := total * 3 / 5
			previewContentW := total - listW - 4
			listView := lipgloss.NewStyle().Width(listW).Render(m.list.View())
			previewView := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
				Padding(0, 1).
				Width(previewContentW).
				Height(m.height - 8).
				Render(m.viewport.View())
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listView, previewView))
		} else {
			b.WriteString(m.list.View())
		}
	} else if m.term != "" {
		b.WriteString("\n  No matches found.\n")
	}

	// Status bar
	selected := 0
	var totalSize int64
	for _, item := range m.list.Items() {
		if f, ok := item.(fileItem); ok && f.selected {
			selected++
			totalSize += f.size
		}
	}
	status := m.status
	if selected > 0 {
		status = fmt.Sprintf("%d selected (%s) — Enter to trash", selected, HumanSize(totalSize))
	}
	bar := m.styles.StatusBar.Width(m.width - 4).Render(status)
	b.WriteString("\n" + bar)

	return b.String()
}

func (m UninstallModel) SelectedItems() []fileItem {
	var sel []fileItem
	for _, item := range m.list.Items() {
		if f, ok := item.(fileItem); ok && f.selected {
			sel = append(sel, f)
		}
	}
	return sel
}
