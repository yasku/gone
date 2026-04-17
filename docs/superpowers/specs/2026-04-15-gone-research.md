# gone — Verified Research & Code Patterns

Reference file. The spec links here. Do not read top-to-bottom — jump to the section you need.

## Import Paths (v2 vanity domains)

```go
import (
    tea       "charm.land/bubbletea/v2"
    "charm.land/bubbles/v2/list"
    "charm.land/bubbles/v2/spinner"
    "charm.land/bubbles/v2/viewport"
    "charm.land/bubbles/v2/textinput"
    "charm.land/lipgloss/v2"
    "github.com/evertras/bubble-table/table"
    "github.com/shirou/gopsutil/v4/cpu"
    "github.com/shirou/gopsutil/v4/mem"
    "github.com/shirou/gopsutil/v4/disk"
    "github.com/shirou/gopsutil/v4/process"
    "github.com/charlievieth/fastwalk"
)
```

## Pattern: Bubble Tea v2 Minimal App

```go
type model struct {
    width, height int
    ready         bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if msg.String() == "q" { return m, tea.Quit }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.ready = true
    }
    return m, nil
}

func (m model) View() tea.View {
    if !m.ready { return tea.NewView("loading...") }
    return tea.NewView(fmt.Sprintf("Terminal: %dx%d\nPress q to quit", m.width, m.height))
}

func main() { tea.NewProgram(model{}).Run() }
```

## Pattern: Tabs (sub-model routing)

```go
type tab int
const ( uninstallTab tab = iota; monitorTab )

type rootModel struct {
    active   tab
    uninst   uninstallModel
    monitor  monitorModel
    width, height int
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        if msg.String() == "tab" {
            m.active = (m.active + 1) % 2
            return m, nil
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    // route to active tab
    var cmd tea.Cmd
    switch m.active {
    case uninstallTab:
        m.uninst, cmd = m.uninst.Update(msg)
    case monitorTab:
        m.monitor, cmd = m.monitor.Update(msg)
    }
    return m, cmd
}
```

Key: always route tick messages to monitor regardless of active tab, or it freezes.

## Pattern: Split Pane (40/60)

```go
func (m model) View() tea.View {
    leftW := m.width*40/100 - leftStyle.GetHorizontalFrameSize()
    rightW := m.width - leftW - rightStyle.GetHorizontalFrameSize()
    left := leftStyle.Width(leftW).Render(m.list.View())
    right := rightStyle.Width(rightW).Render(m.viewport.View())
    return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, left, right))
}
```

## Pattern: List Multi-Select

```go
type fileItem struct {
    path     string
    size     int64
    modTime  time.Time
    selected bool
}

func (f fileItem) FilterValue() string { return f.path }
func (f fileItem) Title() string {
    check := "[ ]"
    if f.selected { check = "[x]" }
    return check + " " + f.path
}
func (f fileItem) Description() string {
    return fmt.Sprintf("%s  %s", humanSize(f.size), f.modTime.Format("2006-01-02"))
}

// In Update, on space key:
item := m.list.SelectedItem().(fileItem)
item.selected = !item.selected
cmd := m.list.SetItem(m.list.Index(), item)
```

## Pattern: Custom Delegate (single-line with checkbox)

```go
type fileDelegate struct{ styles delegateStyles }

func (d fileDelegate) Height() int  { return 1 }
func (d fileDelegate) Spacing() int { return 0 }
func (d fileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d fileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    f, ok := item.(fileItem)
    if !ok { return }
    check := "[ ]"; if f.selected { check = "[x]" }
    cursor := "  "; if index == m.Index() { cursor = "> " }
    line := fmt.Sprintf("%s%s %-50s %8s  %s",
        cursor, check, f.path, humanSize(f.size), f.modTime.Format("Jan 02"))
    fmt.Fprint(w, line)
}
```

## Pattern: Spinner + Async Scan

```go
type scanResultMsg []fileItem
type scanDoneMsg struct{}

func runScan(term string, paths []string) tea.Cmd {
    return func() tea.Msg {
        var results []fileItem
        for _, root := range paths {
            fastwalk.Walk(nil, root, func(path string, d fs.DirEntry, err error) error {
                if err != nil { return nil }
                if strings.Contains(strings.ToLower(d.Name()), strings.ToLower(term)) {
                    info, _ := d.Info()
                    results = append(results, fileItem{path: path, size: info.Size(), modTime: info.ModTime()})
                }
                return nil
            })
        }
        return scanResultMsg(results)
    }
}

func (m model) Init() tea.Cmd {
    return tea.Batch(m.spinner.Tick, runScan(m.searchTerm, scanPaths))
}
```

## Pattern: tea.Tick (2s dashboard refresh)

```go
type refreshMsg time.Time

func doRefresh() tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

func (m monitorModel) Init() tea.Cmd { return doRefresh() }

func (m monitorModel) Update(msg tea.Msg) (monitorModel, tea.Cmd) {
    switch msg.(type) {
    case refreshMsg:
        m.cpuPct, _ = cpu.Percent(0, false)
        m.memInfo, _ = mem.VirtualMemory()
        m.swapInfo, _ = mem.SwapMemory()
        m.diskInfo, _ = disk.Usage("/")
        m.procs = getTopProcesses(10)
        return m, doRefresh()
    }
    return m, nil
}
```

## Pattern: macOS Trash (with Put Back)

