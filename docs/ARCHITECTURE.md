# gone — Architecture

## Project Overview

**gone** is a macOS-exclusive TUI application that combines an intelligent uninstaller with a real-time system monitor. Built with Bubble Tea v2, it provides five integrated tabs for hunting leftover files, monitoring CPU/RAM, tracking network I/O, reviewing operation history, and running security audits.

```
gone v2.0.0
├── OS: macOS only
├── Language: Go 1.26+
├── TUI Framework: Bubble Tea v2 (charm.land/bubbletea/v2)
└── Styling: Lip Gloss v2 (charm.land/lipgloss/v2)
```

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           gone TUI Application                          │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────────────────────────────────────┐    │
│  │ SplashModel │→ │              AppModel (Root)                    │    │
│  └─────────────┘  │  ┌──────────────────────────────────────────┐   │    │
│                   │  │ Header: Tab bar + Brand                   │   │    │
│                   │  ├──────────────────────────────────────────┤   │    │
│                   │  │                                          │   │    │
│                   │  │        Active Tab Content                │   │    │
│                   │  │   (Uninstall|Monitor|Network|Logs|Audit)│   │    │
│                   │  │                                          │   │    │
│                   │  ├──────────────────────────────────────────┤   │    │
│                   │  │ Footer: Key hints                        │   │    │
│                   │  └──────────────────────────────────────────┘   │    │
│                   │  Help overlay (? key)                          │    │
│                   └─────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────────────────┤
│                          Message Routing                                 │
│  tea.KeyPressMsg → tab switching, quit, help toggle                     │
│  tea.WindowSizeMsg → SetSize() on all tabs                              │
│  refreshMsg (2s) → Monitor tab                                          │
│  networkRefreshMsg (2s) → Network tab                                    │
│  auditRefreshMsg (30s) → Audit tab                                      │
│  scanItemMsg/scanDoneMsg → Uninstall tab (streaming)                   │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                         Package Structure                                │
├─────────────────────────────────────────────────────────────────────────┤
│  cmd/gone/main.go                                                       │
│  └── Entry point: tea.NewProgram(tui.NewApp(initialSearch))             │
│                                                                          │
│  internal/tui/          Bubble Tea Models                               │
│  ├── app.go             AppModel (root), tab routing, keybindings       │
│  ├── splash.go          Animated startup splash                         │
│  ├── uninstall.go       Search → scan → select → trash workflow        │
│  ├── monitor.go         CPU/RAM gauges + process table                 │
│  ├── network.go         Network interface RX/TX gauges                  │
│  ├── logs.go            Operations log viewer                           │
│  ├── audit.go           osquery-based security audit                   │
│  └── styles.go          Lip Gloss styles + gradientText()               │
│                                                                          │
│  internal/cli/          External tool integration                       │
│  ├── runner.go          ExecJSON, ExecStream, ExecSimple                │
│  ├── tool.go            Tool discovery via `which`                      │
│  ├── fd.go              fd wrapper (optional, falls back to fastwalk)  │
│  └── osquery.go         osqueryi wrapper (optional, graceful degr.)   │
│                                                                          │
│  internal/scanner/      Filesystem + RC scanning                        │
│  ├── scanner.go         SearchStream() parallel scanner                 │
│  ├── locations.go       Scan paths, skip lists                          │
│  └── rcscanner.go       Shell RC line scanner                           │
│                                                                          │
│  internal/remover/      Trash operations                                │
│  ├── trash.go           MoveToTrash() via osascript                     │
│  └── log.go             JSONL operation log                             │
│                                                                          │
│  internal/sysinfo/      gopsutil wrapper                                │
│  └── sysinfo.go         Snapshot, ProcInfo, NetInterface               │
└─────────────────────────────────────────────────────────────────────────┘
```

## Package Responsibilities

### `cmd/gone/main.go`

Entry point. Creates the Bubble Tea program with optional initial search term from CLI args.

```go
func main() {
    initialSearch := strings.Join(os.Args[1:], " ")
    p := tea.NewProgram(tui.NewApp(initialSearch))
    // ...
}
```

### `internal/tui/`

All Bubble Tea models. **All models use value receivers.**

| File | Model | Responsibility |
|------|-------|----------------|
| `app.go` | `AppModel` | Root model: owns all 5 tabs, tab navigation, global keybindings, WindowSize propagation |
| `splash.go` | `SplashModel` | Animated startup splash with spinner and gradient banner |
| `uninstall.go` | `UninstallModel` | Search bar → streaming scan → list with selection → confirmation → trash |
| `monitor.go` | `MonitorModel` | CPU/RAM/Swap/Disk gauges + sortable process table |
| `network.go` | `NetworkModel` | Per-interface RX/TX gauges with filtering |
| `logs.go` | `LogsModel` | Viewport-based log viewer with filtering |
| `audit.go` | `AuditModel` | osquery categories with status indicators |
| `styles.go` | `Styles` struct | All Lip Gloss styles + `gradientText()` helper |

### `internal/cli/`

Subprocess wrappers for external tools. **All tools called via `exec.Command`, NOT library bindings.**

| File | Function | Purpose |
|------|----------|---------|
| `runner.go` | `ExecJSON()` | Run command, decode JSON stdout into struct |
| `runner.go` | `ExecStream()` | Run command, process stdout line-by-line |
| `runner.go` | `ExecSimple()` | Run command, return combined output |
| `tool.go` | `Which()` | Find tool in PATH with caching |
| `tool.go` | `IsAvailable()` | Check if tool exists |
| `tool.go` | `AvailableTools()` | List all available optional tools |
| `fd.go` | `FastFind()` | Use `fd` if available, else fallback |
| `osquery.go` | `RunQuery()` | Execute osquery SQL query |

Graceful degradation: if `fd` is unavailable, scanner uses Go's `fastwalk`. If `osquery` is unavailable, audit tab shows install instructions.

### `internal/scanner/`

Parallel filesystem scanner using `fastwalk`.

| Function | Purpose |
|----------|---------|
| `Search(term, paths)` | Walk paths, return all matching entries |
| `SearchStream(term, paths)` | Channel-based streaming results |
| `DirSize(path)` | Calculate total size of directory |
| `SearchRC(term)` | Scan shell RC files for matching lines |

**Scan locations** (`scanner/locations.go`):
- `~/Library/Caches`
- `~/Library/Application Support`
- `~/Library/Preferences`
- `~/Library/Logs`
- `~/.config`
- `~/.local`
- `/usr/local`
- `/opt`

**Skip directories**: `node_modules`, `.git`, `.svn`, `vendor`, `.cargo`, `target`, `__pycache__`

### `internal/remover/`

Safe file removal via macOS Trash.

| File | Function | Purpose |
|------|----------|---------|
| `trash.go` | `MoveToTrash(absPath)` | AppleScript `osascript` to Finder → delete |
| `log.go` | `AppendLog(entry)` | Append JSONL entry to `~/.config/gone/operations.log` |

Files are **never deleted with `rm`** — always sent to Trash via Finder AppleScript.

### `internal/sysinfo/`

gopsutil v4 wrapper for system metrics.

| Function | Purpose |
|----------|---------|
| `TakeSnapshot(topN)` | CPU, RAM, Swap, Disk + top N processes by CPU |
| `GetNetInterfaces()` | Per-interface I/O statistics |
| `GetHostInfo()` | Uptime and hostname |
| `HumanBytes(b)` | Human-readable byte formatting |

## Message Routing

AppModel routes messages to child models:

```
tea.WindowSizeMsg → uninstall.SetSize() + monitor.SetSize() + network.SetSize()
                   → logs.SetSize() + audit.SetSize()

