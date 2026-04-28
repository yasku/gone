# gone — AGENTS.md

## IMPORTANT: These are HARD RULES

1. **Before reading files**: Check if there's an existing file that answers your question.
2. **Before implementing**: Run `./scripts/test-all.sh` to verify baseline works.
3. **After any change**: Run `go test ./internal/...` + `go vet ./...` + `go build -o gone ./cmd/gone`.
4. **Never commit** without running full test suite first.
5. **Always use value receivers** for Bubble Tea models.
6. **Always propagate progress.FrameMsg** to all progress bars.
7. **Always return tea.View** from View() method, NOT string.
8. **Always use charm.land/*/v2** import paths, NOT github.com/charmbracelet/*.

---

## Workflow (ALWAYS FOLLOW THIS ORDER)

### 1. Before Starting Any Task
```bash
# Verify project compiles and tests pass
./scripts/test-all.sh
```

### 2. Investigate Before Implementing
Use available CLIs to research (see CLIs section below).

### 3. Make Changes
- Implement feature/fix
- Keep value receivers for models
- Keep progress bars animation working

### 4. After Changes
```bash
go test ./internal/...      # Test all packages
go vet ./internal/...        # Vet all packages
go build -o gone ./cmd/gone # Build binary
```
If adding new functionality: add tests for new package.

### 5. Before Committing
```bash
./scripts/test-all.sh       # Full suite: test + vet + fmt
```

---

## Available CLIs

### mmx-cli (MiniMax)
**Purpose:** Web search, image analysis, text generation for research.

```bash
# Install (already done)
npm install -g mmx-cli
mmx auth login --api-key YOUR_KEY

# Search
mmx search query "search text"

# Analyze image (pass screenshot, ask question)
mmx vision /path/to/image "What is this?"

# Text generation
mmx text chat --message "your question"
```

**When to use:** Before implementing a feature, research libraries/APIs. Debug errors. Understand patterns.

### gh cli (GitHub)
**Purpose:** Read repos, search code, get docs directly from source.

```bash
# Read README from a repo
gh api repos/OWNER/REPO/contents/README.md --jq '.content' | base64 -d | head -100

# Read specific file
gh api repos/OWNER/REPO/contents/path/to/file.md --jq '.content' | base64 -d

# Search code across repo
gh search code "SearchTerm" --repo OWNER/REPO

# Get releases
gh api repos/OWNER/REPO/releases --jq '.[0].tag_name'
```

**When to use:** Get official docs from GitHub directly. Verify library API before using. Find examples in upstream repos.

### ctx7 (context7)
**Purpose:** Library-specific code snippets and documentation.

```bash
npx ctx7 library LIBRARY_NAME
```

**When to use:** Find how to use a specific library (e.g., `npx ctx7 library minimax` for MiniMax SDK). Get code snippets directly.

---

## Versions (verify before referencing docs)
| Library | Version | Docs |
|---------|---------|------|
| `charm.land/bubbletea/v2` | v2.0.5 | https://pkg.go.dev/charm.land/bubbletea/v2 |
| `charm.land/bubbles/v2` | v2.1.0 | https://pkg.go.dev/charm.land/bubbles/v2 |
| `charm.land/lipgloss/v2` | v2.0.3 | https://pkg.go.dev/charm.land/lipgloss/v2 |
| `github.com/shirou/gopsutil/v4` | v4.26.3 | https://pkg.go.dev/github.com/shirou/gopsutil/v4 |

> **Critical:** We're on v2 of Bubble Tea, Bubbles, and Lip Gloss. Import path is `charm.land/*/v2`, NOT `github.com/charmbracelet/*`. Docs for v1 are outdated.

---

## Build & Test Commands
```bash
go build -o gone ./cmd/gone      # Build binary
go test ./...                     # Test all packages
go test -race ./...              # Race detector
./scripts/test-all.sh             # Full: test + vet + fmt
./scripts/check-tools.sh          # Verify optional tools
```

---

## Project Structure
```
gone/
├── cmd/gone/main.go             # Entry: tea.NewProgram(tui.NewApp(initialSearch))
├── internal/
│   ├── cli/                     # Subprocess wrappers: ExecJSON, ExecStream, ExecSimple
│   ├── scanner/                # SearchStream() parallel scanner
│   ├── remover/                # Trash operations + JSONL operation log
│   ├── sysinfo/                # gopsutil wrapper: Snapshot, ProcInfo, NetInterface
│   └── tui/                    # All TUI models (AppModel owns 5 tabs)
│       ├── app.go              # Root model, tab routing, keybindings
│       ├── uninstall.go        # Search → scan → select → trash
│       ├── monitor.go         # CPU/RAM gauges + process table
│       ├── network.go          # Interface RX/TX gauges
│       ├── logs.go            # Operations log viewer
│       ├── audit.go           # Security audit via osquery
│       ├── splash.go          # Startup splash screen
│       └── styles.go          # All Lipgloss styles + gradientText()
├── scripts/                    # Automation scripts
└── docs/                       # Documentation (ARCHITECTURE, SETUP, USER_GUIDE, DEVELOPER_GUIDE)
```

---

## How the App Works

### AppModel (app.go)
Root model owns all 5 tabs. Handles:
- Tab cycling (`tab` key)
- Global keybindings (`ctrl+c`, `?` help, `tab`)
- Message routing to child models
- Window sizing propagation

### Tab Ownership
Each tab is its own model with value receiver methods. AppModel routes messages:

```
refreshMsg        → Monitor (auto-refresh every 2s)
networkRefreshMsg → Network (auto-refresh every 2s)
auditRefreshMsg   → Audit (auto-refresh every 30s)
scanItemMsg/scanDoneMsg → Uninstall (streaming during scan)
tea.WindowSizeMsg → calls SetSize() on all tabs
```

### Message Loop
1. `Init()` called once, returns initial commands (timers, subscriptions)
2. `Update()` called on every message, returns updated model + optional command
3. `View()` called after Update, returns `tea.View` for rendering

### 5 Tabs Flow
1. **Uninstall** — type search term → fastwalk parallel scan → select files → trash via osascript → log to JSONL
2. **Monitor** — gopsutil snapshot every 2s → animated gauges → sortable process table → `x` to kill process
3. **Network** — net.IOCounters every 2s → per-interface RX/TX gauges → filter by interface name
4. **Logs** — read `~/.config/gone/operations.log` → color-coded entries → viewport scroll
5. **Audit** — osquery queries → category list → graceful degradation if osquery unavailable

---

## Bubble Tea v2 Key Concepts

### Model = State + Init + Update + View
```go
type Model struct {
    field type
}

func (m Model) Init() tea.Cmd           // Return initial command(s)
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)  // Handle messages
func (m Model) View() tea.View          // Return tea.View, NOT string
```

### Value Receivers
Models use value receivers. Mutations via `model, cmd := model.Update(msg)`.
Parent reassigns: `m.monitor = m.monitor.SetSize(w, h)`

### tea.Msg = Any Type
Messages can be any type. Use type switch:
```go
switch msg := msg.(type) {
case tea.KeyPressMsg:
    // handle key
case refreshMsg:
    // handle refresh
}
```

### tea.Cmd = Function That Returns Msg
```go
type Cmd func() Msg

// Common pattern for timers:
func doRefresh() tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

// Async work:
func killProc(pid int32) tea.Cmd {
    return func() tea.Msg {
        err := syscall.Kill(int(pid), syscall.SIGTERM)
        return killDoneMsg{pid: pid, err: err}
    }
}
```

### tea.Batch()
Combine multiple commands:
```go
return m, tea.Batch(cmd1, cmd2, cmd3)
```

### tea.Tick()
```go
tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
```

### tea.WindowSizeMsg
Sent when terminal resizes. Parent propagates to children via SetSize().

### Custom Messages
Define as type aliases or structs:
```go
type refreshMsg time.Time
type scanItemMsg struct { item fileItem }
type killDoneMsg struct { pid int32; err error }
```

---

## Lipgloss v2 Key Concepts

### NewStyle() Chain
```go
lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#00BCD4")).
    Padding(0, 1).
    Border(lipgloss.RoundedBorder())
```

### Color Types
```go
lipgloss.Color("#00BCD4")  // Hex (true color)
lipgloss.Color("240")      // ANSI 256
lipgloss.Color("5")        // ANSI 16
lipgloss.Blend1D(n, color1, color2)  // Gradient
lipgloss.Darken(c, 0.5)              // Darken
lipgloss.Lighten(c, 0.35)            // Lighten
```

### Layout Helpers
```go
lipgloss.JoinHorizontal(lipgloss.Center, elem1, elem2, elem3)
lipgloss.JoinVertical(lipgloss.Center, elem1, elem2)
lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
```

### View Returns tea.View
```go
func (m Model) View() tea.View {
    return tea.NewView("content string")
}
// OR for setting fields:
func (m Model) View() tea.View {
    v := tea.NewView("")
    v.AltScreen = true
    v.SetContent(content)
    return v
}
```

### Conditional Styling (for tables, lists)
```go
StyleFunc func(row, col int) lipgloss.Style
// Return different styles based on row/index for visual feedback
```

---

## Bubbles Components We Use
| Component | Purpose |
|-----------|---------|
| `textinput.Model` | Search bars, filter inputs |
| `spinner.Model` | Loading indicators |
| `progress.Model` | Gauge bars (CPU, RAM, RX/TX) |
| `viewport.Model` | Scrollable content (preview, logs) |
| `list.Model` | File list in Uninstall tab |
| `help.Model` | Help overlay and footer hints |
| `table.Model` | Process table in Monitor tab |
| `key.Binding` | Keybinding definitions |

### Progress Bar Animation
Must call `.Update(progress.FrameMsg)` each frame:
```go
case progress.FrameMsg:
    var cmd tea.Cmd
    m.cpuBar, cmd = m.cpuBar.Update(msg)
    cmds = append(cmds, cmd)
    return m, tea.Batch(cmds...)
```

### SetPercent() Uses Pointer Receiver
```go
m.cpuBar.SetPercent(0.75)  // Returns nothing, mutates in place
```

---

## External Tool Integration (internal/cli)
All external tools via subprocess, NOT library bindings:
- `fd` — fast file finder

Graceful degradation if unavailable:
```go
if cli.IsAvailable("fd") {
    // use fd
}
// else fallback to Go implementation
```

---

## Common Mistakes to Avoid

1. **Init() returning nil** — Init() must return a Cmd. Use `tea.Batch()` for multiple.
2. **Mutating model in Update without reassigning** — `model, cmd := model.Update(msg)` requires reassignment.
3. **SetSize() called by child instead of parent** — Parent calls SetSize() on WindowSizeMsg.
4. **Progress bars not animating** — Must propagate `progress.FrameMsg` to each progress bar.
5. **Using v1 docs/imports** — v2 uses `charm.land/*/v2`, not `github.com/charmbracelet/*`.
6. **View() returning string** — v2 returns `tea.View`, use `tea.NewView()`.
7. **tea.KeyMsg instead of tea.KeyPressMsg** — v2 uses `tea.KeyPressMsg`.
8. **Center aligned elements not truly centered** — When joining multiple elements (gauges, boxes), wrap with `lipgloss.Place(m.width, ..., lipgloss.Center, lipgloss.Left, content)` to center the entire block.

---

## Lipgloss Layout Patterns (Critical)

### Centering Multiple Elements as a Block
When you need to center multiple elements together (e.g., 4 gauge boxes centered), use `lipgloss.Place`:

```go
// ❌ WRONG - elements are left-aligned within their container
gauges := lipgloss.JoinHorizontal(lipgloss.Center, gauge1, gauge2, gauge3, gauge4)
b.WriteString(gauges)

// ✅ CORRECT - the entire block is centered on the screen
gauges := lipgloss.JoinHorizontal(lipgloss.Center, gauge1, gauge2, gauge3, gauge4)
b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Left, gauges))
```

**Syntax:** `lipgloss.Place(width, height, horizontalAlign, verticalAlign, content)`
- `width, height` — dimensions of the container area
- `horizontalAlign` — Center/Left/Right
- `verticalAlign` — Top/Center/Bottom

**Rule:** When joining multiple elements and they need to be centered as a group, ALWAYS wrap with `lipgloss.Place(m.width, ..., lipgloss.Center, ...)`.

### Available Alignments
- Horizontal: `lipgloss.Center`, `lipgloss.Left`, `lipgloss.Right`
- Vertical: `lipgloss.Top`, `lipgloss.Center`, `lipgloss.Bottom`

---

## mmx-cli Reference
See [mmx-cli.md](./mmx-cli.md) — use for research before implementing features.

---

## Custom Message Types Reference

| Message Type | Location | Purpose |
|-------------|----------|---------|
| `refreshMsg` | `internal/tui/monitor.go` | Trigger Monitor refresh (2s tick) |
| `networkRefreshMsg` | `internal/tui/network.go` | Trigger Network refresh (2s tick) |
| `auditRefreshMsg` | `internal/tui/audit.go` | Trigger Audit refresh (30s tick) |
| `scanItemMsg` | `internal/tui/uninstall.go` | Streaming scan result item |
| `scanDoneMsg` | `internal/tui/uninstall.go` | Streaming scan complete |
| `trashDoneMsg` | `internal/tui/uninstall.go` | Trash operation complete |
| `splashDoneMsg` | `internal/tui/app.go` | Splash screen animation done |

---

## Documentation Files

| File | Purpose |
|------|---------|
| `docs/ARCHITECTURE.md` | System architecture, package responsibilities, data flow |
| `docs/SETUP.md` | Installation, prerequisites, build from source |
| `docs/USER_GUIDE.md` | User documentation, keybindings, FAQ |
| `docs/DEVELOPER_GUIDE.md` | Contributor guide, adding tabs/tools, code style |