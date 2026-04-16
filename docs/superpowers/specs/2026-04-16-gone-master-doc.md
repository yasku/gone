# gone -- Master Project Documentation

> **Purpose:** This document is the single source of truth for the `gone` project. It contains everything a new developer or AI agent needs to understand, build, run, extend, and continue work on this project. Written at the end of the initial build session on 2026-04-16.

---

## Table of Contents

1. [Session Summary](#1-session-summary)
2. [Project Overview](#2-project-overview)
3. [Architecture](#3-architecture)
4. [Components](#4-components)
5. [Design Decisions](#5-design-decisions)
6. [How to Run](#6-how-to-run)
7. [How to Extend](#7-how-to-extend)
8. [Orchestrator](#8-orchestrator)
9. [Orchestration Methodology](#9-orchestration-methodology)
10. [Future Work](#10-future-work)
11. [Continuation Prompt](#11-continuation-prompt)

---

## 1. Session Summary

### What Happened

In a single Claude Code session on 2026-04-15 through 2026-04-16, the entire `gone` project was conceived, designed, researched, planned, implemented, reviewed, and QA'd. Here is the chronological sequence:

**Phase 1: Brainstorm & Problem Definition**
- Identified the problem: macOS has no clean uninstaller for CLI tools and dev environments. Files scatter across `~/.config`, `~/Library`, `/usr/local/bin`, `.zshrc`, and more. After years of installs, the system accumulates ghost configs and orphaned caches.
- Defined scope: a Go TUI with two tabs -- Uninstall (search + scan + trash) and Monitor (live system dashboard).

**Phase 2: Research (3 parallel agents)**
- Used **context7 MCP** to pull live documentation for Bubble Tea v2, Bubbles v2, Lipgloss v2, fastwalk, gopsutil v4, and evertras/bubble-table.
- Used **WebSearch** and **WebFetch** to verify import paths, API breaking changes (v1 vs v2), and discover gotchas.
- Used **GitHub CLI (gh)** to check release tags, open issues, and compatibility.
- Three parallel research agents ran simultaneously: one for TUI frameworks, one for system libraries, one for macOS-specific APIs (osascript, Finder trash).
- All findings were saved to `docs/superpowers/specs/2026-04-15-gone-research.md` -- a verified knowledge base with real code patterns, not hallucinated APIs.

**Phase 3: Design & Planning**
- Wrote a detailed design spec: `docs/superpowers/specs/2026-04-15-gone-design.md`
- Wrote a Karpathy-method implementation plan with 9 incremental tasks (0-8), each producing a runnable binary: `docs/superpowers/plans/2026-04-15-gone.md`
- Each task had explicit verify-and-commit checkpoints.

**Phase 4: Orchestrator Development**
- Built `gone/orchestrator/supervisor.ts` -- a Bun/TypeScript script that dispatches Claude Code workers as subprocesses.
- Each worker gets a fresh session, reads the plan file, implements exactly ONE task, commits, and stops.
- Researched Mini-Agent (MiniMax) as an alternative orchestrator: `docs/superpowers/specs/2026-04-16-mini-agent-research.md`
- Wrote a comprehensive orchestration guide: `docs/superpowers/specs/2026-04-16-orchestration-guide.md`

**Phase 5: Implementation (9 autonomous workers)**
- Dispatched 9 Claude Code worker sessions via `supervisor.ts`, running tasks 0 through 8 sequentially.
- Each worker read the plan and knowledge base, implemented its task, ran tests, committed.
- Workers adapted to real-world conditions: Bubble Tea v2's API differences from the plan (e.g., `WithAltScreen()` removed, `viewport.New()` takes option funcs, `Width` is a method not a field), `evertras/bubble-table` incompatible with v2 (replaced with hand-rolled table).

**Phase 6: Code Review (Opus)**
- Dispatched a senior Go code review agent that ran `go build`, `go test -race`, `go vet`, and inspected every file against the spec.
- Found and fixed 1 critical bug: **data race in `scanner.Search()`** -- fastwalk's concurrent callbacks were appending to a shared slice without a mutex.
- Found and fixed 2 medium bugs: duplicate results from overlapping scan paths, package-init HOME resolution.
- Report: `docs/superpowers/specs/2026-04-16-gone-code-review.md`

**Phase 7: QA (Opus)**
- Dispatched a QA agent that verified all spec requirements, ran the full test suite with `-race`, and checked every function against its documented behavior.
- Found and fixed 8 issues total (1 critical, 4 medium, 3 low).
- Added 2 new tests: `TestSearchConcurrentSafety`, `TestSearchDeduplicatesOverlappingPaths`.
- Report: `docs/superpowers/specs/2026-04-16-gone-qa-report.md`

### Final State

- **8 tests pass** across 3 packages (scanner: 5, remover: 1, sysinfo: 1, plus 1 RC scanner test)
- `go build`, `go test -race`, `go vet` all clean
- All 9 tasks from the plan implemented and committed
- Full CHANGELOG maintained at `gone/CHANGELOG.md`

---

## 2. Project Overview

### What Is `gone`?

`gone` is a **macOS TUI application** for finding and removing leftover files from uninstalled CLI tools and dev environments. It also includes a live system monitoring dashboard.

### The Problem It Solves

When you install developer tools on macOS (via npm, curl, Homebrew, or direct download), files scatter across dozens of locations:
- `~/.claude/`, `~/.nvm/`, `~/.rustup/`
- `~/Library/Application Support/...`
- `/usr/local/bin/`, `/opt/homebrew/bin/`
- Lines injected into `~/.zshrc`, `~/.bashrc`, `~/.profile`
- `~/.config/...`, `~/.local/...`

There is no standard uninstaller. After years, the system accumulates ghost configs, orphaned caches, and wasted disk space.

### What It Does

**Tab 1 -- Uninstall:**
1. Type a search term (e.g., "claude", "nvm", "rustup")
2. `gone` scans your entire filesystem and shell RC files for matches
3. Results appear in a fuzzy-filterable list with checkboxes
4. A preview pane shows file details (path, size, type, directory contents)
5. Select items with Space, press Enter to send them to macOS Trash
6. Trash uses Finder's "Put Back" mechanism -- items are recoverable

**Tab 2 -- Monitor:**
1. Live dashboard showing CPU%, RAM, Swap, Disk usage
2. Top 15 processes by CPU/MEM/RSS in a sortable table
3. Refreshes every 2 seconds
4. Sort by CPU (1), MEM (2), RSS (3), PID (4)

### Who It's For

This is a **personal tool**. macOS only. No safety rails, no dry-run mode, no recipes. The user sees everything before pressing Enter. Trash (not hard-delete) is the safety net.

### Key Principles

- **User takes full responsibility.** The tool finds files; you decide what to delete.
- **Trash only, never hard-delete.** Finder AppleScript ensures Put Back works.
- **Speed matters.** fastwalk scans 100k+ files in <2s on SSD.
- **No recipes database.** Pure filesystem scan + name matching. No config files describing what "nvm" installed.
- **Single binary.** `go build` produces one executable. No runtime dependencies.

---

## 3. Architecture

### High-Level Architecture

```
gone (single Go binary)
├── cmd/gone/main.go          -- Entry point, creates tea.Program
└── internal/
    ├── tui/                   -- Bubble Tea TUI layer
    │   ├── app.go             -- Root model, tab routing, help overlay
    │   ├── uninstall.go       -- Search + list + preview + trash flow
    │   ├── monitor.go         -- System dashboard + process table
    │   └── styles.go          -- Lipgloss theme + HumanSize formatter
    ├── scanner/               -- Filesystem scanning
    │   ├── scanner.go         -- fastwalk-based file/dir finder
    │   ├── rcscanner.go       -- Shell RC line matcher
    │   └── locations.go       -- Scan paths + skip dirs
    ├── remover/               -- Trash + logging
    │   ├── trash.go           -- osascript Finder trash
    │   └── log.go             -- JSONL operation log
    └── sysinfo/               -- System metrics
        └── sysinfo.go         -- gopsutil wrapper -> Snapshot
```

### Bubble Tea v2 Model Architecture

`gone` uses the **Elm Architecture** (Model-Update-View) via Bubble Tea v2.

```
┌──────────────────────────────────────────────────────┐
│                    AppModel (app.go)                  │
│  Fields: active tab, width, height, showHelp         │
│                                                      │
│  ┌────────────────────┐  ┌────────────────────────┐  │
│  │  UninstallModel    │  │    MonitorModel         │  │
│  │  (uninstall.go)    │  │    (monitor.go)         │  │
│  │                    │  │                         │  │
│  │  textinput         │  │  snapshot (Snapshot)    │  │
│  │  spinner           │  │  cursor (int)           │  │
│  │  list (bubbles)    │  │  sortBy (sortCol)       │  │
│  │  viewport          │  │                         │  │
│  │  focus state       │  │  Receives: refreshMsg   │  │
│  └────────────────────┘  └────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

### Message Flow

```
tea.Program.Run()
  │
  ▼
AppModel.Init()
  ├── UninstallModel.Init() → textinput.Blink
  └── MonitorModel.Init()   → doRefresh() (first 2s tick)
  │
  ▼ (event loop)
AppModel.Update(msg)
  │
  ├── tea.KeyPressMsg
  │   ├── "ctrl+c" → tea.Quit
  │   ├── "?"      → toggle showHelp
  │   └── "tab"    → cycle active tab
  │
  ├── tea.WindowSizeMsg → resize both sub-models
  │
  ├── refreshMsg → ALWAYS routed to MonitorModel (prevents freeze)
  │
  └── (other msgs) → routed to ACTIVE tab only
        │
        ├── UninstallModel.Update(msg)
        │   ├── KeyPressMsg "enter" (search) → runFullScan() cmd
        │   ├── KeyPressMsg "space"          → toggle selection
        │   ├── KeyPressMsg "enter" (list)   → trashSelected() cmd
        │   ├── KeyPressMsg "esc"            → focus search / quit
        │   ├── scanResultMsg                → populate list
        │   ├── trashDoneMsg                 → update status, remove items
        │   └── spinner.TickMsg              → animate spinner
        │
        └── MonitorModel.Update(msg)
            ├── refreshMsg       → TakeSnapshot(15), schedule next tick
            └── KeyPressMsg      → cursor movement, sort column change
```

### Critical Pattern: Tick Routing

The `refreshMsg` from `tea.Tick` must ALWAYS be routed to `MonitorModel`, even when the Uninstall tab is active. Without this, the monitor's tick chain breaks and it freezes when you switch back to it. This is handled in `app.go` lines 75-79:

```go
// Always route refresh ticks to monitor (prevents freeze)
if _, ok := msg.(refreshMsg); ok {
    var cmd tea.Cmd
    m.monitor, cmd = m.monitor.Update(msg)
    cmds = append(cmds, cmd)
}
```

### Async I/O Pattern

All blocking I/O (filesystem scanning, trash operations, system metric collection) runs inside `tea.Cmd` functions, NOT in `Update()`. This keeps the TUI responsive:

```go
// This returns a tea.Cmd (function that returns tea.Msg)
func runFullScan(term string) tea.Cmd {
    return func() tea.Msg {
        // Blocking I/O happens here, in a goroutine managed by Bubble Tea
        matches, _ := scanner.Search(term, scanner.GetScanPaths())
        // ...
        return scanResultMsg{items: items}
    }
}
```

### View Rendering

Bubble Tea v2 changed `View()` to return `tea.View` instead of `string`. The `tea.View` struct has an `AltScreen` field:

```go
func (m AppModel) View() tea.View {
    v := tea.NewView(renderedString)
    v.AltScreen = true  // Enable alternate screen buffer
    return v
}
```

Sub-models (`UninstallModel`, `MonitorModel`) return `string` from their `View()` methods. Only `AppModel.View()` wraps the final output in `tea.View`.

---

## 4. Components

### `cmd/gone/main.go`

**Purpose:** Application entry point. Minimal -- creates the root TUI model and runs the Bubble Tea program.

**Key code:**
```go
func main() {
    p := tea.NewProgram(tui.NewApp())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

**Note:** No `tea.WithAltScreen()` -- v2 removed this option. Alt screen is set via `v.AltScreen = true` on the returned `tea.View`.

---

### `internal/tui/app.go` -- Root Model + Tab Routing

**Purpose:** Top-level Bubble Tea model. Manages tab switching, routes messages to sub-models, renders the tab bar and help overlay.

**Key types:**
- `activeTab` -- enum: `tabUninstall`, `tabMonitor`
- `AppModel` -- root model with `active activeTab`, `uninstall UninstallModel`, `monitor MonitorModel`, `showHelp bool`

**Key functions:**
- `NewApp() AppModel` -- constructor, initializes both sub-models
- `Init() tea.Cmd` -- batches both sub-model Init commands
- `Update(msg) (tea.Model, tea.Cmd)` -- handles ctrl+c, ?, tab, WindowSizeMsg; routes refreshMsg to monitor always; routes other messages to active tab
- `View() tea.View` -- renders tab bar (active/inactive lipgloss styles), content from active sub-model, and help overlay if `showHelp`

**Connections:** Imports and composes `UninstallModel` and `MonitorModel`. References `refreshMsg` type from `monitor.go`. Uses `Styles` from `styles.go`.

---

### `internal/tui/uninstall.go` -- Uninstall Tab

**Purpose:** The main search-scan-select-trash workflow. This is the largest file (~420 lines).

**Key types:**
- `fileItem` -- implements `list.Item`; fields: `path`, `size`, `modTime`, `kind`, `selected`
- `fileDelegate` -- custom single-line list delegate with checkbox rendering
- `scanResultMsg` -- message carrying `[]fileItem` after scan completes
- `trashDoneMsg` -- message with count, freed bytes, errors after trash completes
- `focus` -- enum: `focusSearch`, `focusList`
- `UninstallModel` -- main model with textinput, spinner, list, viewport, focus state

**Key functions:**
- `NewUninstallModel() UninstallModel` -- sets up textinput (placeholder, char limit), spinner (dot style), list (custom delegate, no title/help/status), viewport
- `Init() tea.Cmd` -- returns `textinput.Blink`
- `Update(msg) (UninstallModel, tea.Cmd)` -- state machine:
  - `focusSearch` + Enter → start scan (batch spinner tick + runFullScan)
  - `focusList` + Space → toggle selection via `SetItem`
  - `focusList` + Enter → trash selected items
  - `focusList` + Esc → return to search
  - `scanResultMsg` → populate list, set initial preview, switch to focusList
  - `trashDoneMsg` → update status, remove trashed items from list
- `SetSize(w, h int) UninstallModel` -- splits width 50/50 for list/preview, hides preview if width <= 80
- `previewContent(item fileItem) string` -- builds preview text (path, type, size, date, dir entries, RC line info)
- `View() string` -- renders search bar, spinner/list+preview/empty state, status bar
- `SelectedItems() []fileItem` -- returns all items with `selected == true`
- `runFullScan(term string) tea.Cmd` -- async: runs `scanner.Search` + `scanner.SearchRC`, returns `scanResultMsg`
- `trashSelected(items []fileItem, term string) tea.Cmd` -- async: calls `remover.MoveToTrash` + `remover.AppendLog` for each item

**Connections:** Imports `scanner` (Search, GetScanPaths, SearchRC, DirSize), `remover` (MoveToTrash, AppendLog, LogEntry), `bubbles/list`, `bubbles/spinner`, `bubbles/textinput`, `bubbles/viewport`, `lipgloss`.

---

### `internal/tui/monitor.go` -- Monitor Tab

**Purpose:** Live system dashboard with CPU/RAM/Swap/Disk gauges and a sortable process table.

**Key types:**
- `refreshMsg` -- `time.Time` alias, triggers 2-second refresh cycle
- `sortCol` -- enum: `sortCPU`, `sortMem`, `sortRSS`, `sortPID`
- `MonitorModel` -- model with `snapshot sysinfo.Snapshot`, `cursor int`, `sortBy sortCol`

**Key functions:**
- `doRefresh() tea.Cmd` -- returns `tea.Tick(2*time.Second, ...)` that sends `refreshMsg`
- `NewMonitorModel() MonitorModel` -- default sort: CPU
- `Init() tea.Cmd` -- starts the refresh tick chain
- `Update(msg) (MonitorModel, tea.Cmd)` -- handles refreshMsg (take snapshot, clamp cursor, schedule next), key presses (up/down/j/k for cursor, 1/2/3/4 for sort column)
- `SetSize(w, h int) MonitorModel` -- stores dimensions for gauge/table layout
- `View() string` -- renders 4 bordered gauge boxes + sort hint + table header + process rows (cursor row highlighted)
- `sortedProcs() []sysinfo.ProcInfo` -- copies and re-sorts snapshot procs by the selected column
- `gauge(label, value string, width int) string` -- renders a single gauge box

**Note:** The plan called for `evertras/bubble-table`, but it's incompatible with Bubble Tea v2 (type mismatch). The implementation uses a hand-rolled lipgloss table instead. Same visual result, simpler dependency tree.

**Connections:** Imports `sysinfo` (TakeSnapshot, HumanBytes, Snapshot, ProcInfo). References `Styles` from `styles.go`.

---

### `internal/tui/styles.go` -- Theme & Formatting

**Purpose:** Centralized Lipgloss styles and the `HumanSize` formatter.

**Key types:**
- `Styles` -- struct with 12 lipgloss.Style fields: App, TabActive, TabInactive, SearchBar, StatusBar, Preview, Selected, Cursor, DimText, SizeSmall, SizeMedium, SizeLarge

**Key functions:**
- `DefaultStyles() Styles` -- returns the default theme (pink accent #205, gray dims #240, green/yellow/red size colors)
- `HumanSize(b int64) string` -- formats bytes as "1.2 KB", "340.5 MB", etc. Uses int64 (for file sizes)

**Color scheme:**
| Style | Color | Usage |
|---|---|---|
| TabActive, Selected, Cursor | `#205` (pink) | Active tab, selected items, cursor highlight |
| TabInactive, DimText | `#240` (gray) | Inactive elements |
| SizeSmall | `#82` (green) | Files < 1 MB |
| SizeMedium | `#214` (yellow/orange) | Files 1-100 MB |
| SizeLarge | `#196` (red) | Files > 100 MB |
| StatusBar bg | `#236` (dark gray) | Status bar background |

---

### `internal/scanner/scanner.go` -- File Scanner

**Purpose:** Parallel filesystem scanning using fastwalk. The core discovery engine.

**Key types:**
- `Match` -- struct: `Path string`, `IsDir bool`, `Size int64`, `ModTime time.Time`, `Kind string` (one of "file", "dir", "symlink", "rc-line")

**Key functions:**
- `Search(term string, paths []string) ([]Match, error)` -- walks each path using fastwalk, case-insensitive name matching, skips directories in `SkipDirs`, deduplicates results using a `seen` map, thread-safe via `sync.Mutex`
- `DirSize(path string) int64` -- recursively sums file sizes in a directory (used for preview/display, not scanning)

**Thread safety:** fastwalk calls its callback from multiple goroutines. The `results` slice and `seen` map are protected by a `sync.Mutex`. This was a critical bug found during code review.

**Deduplication:** `ScanPaths` has overlapping roots (e.g., `$HOME` includes `$HOME/Library`). The `seen` map prevents the same path from appearing twice.

**Connections:** Uses `fastwalk.Walk`, references `SkipDirs` from `locations.go`.

---

### `internal/scanner/locations.go` -- Scan Paths & Skip Dirs

**Purpose:** Defines WHERE to scan and what to SKIP.

**Key functions/variables:**
- `GetScanPaths() []string` -- returns scan roots, computed at call time using `os.UserHomeDir()`. Roots:
  1. `$HOME`
  2. `$HOME/Library`
  3. `$HOME/.config`
  4. `$HOME/.local`
  5. `/usr/local`
  6. `/opt/homebrew`
  7. `/opt`
- `ScanPaths` -- package-level var for backward compatibility (calls `GetScanPaths()` at init)
- `SkipDirs` -- map of directory names to skip: `node_modules`, `.git`, `.Trash`, `Caches`, `DerivedData`, `CachedData`, `.npm`, `vendor`, `__pycache__`, `cache`, `Cache`

---

### `internal/scanner/rcscanner.go` -- Shell RC Scanner

**Purpose:** Scans shell configuration files for lines matching a search term.

**Key types:**
- `RCMatch` -- struct: `File string`, `LineNum int`, `Line string`

**Key functions/variables:**
- `RCFiles` -- list of RC file basenames: `.zshrc`, `.zshenv`, `.zprofile`, `.bashrc`, `.bash_profile`, `.profile`
- `SearchRC(term string) []RCMatch` -- opens each RC file in `$HOME`, scans line-by-line for case-insensitive match, returns file path + line number + line content

---

### `internal/scanner/scanner_test.go` -- Scanner Tests

**Tests (5 total):**
1. `TestSearchFindsMatchingFiles` -- creates temp dir with foo-app/foo-main.bin/unrelated.txt, verifies >= 2 matches for "foo"
2. `TestSearchSkipsIgnoredDirs` -- creates node_modules/foo-pkg, verifies it's not matched
3. `TestSearchConcurrentSafety` -- creates 50 dirs x 20 files, runs Search 5 times under `-race`
4. `TestSearchDeduplicatesOverlappingPaths` -- searches with overlapping paths [tmp, tmp/sub], verifies no duplicates

### `internal/scanner/rcscanner_test.go` -- RC Scanner Test

**Tests (1):**
1. `TestSearchRCFindsMatchingLines` -- writes temp .zshrc with 4 lines (2 containing "claude"), verifies 2 matches at lines 2 and 4. Uses `t.Setenv` for clean HOME override.

---

### `internal/remover/trash.go` -- macOS Trash

**Purpose:** Moves files to macOS Trash with Put Back support via Finder AppleScript.

**Key functions:**
- `MoveToTrash(absPath string) error` -- executes `osascript -e 'tell application "Finder" to delete POSIX file "/abs/path"'`. Escapes double quotes in paths. Returns wrapped error with stderr on failure.

**Requirements:** Finder must be running. Path must be absolute. Takes ~200ms per file.

---

### `internal/remover/log.go` -- Operation Log

**Purpose:** Appends JSONL entries to `~/.config/gone/operations.log` for audit trail.

**Key types:**
- `LogEntry` -- struct: `Timestamp`, `Op`, `Path`, `Size`, `Kind`, `SearchTerm` (all JSON-tagged)

**Key functions:**
- `logPath() string` -- returns `~/.config/gone/operations.log`
- `AppendLog(entry LogEntry) error` -- sets timestamp (RFC3339) and op ("trash"), creates directory if needed, appends JSON line

### `internal/remover/log_test.go` -- Log Test

**Tests (1):**
1. `TestAppendLog` -- overrides HOME, writes entry, reads log file, verifies JSON fields (path, op)

---

### `internal/sysinfo/sysinfo.go` -- System Metrics

**Purpose:** Wraps gopsutil v4 to collect CPU, memory, swap, disk, and process info.

**Key types:**
- `Snapshot` -- struct: CPUPercent, MemTotal/Used/Avail, SwapTotal/Used, DiskTotal/Used/Free, Procs []ProcInfo
- `ProcInfo` -- struct: PID, Name, CPU, Mem (float32), RSS (uint64)

**Key functions:**
- `TakeSnapshot(topN int) Snapshot` -- collects all metrics in one call. Processes sorted by CPU% descending, capped at topN.
- `HumanBytes(b uint64) string` -- formats bytes as "1.2 KB", etc. Uses uint64 (for memory sizes)

**Gotcha:** First `cpu.Percent(0, false)` call returns 0 (seeds baseline). The monitor's 2-second tick naturally handles this -- the first refresh shows 0, subsequent refreshes show real values.

### `internal/sysinfo/sysinfo_test.go` -- Sysinfo Test

**Tests (1):**
1. `TestTakeSnapshotReturnsData` -- verifies non-zero MemTotal, DiskTotal, and at least one process

---

### `go.mod` -- Module Definition

```
module gone
go 1.26.1

require (
    charm.land/bubbles/v2 v2.1.0
    charm.land/bubbletea/v2 v2.0.5
    charm.land/lipgloss/v2 v2.0.3
    github.com/charlievieth/fastwalk v1.0.14
    github.com/shirou/gopsutil/v4 v4.26.3
)
```

**Notable:** Charm libraries use vanity domain imports (`charm.land/...`), NOT `github.com/charmbracelet/...`. This is a v2 change. If the vanity domain ever fails, fall back to the GitHub paths with v1 API adjustments (see plan header).

---

## 5. Design Decisions

### fastwalk over filepath.WalkDir

**Choice:** `github.com/charlievieth/fastwalk` instead of the stdlib `filepath.WalkDir`.

**Why:** fastwalk is 4-5x faster for large directory trees. It uses multiple goroutines (auto-capped at 4 on macOS APFS). For scanning `$HOME` with 100k-500k entries, this means <2s instead of 5-10s. The tradeoff is concurrent callback handling (requires mutex), which we handle.

### Hand-rolled table over evertras/bubble-table

**Choice:** Manual lipgloss table rendering in `monitor.go` instead of `evertras/bubble-table`.

**Why:** `evertras/bubble-table` targets Bubble Tea v1 and produces type incompatibilities with v2's `tea.Model` interface (`View()` returns `string` in v1, `tea.View` in v2). Rather than fight the type system or pin to v1, the workers built a simple text table with `sort.Slice`. For 15 rows, this is more than adequate and eliminates a dependency.

### osascript for trash (not os.Remove)

**Choice:** Send files to Trash via `osascript -e 'tell application "Finder" to delete POSIX file "..."'` instead of hard-deleting.

**Why:** This is the ONLY way to get macOS Trash "Put Back" functionality. When you delete via `os.Remove()` or even `mv` to `~/.Trash/`, the Finder doesn't record the original location. The AppleScript approach tells Finder to handle the move, preserving the undo path. The cost is ~200ms per file and requiring Finder to be running.

### Lipgloss v2 over v1

**Choice:** `charm.land/lipgloss/v2` with vanity domain imports.

**Why:** v2 is the actively maintained version. It has cleaner APIs for borders, padding, and layout. The Charm team moved to vanity domain imports (`charm.land/...`) starting with v2. The key API differences from v1: `Width()` is a method on textinput (use `SetWidth()`), `View()` returns `tea.View` not `string`, `WithAltScreen()` removed (use `v.AltScreen = true`), `viewport.New()` takes option funcs.

### No Cobra/CLI framework

**Choice:** Direct `tea.NewProgram()` call in main. No subcommands, no flags.

**Why:** `gone` is a single-purpose TUI. There's nothing to parse. The binary opens the TUI. If you want to add a CLI scanner mode later, a simple `os.Args` check is sufficient.

### No recipes/database

**Choice:** Pure filesystem scan + name matching. No config files describing what each tool installed.

**Why:** Recipes go stale. Every tool changes its install layout between versions. Filesystem scanning is always current. The user sees every match and decides. This is the "user takes responsibility" philosophy.

### `bubbles/list` with custom delegate (not a raw viewport)

**Choice:** Use the Bubbles list component with a custom single-line delegate for the file list.

**Why:** Bubbles list provides built-in fuzzy filtering (via `sahilm/fuzzy`), keyboard navigation, pagination, and focus management. The custom delegate adds checkboxes (`[x]`/`[ ]`) and color-coded sizes. Building all this from scratch would be hundreds of lines.

### JSONL operation log (not SQLite)

**Choice:** Append-only JSONL file at `~/.config/gone/operations.log`.

**Why:** Simple, greppable, no dependencies. Each line is a self-contained JSON object. Easy to pipe through `jq`. No schema migrations. For a personal audit trail, this is the right tradeoff.

### `sync.Mutex` in scanner (not channels)

**Choice:** Protect the shared `results` slice with a mutex instead of using a channel for collecting results.

**Why:** fastwalk's callback API doesn't support channel-based collection cleanly. The callback must return an error to control traversal (e.g., `filepath.SkipDir`). Using a mutex to protect `append()` is the simplest correct approach and matches fastwalk's documented patterns.

---

## 6. How to Run

### Prerequisites

- **macOS** (uses osascript, Finder, APFS-optimized fastwalk)
- **Go 1.26+** (`go version` to check)
- **Finder.app** must be running (for trash functionality)

### Build & Run

```bash
cd /Users/agustin/Developments/personal/scripts/gone

# Run directly
go run ./cmd/gone

# Build binary
go build -o gone ./cmd/gone
./gone

# Install to $GOPATH/bin
go install ./cmd/gone
# Then: gone
```

### Run Tests

```bash
cd /Users/agustin/Developments/personal/scripts/gone

# All tests
go test ./...

# Verbose
go test ./... -v

# With race detector
go test -race ./...

# Specific package
go test ./internal/scanner/ -v
```

### Verify Build

```bash
go build ./...    # compile all packages
go vet ./...      # static analysis
go test -race ./... # race detector
```

### Usage

1. Launch: `go run ./cmd/gone` or `./gone`
2. **Uninstall tab** (default):
   - Type a search term, press Enter
   - Wait for scan (spinner shows progress)
   - Navigate results with arrow keys
   - Press `/` to fuzzy-filter the list
   - Press Space to toggle selection
   - Press Enter to send selected items to Trash
   - Press Esc to return to search bar
3. **Monitor tab**:
   - Press Tab to switch
   - View CPU/RAM/Swap/Disk gauges (auto-refresh 2s)
   - Navigate process table with arrow keys or j/k
   - Press 1/2/3/4 to sort by CPU/Mem/RSS/PID
4. Press `?` for help overlay
5. Press `Ctrl+C` to quit

### Operation Log

After trashing files, check the log:
```bash
cat ~/.config/gone/operations.log | jq .
```

Each line is a JSON object:
```json
{"ts":"2026-04-16T12:00:00Z","op":"trash","path":"/Users/you/.claude/config","size":1234,"kind":"file","term":"claude"}
```

---

## 7. How to Extend

### Adding New Scan Paths

Edit `internal/scanner/locations.go`, add to `GetScanPaths()`:

```go
func GetScanPaths() []string {
    home, err := os.UserHomeDir()
    if err != nil {
        home = os.Getenv("HOME")
    }
    return []string{
        home,
        filepath.Join(home, "Library"),
        filepath.Join(home, ".config"),
        filepath.Join(home, ".local"),
        "/usr/local",
        "/opt/homebrew",
        "/opt",
        "/Applications",                          // NEW
        filepath.Join(home, ".cache"),             // NEW
    }
}
```

**Note:** Adding overlapping paths is safe -- the `seen` map in `Search()` deduplicates.

To add new skip directories, add to the `SkipDirs` map:

```go
var SkipDirs = map[string]bool{
    // ... existing ...
    ".cargo":       true,  // NEW
    ".rustup":      true,  // NEW
}
```

### Adding a New Tab

1. **Create the model** in `internal/tui/newtab.go`:

```go
package tui

type NewTabModel struct {
    // fields
    width, height int
}

func NewNewTabModel() NewTabModel { return NewTabModel{} }
func (m NewTabModel) Init() tea.Cmd { return nil }
func (m NewTabModel) Update(msg tea.Msg) (NewTabModel, tea.Cmd) { return m, nil }
func (m NewTabModel) SetSize(w, h int) NewTabModel { m.width = w; m.height = h; return m }
func (m NewTabModel) View() string { return "New tab content" }
```

2. **Register in `app.go`**:

```go
const (
    tabUninstall activeTab = iota
    tabMonitor
    tabNew        // ADD
)

type AppModel struct {
    // ...
    newTab NewTabModel  // ADD
}

func NewApp() AppModel {
    return AppModel{
        // ...
        newTab: NewNewTabModel(),  // ADD
    }
}
```

3. **Route messages** in `AppModel.Update()`:

```go
if msg.String() == "tab" {
    m.active = (m.active + 1) % 3  // CHANGE from 2 to 3
}
```

4. **Render** in `AppModel.View()`:

```go
// Add tab label
newTabLabel := " New "
// ... style it based on m.active ...

// Add content case
case tabNew:
    b.WriteString(m.newTab.View())
```

### Changing the Theme

Edit `internal/tui/styles.go`. All colors use ANSI 256-color codes:

```go
func DefaultStyles() Styles {
    return Styles{
        TabActive:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),  // blue
        SizeSmall:   lipgloss.NewStyle().Foreground(lipgloss.Color("48")),              // teal
        // ... etc
    }
}
```

For true color support, use hex codes:
```go
lipgloss.Color("#FF6B6B")  // coral red
```

### Adding New RC Files to Scan

Edit `internal/scanner/rcscanner.go`:

```go
var RCFiles = []string{
    ".zshrc", ".zshenv", ".zprofile",
    ".bashrc", ".bash_profile", ".profile",
    ".config/fish/config.fish",    // NEW: fish shell
    ".config/nushell/config.nu",   // NEW: nushell
}
```

### Adding a New Tool (e.g., hard-delete mode)

1. Add to `internal/remover/`:

```go
// harddelete.go
func HardDelete(absPath string) error {
    return os.RemoveAll(absPath)
}
```

2. Wire into `uninstall.go` -- add a mode toggle, modify `trashSelected()`.

---

## 8. Orchestrator

### What Is `supervisor.ts`?

`gone/orchestrator/supervisor.ts` is a ~325-line Bun/TypeScript script that automates the build process by dispatching Claude Code workers as subprocesses. Each worker implements exactly one task from the plan, then stops.

**Location:** `gone/orchestrator/supervisor.ts`
**Runtime:** Bun (`bun run orchestrator/supervisor.ts`)
**Log output:** `gone/orchestrator/logs/sessions.jsonl`

### How It Works

```
Supervisor (TypeScript/Bun)
    │
    │ for each task in [0..8]:
    │   1. Build the claude CLI command
    │   2. Strip CLAUDECODE/CLAUDE_CODE_ENTRYPOINT env vars
    │   3. Spawn: claude --system-prompt "..." -p --output-format json "prompt"
    │   4. Pipe stdout/stderr
    │   5. Parse JSON output (session_id, result)
    │   6. Log to sessions.jsonl
    │   7. If exit != 0: STOP, print resume instructions
    │   8. If exit == 0: print summary, continue to next task
    │
    └── After all tasks: print summary table
```

### Configuration Constants

| Constant | Value | Purpose |
|---|---|---|
| `PLAN_FILE` | `docs/superpowers/plans/2026-04-15-gone.md` | Plan each worker reads |
| `KNOWLEDGE_BASE` | `docs/superpowers/specs/2026-04-15-gone-research.md` | Verified code patterns |
| `MODEL` | `sonnet` | Claude model for workers |
| `EFFORT` | `high` | Extended thinking enabled |
| `MAX_BUDGET_PER_TASK` | `$2` | Per-worker spend cap |
| `ALLOWED_TOOLS` | `Edit,Write,Bash,Read,Glob,Grep` | Tool whitelist |
| `TOTAL_TASKS` | `9` | Tasks 0-8 |

### CLI Flags

```bash
# Run all tasks 0-8
bun run orchestrator/supervisor.ts

# Start from task 3 (0-2 already done)
bun run orchestrator/supervisor.ts --from 3

# Run up to task 5
bun run orchestrator/supervisor.ts --to 5

# Run specific tasks
bun run orchestrator/supervisor.ts --only 0,1,2

# Preview commands without executing
bun run orchestrator/supervisor.ts --dry-run

# Debug task 5 interactively (live terminal, no -p flag)
bun run orchestrator/supervisor.ts --interactive 5
```

### System Prompt (per worker)

Each worker gets a strict system prompt via `--system-prompt`:

1. You are a Go developer implementing "gone"
2. Read the plan file FIRST
3. Read the knowledge base
4. Implement ONLY Task N
5. Follow steps IN ORDER
6. If charm.land imports fail, apply v1 fallback
7. Run verify command
8. Fix failing tests before committing
9. Commit with the plan's message
10. Update CHANGELOG.md
11. **STOP after Task N is complete. Do NOT start the next task.**

The explicit stop condition (#11) is critical. Without it, workers will attempt all remaining tasks.

### Session Logging

Every task run appends a line to `orchestrator/logs/sessions.jsonl`:

```json
{"task":3,"name":"Uninstall TUI","success":true,"sessionId":"abc123","duration":187,"ts":"2026-04-16T..."}
```

### Resuming Failed Tasks

When a task fails, the supervisor prints:
```
*** Task 3 FAILED. Stopping. ***
Resume with: bun run orchestrator/supervisor.ts --from 3
Debug session: env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude --resume "abc123xyz"
```

### How to Adapt for Your Project

1. Change `PLAN_FILE` and `KNOWLEDGE_BASE` paths
2. Update `TASK_NAMES` with your task list
3. Update `TOTAL_TASKS`
4. Rewrite `systemPrompt(taskNum)` for your domain
5. Adjust `MODEL`, `EFFORT`, `MAX_BUDGET_PER_TASK`
6. Adjust `ALLOWED_TOOLS`
7. Change `cwd` in the spawn call

---

## 9. Orchestration Methodology

### The Discovery

During this session, we discovered and documented a reusable methodology for building complex projects with Claude Code. The key insight: **one fresh worker per task, coordinated by a dumb supervisor, with plan files as the bridge between sessions.**

### The Problem with Single Sessions

Claude Code breaks down on large projects in three ways:

1. **Context overflow.** A project with 9 tasks + research + plan + codebase overflows the context window before task 6. The agent starts hallucinating.
2. **Memory pollution.** Long sessions accumulate failed attempts, backtracked decisions. Each new task carries the residue of every previous task.
3. **Plan drift.** Given a 9-step plan, the agent tries to implement all 9 at once, skipping checkpoints.

### The Solution: Plan Files as Bridges

```
Plan File (on disk)     ←── shared state, never in any context window
Knowledge Base (on disk) ←── verified patterns, not hallucinated
Supervisor (TypeScript)  ←── dumb dispatcher, reads task list
Worker 1 (fresh session) ←── reads plan, does Task 0, commits, dies
Worker 2 (fresh session) ←── reads plan, does Task 1, commits, dies
...
Worker 9 (fresh session) ←── reads plan, does Task 8, commits, dies
```

**Workers communicate through artifacts (files, commits), not through shared memory.**

### Key Techniques

#### 1. Stripping Claude's Nesting Guard

Claude Code sets `CLAUDECODE=1` and `CLAUDE_CODE_ENTRYPOINT=cli` to prevent nesting. Strip them before spawning workers:

```bash
env -u CLAUDECODE -u CLAUDE_CODE_ENTRYPOINT claude [flags] "prompt"
```

In TypeScript:
```typescript
const env = { ...process.env };
delete env.CLAUDECODE;
delete env.CLAUDE_CODE_ENTRYPOINT;
```

#### 2. Non-Interactive Mode with JSON Output

```bash
claude -p --output-format json "prompt"
```

Returns: `{"result": "...", "session_id": "abc123", "cost_usd": 1.5, "duration_ms": 120000}`

#### 3. Strict System Prompts with Stop Conditions

```
--system-prompt "You are a worker. Do Task 3 ONLY. STOP after Task 3."
```

Without the explicit stop condition, workers drift and attempt all tasks.

#### 4. Permission Bypass for Automation

```
--allowedTools Edit,Write,Bash,Read,Glob,Grep --dangerously-skip-permissions
```

Workers need to run `go build`, `go test`, `git commit`. The permission prompts would block automation.

#### 5. Budget Caps

```
--max-budget-usd 2
```

Prevents runaway tasks from burning through credits.

#### 6. stdin Handling

Pass `stdio: ["ignore", "pipe", "pipe"]` when spawning. `claude -p` with inherited stdin produces a warning.

### Stream-JSON Bidirectional Communication

`--input-format stream-json` enables a different model where the orchestrator can inject messages mid-session:

```
orchestrator stdout → worker stdin  (JSON messages)
worker stdout       → orchestrator  (JSON events)
```

This is more complex but enables real-time steering.

### Mini-Agent as Alternative

[Mini-Agent](https://github.com/MiniMax-AI/Mini-Agent) is a Python ReAct-loop agent framework. Use it when you need a **reasoning orchestrator** that decides what to do next, not just a sequential dispatcher. See `docs/superpowers/specs/2026-04-16-mini-agent-research.md` for the full analysis.

**When supervisor.ts is better:** Fixed sequential task lists.
**When Mini-Agent is better:** Dynamic task routing, adaptive plans, interactive steering.

---

## 10. Future Work

### 10.1 Stream-JSON Persistent Sessions

Currently each worker is a fresh session. Using `--input-format stream-json` with `--resume`, the orchestrator could maintain persistent conversations with workers, injecting follow-up instructions based on output.

### 10.2 Mini-Agent Integration

Build a custom `ClaudeWorkerTool` for Mini-Agent that spawns claude CLI workers. The orchestrator LLM would reason about which task to run next, retry failed tasks with modified prompts, and adapt the plan dynamically.

### 10.3 Recipes System

A `~/.config/gone/recipes/` directory with YAML files describing known tool layouts:

```yaml
# claude.yaml
name: Claude Code
patterns:
  - ~/.claude/
  - ~/Library/Application Support/Claude/
  - /usr/local/bin/claude
  - .zshrc: "claude"
```

The scanner would use recipes as hints but still verify with filesystem scanning.

### 10.4 Audit/Orphan Mode

A mode that scans for "orphaned" config directories -- dirs in `~/.config/` or `~/Library/` that don't correspond to any installed application. Would require comparing against `/Applications/` and `brew list`.

### 10.5 Brew Formula

Package `gone` as a Homebrew formula:

```ruby
class Gone < Formula
  desc "macOS TUI for finding and removing leftover dev tool files"
  homepage "https://github.com/user/gone"
  url "..."
  sha256 "..."
  depends_on :macos
  def install
    system "go", "build", "-o", bin/"gone", "./cmd/gone"
  end
end
```

### 10.6 .app Bundle Scanning

Extend scanning to `/Applications/*.app/Contents/` and `~/Applications/`. Would need to handle `.app` bundles as atomic units (don't scan inside unless explicitly requested).

### 10.7 Per-Task Model/Effort Configuration

Implement a `TASK_CONFIG` map in supervisor.ts for tiered model assignment:

```typescript
const TASK_CONFIG: Record<number, {model: string; effort: string; budget: string}> = {
  0: { model: "haiku",  effort: "low",  budget: "0.25" },  // trivial scaffold
  3: { model: "sonnet", effort: "high", budget: "5" },     // complex TUI
  5: { model: "sonnet", effort: "low",  budget: "1" },     // straightforward
};
```

### 10.8 Parallel Workers with Worktrees

Some tasks have no dependencies. Use git worktrees for parallel execution:

```bash
git worktree add /tmp/gone-task-1 main
git worktree add /tmp/gone-task-2 main
# spawn two workers concurrently
# merge results when both complete
```

### 10.9 Interactive Preview with File Content

Show actual file contents in the preview pane for text files (not just metadata). Use syntax highlighting via lipgloss or a dedicated library.

### 10.10 Undo Support

Use the operation log to implement an undo command that moves items back from Trash to their original locations.

---

## 11. Continuation Prompt

Copy-paste this prompt into a new Claude Code session to pick up where this project left off:

```
I'm continuing work on the "gone" project — a macOS Go TUI for finding and trashing leftover dev tool files, with a live system monitor tab.

## Project Location
Root: /Users/agustin/Developments/personal/scripts
Go module: /Users/agustin/Developments/personal/scripts/gone

## Current State: COMPLETE (v1.0)
All 9 implementation tasks are done. Build passes, 8 tests pass, race detector clean, vet clean.

## Key Files to Read First
1. Master doc: docs/superpowers/specs/2026-04-16-gone-master-doc.md (THIS FILE — full architecture, every component, design decisions)
2. Design spec: docs/superpowers/specs/2026-04-15-gone-design.md
3. Verified patterns: docs/superpowers/specs/2026-04-15-gone-research.md
4. Implementation plan: docs/superpowers/plans/2026-04-15-gone.md
5. Code review: docs/superpowers/specs/2026-04-16-gone-code-review.md
6. QA report: docs/superpowers/specs/2026-04-16-gone-qa-report.md
7. Orchestration guide: docs/superpowers/specs/2026-04-16-orchestration-guide.md
8. CHANGELOG: gone/CHANGELOG.md

## Source Files
- gone/cmd/gone/main.go — entry point
- gone/internal/tui/app.go — root model, tab routing, help overlay
- gone/internal/tui/uninstall.go — search/scan/list/preview/trash flow
- gone/internal/tui/monitor.go — system dashboard + process table
- gone/internal/tui/styles.go — Lipgloss theme + HumanSize
- gone/internal/scanner/scanner.go — fastwalk file finder (mutex-protected)
- gone/internal/scanner/rcscanner.go — shell RC line scanner
- gone/internal/scanner/locations.go — scan paths + skip dirs
- gone/internal/remover/trash.go — osascript Finder trash
- gone/internal/remover/log.go — JSONL operation log
- gone/internal/sysinfo/sysinfo.go — gopsutil wrapper
- gone/orchestrator/supervisor.ts — Bun task dispatcher

## Tech Stack
- Go 1.26.1, Bubble Tea v2 (charm.land/bubbletea/v2), Bubbles v2, Lipgloss v2
- fastwalk (parallel dir walker), gopsutil v4 (system metrics)
- osascript for macOS Trash with Put Back

## What's Done
- [x] Scaffold + hello world (Task 0)
- [x] fastwalk file scanner with mutex + dedup (Task 1)
- [x] Shell RC scanner (Task 2)
- [x] Uninstall TUI: search bar, spinner, list, multi-select (Task 3)
- [x] Preview pane with split layout (Task 4)
- [x] Trash via Finder + JSONL log (Task 5)
- [x] System monitor: gauges + sortable process table (Task 6)
- [x] Root model + tab switching + tick routing (Task 7)
- [x] Polish: color-coded sizes, help overlay (Task 8)
- [x] Code review: data race fixed, dedup added
- [x] QA: 8 issues fixed, 2 new tests added

## What's Pending (see Future Work in master doc)
- Recipes system
- Audit/orphan mode
- Brew formula
- .app bundle scanning
- Stream-JSON persistent sessions
- Mini-Agent integration
- Interactive file content preview
- Undo support

## Verify Everything Works
cd /Users/agustin/Developments/personal/scripts/gone
go build ./... && go test -race ./... && go vet ./...
```

---

## Appendix A: Complete File Tree

```
/Users/agustin/Developments/personal/scripts/
├── docs/superpowers/
│   ├── plans/
│   │   └── 2026-04-15-gone.md                     # Implementation plan (9 tasks)
│   └── specs/
│       ├── 2026-04-15-gone-summary.md              # Session summary
│       ├── 2026-04-15-gone-design.md               # Design spec
│       ├── 2026-04-15-gone-research.md             # Verified code patterns
│       ├── 2026-04-16-gone-code-review.md          # Code review report
│       ├── 2026-04-16-gone-qa-report.md            # QA report
│       ├── 2026-04-16-gone-master-doc.md           # THIS FILE
│       ├── 2026-04-16-mini-agent-research.md       # Mini-Agent analysis
│       └── 2026-04-16-orchestration-guide.md       # Orchestration methodology
└── gone/
    ├── cmd/gone/
    │   └── main.go                                 # Entry point (17 lines)
    ├── internal/
    │   ├── scanner/
    │   │   ├── locations.go                        # Scan paths + skip dirs
    │   │   ├── scanner.go                          # fastwalk file finder
    │   │   ├── scanner_test.go                     # 4 tests
    │   │   ├── rcscanner.go                        # Shell RC scanner
    │   │   └── rcscanner_test.go                   # 1 test
    │   ├── remover/
    │   │   ├── trash.go                            # osascript trash
    │   │   ├── log.go                              # JSONL log
    │   │   └── log_test.go                         # 1 test
    │   ├── sysinfo/
    │   │   ├── sysinfo.go                          # gopsutil wrapper
    │   │   └── sysinfo_test.go                     # 1 test
    │   └── tui/
    │       ├── app.go                              # Root model + tabs
    │       ├── uninstall.go                        # Search/scan/list/trash
    │       ├── monitor.go                          # System dashboard
    │       └── styles.go                           # Theme + HumanSize
    ├── orchestrator/
    │   ├── supervisor.ts                           # Bun task dispatcher
    │   └── logs/
    │       └── sessions.jsonl                      # Worker session log
    ├── go.mod
    ├── go.sum
    └── CHANGELOG.md
```

## Appendix B: Test Matrix

| Package | Test | What It Verifies |
|---|---|---|
| scanner | `TestSearchFindsMatchingFiles` | Finds dirs and files by name substring |
| scanner | `TestSearchSkipsIgnoredDirs` | node_modules, .git, etc. are skipped |
| scanner | `TestSearchConcurrentSafety` | No data race under parallel fastwalk |
| scanner | `TestSearchDeduplicatesOverlappingPaths` | Overlapping scan roots don't produce dupes |
| scanner | `TestSearchRCFindsMatchingLines` | RC scanner finds correct lines + line numbers |
| remover | `TestAppendLog` | JSONL log created with correct fields |
| sysinfo | `TestTakeSnapshotReturnsData` | Returns non-zero metrics from live system |

## Appendix C: Known Spec Deviations

| Spec Item | Actual Implementation | Assessment |
|---|---|---|
| `evertras/bubble-table` for process list | Hand-rolled table with `sort.Slice` | Equivalent; avoids v1/v2 incompatibility |
| Top 10 processes | Top 15 processes | Slightly more data |
| `tea.WithAltScreen()` option | `v.AltScreen = true` on tea.View | v2 API change; equivalent behavior |
| `textinput.Width = w` assignment | `textinput.SetWidth(w)` method call | v2 API change; equivalent behavior |
| `viewport.New(w, h)` constructor | `viewport.New()` + `SetWidth`/`SetHeight` | v2 API change; equivalent behavior |