refreshMsg (2s tick) → monitor.Update() only
networkRefreshMsg (2s tick) → network.Update() only
auditRefreshMsg (30s tick) → audit.Update() only

scanItemMsg → uninstall.Update() (streaming during scan)
scanDoneMsg → uninstall.Update()

tea.KeyPressMsg → tab switching, quit, help toggle (handled in AppModel)
All other messages → active tab Update()
```

## Data Flow by Tab

### Uninstall Tab
```
User types term → Enter → startScan(term)
                                    │
                    scanner.SearchStream() → fastwalk parallel walk
                                    │
                    chan scanner.Match ←──────────────┐
                         │                             │
                   scanItemMsg ───────────────────────┘
                         │ (streaming)
                   list.InsertItem()
                         │
User selects (Space) → toggle selected flag
                         │
User presses Enter → confirmPending = true
                         │
User confirms → trashSelected(items)
                         │
              remover.MoveToTrash() → osascript Finder delete
                         │
              remover.AppendLog() → JSONL to ~/.config/gone/operations.log
```

### Monitor Tab
```
Init() → tea.Tick(2s, refreshMsg)
              │
         refreshMsg ──────────────────────────┐
              │                               │
         sysinfo.TakeSnapshot(50)             │
              │                               │
         ProcInfo{} slice ────────────────────┘
              │
         Update gauges (SetPercent)
         Update process table (sorted by CPU)
