# gone — macOS Uninstaller & System Monitor TUI

> Knowledge base: [gone-research.md](./2026-04-15-gone-research.md) — all verified code patterns, imports, gotchas.

## What

Single Go binary. Two modes in a tabbed TUI:

1. **Uninstall** — type a word, see every file/dir/rc-line matching it across macOS, multi-select, trash.
2. **Monitor** — live CPU/RAM/swap/disk + top 10 processes by CPU and RAM, sortable table.

Personal tool. macOS only. No recipes, no warnings, no dry-run UI. User decides what to delete.

## Stack

| Dep | Purpose |
|---|---|
| `charm.land/bubbletea/v2` | TUI framework |
| `charm.land/bubbles/v2` | list (fuzzy filter + multi-select), spinner, viewport, textinput |
| `charm.land/lipgloss/v2` | styling, layout |
| `evertras/bubble-table` | sortable process table |
| `shirou/gopsutil/v4` | CPU, RAM, swap, disk, processes |
| `charlievieth/fastwalk` | parallel filesystem walker |

## Build Order (Karpathy method)

Each step produces a runnable binary. Test it before moving on.

### Step 0 — Scaffold + hello world

- `go mod init gone`
- `cmd/gone/main.go`: minimal Bubble Tea app that prints terminal size and quits on `q`
- **Verify:** `go run ./cmd/gone` shows `WxH`, resizing updates it, `q` exits

### Step 1 — Scanner

- `internal/scanner/scanner.go`: takes a search term, walks `scanPaths` via fastwalk, returns `[]Match{Path, IsDir, Size, ModTime}`
- `internal/scanner/locations.go`: define `scanPaths` and `skipDirs`
- Wire it into main as a CLI test: `go run ./cmd/gone claude` prints matches to stdout
- **Verify:** finds `~/.claude`, `/usr/local/bin/claude`, etc. Runs in <3s

### Step 2 — RC scanner

- `internal/scanner/rcscanner.go`: scans shell rc files for lines matching the term
- Returns `[]RCMatch{File, LineNum, Line}`
- Merge into scanner output
- **Verify:** `go run ./cmd/gone claude` also shows `.zshrc:42 export CLAUDE_...`

### Step 3 — Uninstall TUI (list + filter)

Search flow: TUI opens → user types search term in textinput → hits Enter → spinner + scanner runs → results fill list → user navigates/filters with `/` (fuzzy, built-in) → Space toggles select → Enter trashes selected.

- `internal/tui/uninstall.go`: Bubble Tea model with:
  - `textinput` at top (search bar, focused on start)
  - `spinner` while scanning (appears after user hits Enter)
  - `bubbles/list` with custom single-line delegate showing `[x] path  size  date`
  - Space toggles selection, Enter on list triggers trash, Enter on textinput triggers scan
  - Esc returns focus to textinput for a new search
- `internal/tui/styles.go`: Lipgloss theme struct
- **Verify:** `go run ./cmd/gone` opens TUI, type `claude`, Enter, spinner, results appear, space selects, q quits

### Step 4 — Preview pane

- Add `viewport` to right side of uninstall view
- Split layout: 50/50 (adjust as needed)
- Preview shows: full path, size, last modified, type (file/dir/symlink), contents count if dir
- **Verify:** arrow up/down changes preview content

### Step 5 — Trash + log

- `internal/remover/trash.go`: `MoveToTrash(absPath)` via osascript
- `internal/remover/log.go`: append JSONL to `~/.config/gone/operations.log`
- Wire Enter key → trash selected items → show summary line at bottom
- **Verify:** select items, Enter, check they're in macOS Trash with Put Back working. Check log file exists.

### Step 6 — Monitor tab

- `internal/tui/monitor.go`: Bubble Tea model with:
  - Top section: CPU%, RAM (used/total), Swap (used/total), Disk (used/free)
  - Bottom section: `evertras/bubble-table` with PID, Name, CPU%, MEM%, RSS — sortable
  - 2-second refresh via `tea.Tick`
- `internal/sysinfo/sysinfo.go`: wraps gopsutil calls, returns `Snapshot` struct
- **Verify:** `go run ./cmd/gone` → Tab key switches to monitor, shows live data, sorts work

### Step 7 — Root model + tabs

- `internal/tui/app.go`: root model with `activeTab`, routes messages
- Tab bar at top: `[ Uninstall | Monitor ]` with Lipgloss active/inactive styles
- Tab key switches tabs
- Route tick messages to monitor sub-model ALWAYS (even when uninstall tab is active)
- **Verify:** both tabs work, switching is smooth, monitor doesn't freeze

### Step 8 — Polish

- Color-coded sizes (green < 1MB, yellow < 100MB, red > 100MB)
- Status bar at bottom (item count, total selected size, keybind hints)
- Smooth transitions: spinner fades when scan completes
- Help overlay on `?` key (simple lipgloss.Place centered box)

## Project Layout

```
scripts/gone/
├── cmd/gone/main.go
├── internal/
│   ├── scanner/
│   │   ├── scanner.go      # fastwalk-based file finder
│   │   ├── rcscanner.go    # shell rc line matcher
│   │   └── locations.go    # scanPaths, skipDirs
│   ├── remover/
│   │   ├── trash.go        # osascript Finder trash
│   │   └── log.go          # JSONL operation log
│   ├── sysinfo/
│   │   └── sysinfo.go      # gopsutil wrapper → Snapshot
│   └── tui/
│       ├── app.go           # root model + tab routing
│       ├── uninstall.go     # search + list + preview
│       ├── monitor.go       # dashboard + process table
│       └── styles.go        # Lipgloss theme
├── go.mod
└── go.sum
```

## Key Decisions

- **No Cobra/CLI framework.** Single binary, `go run ./cmd/gone` opens TUI. No subcommands needed.
- **No recipes.** Pure filesystem scan + name matching. User decides what's relevant.
- **Trash only, never hard-delete.** Via Finder AppleScript for Put Back support.
- **No dry-run UI.** The list IS the preview. User sees everything before pressing Enter.
- **fastwalk over filepath.WalkDir.** 4-5x faster, darwin-optimized.
- **evertras/bubble-table over bubbles/table.** Need sorting for process list.
