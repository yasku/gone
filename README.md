# gone

A macOS TUI for hunting down and removing every last trace of uninstalled tools — caches, configs, shell RC lines, the works. Plus a live system monitor because why not.

Built with Go, [Bubble Tea v2](https://github.com/charmbracelet/bubbletea), and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## What it does

**Uninstall tab** — type a name, hit Enter, and `gone` scans your entire system for matching files, directories, and shell RC references. Select what to trash with Space, preview details in the side pane, hit Enter to send them to macOS Trash (with Put Back support).

**Monitor tab** — live CPU, RAM, swap, and disk gauges with a sortable process table. Press `1`-`4` to sort by CPU/Mem/RSS/PID.

## Install

```bash
cd gone && go build -o gone ./cmd/gone
./gone
```

## Keys

| Key | Action |
|-----|--------|
| `Tab` | Switch between Uninstall / Monitor |
| `Enter` | Scan (in search) · Trash selected (in list) |
| `Space` | Toggle selection |
| `/` | Filter results |
| `Esc` | Back to search input |
| `?` | Help overlay |
| `q` · `Ctrl+C` | Quit |

## How it finds things

- Parallel filesystem walk ([fastwalk](https://github.com/charlievieth/fastwalk)) across `~/Library`, `/usr/local`, `~/.config`, `~/.local`, `/opt`, and more
- Shell RC scanner — checks `.zshrc`, `.bashrc`, `.bash_profile`, `.profile`, `.zshenv`, `.zprofile` for matching lines
- Files go to macOS Trash via Finder AppleScript (not `rm`) — you can always Put Back

## Tech

| | |
|---|---|
| Language | Go 1.26 |
| TUI | Bubble Tea v2 + Bubbles v2 |
| Styling | Lipgloss v2 |
| Filesystem | charlievieth/fastwalk |
| System metrics | gopsutil v4 |
| Trash | osascript + Finder |

## Project Structure

```
gone/
├── cmd/gone/main.go           # Entry point
├── internal/
│   ├── scanner/
│   │   ├── scanner.go         # Parallel file scanner
│   │   ├── locations.go       # Scan paths & skip dirs
│   │   └── rcscanner.go       # Shell RC file scanner
│   ├── remover/
│   │   ├── trash.go           # macOS Trash via osascript
│   │   └── log.go             # JSONL operation log
│   ├── sysinfo/
│   │   └── sysinfo.go         # gopsutil wrapper
│   └── tui/
│       ├── app.go             # Root model + tab routing
│       ├── uninstall.go       # Search → scan → select → trash
│       ├── monitor.go         # Live gauges + process table
│       └── styles.go          # Lipgloss theme
└── orchestrator/
    └── supervisor.ts          # Build orchestrator (Bun/TS)
```

## Collaborators

| | |
|---|---|
| **Agustin** | Creator, designer, orchestrator wrangler |
| **Claude Opus 4.6** | Code, architecture, research, bugfixes — **MAD MAX** |