```go
func moveToTrash(absPath string) error {
    script := fmt.Sprintf(
        `tell application "Finder" to delete POSIX file "%s"`,
        strings.ReplaceAll(absPath, `"`, `\"`),
    )
    cmd := exec.Command("/usr/bin/osascript", "-e", script)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("trash %s: %w: %s", absPath, err, stderr.String())
    }
    return nil
}
```

Requires Finder running. Path must be absolute. ~200ms per file.

## Pattern: fastwalk Scanner

```go
var scanPaths = []string{
    os.Getenv("HOME"),
    filepath.Join(os.Getenv("HOME"), "Library"),
    filepath.Join(os.Getenv("HOME"), ".config"),
    filepath.Join(os.Getenv("HOME"), ".local"),
    "/usr/local",
    "/opt/homebrew",
    "/opt",
}

// Skip dirs that are slow/irrelevant
var skipDirs = map[string]bool{
    "node_modules": true, ".git": true, ".Trash": true,
    "Caches": true, "DerivedData": true,
}
```

0.5-2s for 100k-500k files on SSD. Caps at 4 workers on macOS APFS.

## Pattern: gopsutil Process List (Top N)

```go
func getTopProcesses(n int) []procInfo {
    procs, _ := process.Processes()
    var infos []procInfo
    for _, p := range procs {
        name, err := p.Name()
        if err != nil { continue }
        cpuPct, _ := p.Percent(0)
        memPct, _ := p.MemoryPercent()
        memInfo, _ := p.MemoryInfo()
        rss := uint64(0)
        if memInfo != nil { rss = memInfo.RSS }
        infos = append(infos, procInfo{p.Pid, name, cpuPct, memPct, rss})
    }
    sort.Slice(infos, func(i, j int) bool { return infos[i].cpuPct > infos[j].cpuPct })
    if len(infos) > n { infos = infos[:n] }
    return infos
}
```

First call with `Percent(0)` seeds baseline (returns 0). Second call returns delta. Need 1s gap between samples.

## Pattern: Shell RC Scanning

```go
rcFiles := []string{".zshrc", ".bashrc", ".bash_profile", ".profile", ".zprofile", ".zshenv"}
home, _ := os.UserHomeDir()

func scanRC(path string, term string) []rcMatch {
    f, err := os.Open(path)
    if err != nil { return nil }
    defer f.Close()
    var matches []rcMatch
    sc := bufio.NewScanner(f)
    lineNum := 0
    for sc.Scan() {
        lineNum++
        if strings.Contains(strings.ToLower(sc.Text()), strings.ToLower(term)) {
            matches = append(matches, rcMatch{path, lineNum, sc.Text()})
        }
    }
    return matches
}
```

## Pattern: Sortable Process Table (evertras/bubble-table)

```go
cols := []table.Column{
    table.NewColumn("pid", "PID", 8).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
    table.NewColumn("name", "Name", 25),
    table.NewColumn("cpu", "CPU%", 8).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
    table.NewColumn("mem", "MEM%", 8).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
    table.NewColumn("rss", "RSS", 10).WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
}
t := table.New(cols).WithRows(rows).SortByDesc("cpu").Focused(true)
```

## Pattern: Lipgloss Theming

```go
type Styles struct {
    App          lipgloss.Style
    TabActive    lipgloss.Style
    TabInactive  lipgloss.Style
    ListNormal   lipgloss.Style
    ListSelected lipgloss.Style
    Preview      lipgloss.Style
    StatusBar    lipgloss.Style
    Spinner      lipgloss.Style
}

func DefaultStyles() Styles {
    return Styles{
        App:          lipgloss.NewStyle().Padding(1, 2),
        TabActive:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
        TabInactive:  lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        ListSelected: lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true),
        Preview:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
        StatusBar:    lipgloss.NewStyle().Background(lipgloss.Color("236")).Padding(0, 1),
    }
}
```

## Pattern: Spinner (bubbles/v2 options API)

```go
import "charm.land/bubbles/v2/spinner"

s := spinner.New(
    spinner.WithSpinner(spinner.Globe),  // Globe, Dot, Line, Pulse, MiniDot, Hamburger
    spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6"))),
)
// Init:   return s.Tick
// Update: case spinner.TickMsg: m.spinner, cmd = m.spinner.Update(msg)
// View:   m.spinner.View()
```

## Pattern: lipgloss.Place (center content in terminal)

```go
centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
```

## Pattern: BorderForegroundBlend (gradient border)

```go
style := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForegroundBlend(
        lipgloss.Color("#9B59B6"),
        lipgloss.Color("#00BCD4"),
    ).
    Padding(1, 4)
```

## Pattern: AltScreen via tea.View

```go
func (m Model) View() tea.View {
    v := tea.NewView(content)
    v.AltScreen = true
    return v
}
```

## Gotchas Cheat Sheet

| Issue | Fix |
|---|---|
| v2 `View()` returns `tea.View` not `string` | Use `tea.NewView(s)` |
| Monitor freezes when not active tab | Route tick messages to ALL sub-models regardless of active tab |
| `lipgloss.Width()` doesn't account for borders | Subtract `style.GetHorizontalFrameSize()` |
| First `cpu.Percent(0)` returns 0 | Seed on init, discard first result |
| `mem.Free` misleading on macOS | Use `mem.Available` instead |
| Paginator dots lag at 8k+ items | Set `paginator.Type = paginator.Arabic` |
| Trash fails for paths with quotes | Escape `"` in AppleScript or use `quoted form of` |
| fastwalk too many workers on APFS | Library auto-caps at 4; don't override |
| `~` not expanded by Go | Replace with `os.UserHomeDir()` manually |