```

### Network Tab
```
Init() → tea.Tick(2s, networkRefreshMsg)
              │
         networkRefreshMsg ───────────────────┐
              │                              │
         sysinfo.GetNetInterfaces()          │
              │                              │
         NetInterface{} slice ───────────────┘
              │
         Update RX/TX gauges per interface
```

### Logs Tab
```
Init() → read ~/.config/gone/operations.log
              │
         Parse JSONL → LogEntry{}
              │
         viewport.SetContent() with color-coded entries
              │
User presses / → filter input → re-render with filtered entries
```

### Audit Tab
```
Init() → check osquery availability
              │
osquery available → RunQuery() for each category
              │
         Category results → display with status indicators
              │
osquery unavailable → show install instructions
```

## External Tool Integration Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Tool Discovery (tool.go)                      │
│                                                                  │
│  Which() uses exec.LookPath() with sync.Once caching            │
│  Available tools: fd, osqueryi, glances, bmon, mtr, nmap        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Graceful Degradation                         │
│                                                                  │
│  fd:     if available → use fd; else → fastwalk (Go)           │
│  osquery: if available → run queries; else → show instructions  │
└─────────────────────────────────────────────────────────────────┘
```

## Color Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Orange | `#FF6B35` | Accent |
| Purple | `#9B59B6` | Gradient start, borders |
| Cyan | `#00BCD4` | Gradient end, active elements, links |
| Dark Gray | `#1A1A1A` | Background (alt screen) |
| Gray 236 | `#2D2D2D` | Status bar background |
| Gray 240 | `#303030` | Inactive text |
| Gray 241 | `#3D3D3D` | Secondary text |
| Gray 245 | `#4D4D4D` | Dim text |
| Gray 252 | `#545454` | Bright text |

## View Returns `tea.View`

All Bubble Tea v2 models return `tea.View`, NOT string:

```go
func (m Model) View() tea.View {
    v := tea.NewView(contentString)
    v.AltScreen = true
    return v
}
```

## Key Architectural Patterns

1. **Value Receivers**: All model methods use value receivers
   ```go
   func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)
   ```

2. **SetSize Called by Parent**: Parent calls `SetSize()` on WindowSizeMsg
   ```go
   case tea.WindowSizeMsg:
       m.uninstall = m.uninstall.SetSize(msg.Width, contentHeight)
   ```

3. **Progress Bar Animation**: Must propagate `progress.FrameMsg`
   ```go
   case progress.FrameMsg:
       var cmd tea.Cmd
       m.cpuBar, cmd = m.cpuBar.Update(msg)
       cmds = append(cmds, cmd)
   ```

4. **lipgloss.Place for Block Centering**: When joining multiple elements, wrap with Place
   ```go
   lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Left, gauges)
   ```

5. **tea.NewView(content).AltScreen = true**: For proper full-screen rendering
