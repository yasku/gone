<div align="center">

<img src="assets/banner.png" alt="gone" width="100%" />

<br>
<br>

A macOS TUI that hunts down every trace of uninstalled tools<br>
and sends them where they belong.

<br>

[Install](#install) · [How it works](#how-it-works) · [Keys](#keys) · [Stack](#stack)

<br>

---

</div>

<br>

## The problem

You drag an app to Trash. macOS says it's gone.

It's not.

`~/Library/Caches` · `~/Library/Application Support` · `~/.config` · `/usr/local` · shell RC exports · PATH modifications — hundreds of megabytes of ghost data from tools you deleted months ago. Still there. Still taking space. Still polluting your shell.

<br>

## How it works

```
  gone                                    Uninstall · Monitor
 ─────────────────────────────────────────────────────────────

  Search: claude_

  ┌─ Results ──────────────────┐  ┌─ Preview ──────────────┐
  │                            │  │                        │
  │  ● ~/Library/Caches/clau…  │  │  Type       directory  │
  │    ~/Library/Logs/claude…  │  │  Size       48.2 MB    │
  │  ● ~/.config/claude/       │  │  Modified   2 days ago │
  │    ~/.zshrc :14            │  │                        │
  │                            │  │  ├── config.json       │
  │                            │  │  ├── credentials       │
  │                            │  │  └── sessions/         │
  │                            │  │                        │
  └────────────────────────────┘  └────────────────────────┘

  2 selected · 48.6 MB                           [?] help
```

Type a name. Hit Enter. Select what to remove. Trash it.

Files go to macOS Trash via Finder AppleScript — not `rm`. You can always **Put Back**.

<br>

## Install

```bash
go build -o gone ./cmd/gone
./gone
```

<br>

## Keys

| | |
|:--|:--|
| `Tab` | Switch between Uninstall and Monitor |
| `Enter` | Scan (in search) · Trash selected (in list) |
| `Space` | Toggle selection |
| `/` | Filter results |
| `Esc` | Back to search |
| `?` | Help overlay |
| `q` | Quit |

<br>

## What it scans

Parallel filesystem walk via [fastwalk](https://github.com/charlievieth/fastwalk) across every location where macOS tools leave traces:

```
~/Library/Caches                App caches
~/Library/Application Support   App data, configs
~/Library/Preferences           Plist files
~/Library/Logs                  App logs
~/.config                       XDG configs
~/.local                        User binaries, data
/usr/local                      Homebrew, manual installs
/opt                            System packages
```

Plus shell RC files — `.zshrc`, `.bashrc`, `.profile`, `.zshenv`, `.zprofile`, `.bash_profile` — scanned line by line for matching exports, PATH entries, and aliases.

Results are size-coded: **green** under 1 MB · **yellow** under 100 MB · **red** over 100 MB.

<br>

## Monitor

The second tab. Live system dashboard with real-time gauges for CPU, memory, swap, and disk usage. Sortable process table underneath.

| | |
|:--|:--|
| `1` | Sort by CPU |
| `2` | Sort by Memory |
| `3` | Sort by RSS |
| `4` | Sort by PID |

<br>

## Stack

| | |
|:--|:--|
| Go 1.26 | |
| [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [Lipgloss v2](https://github.com/charmbracelet/lipgloss) | Styling |
| [fastwalk](https://github.com/charlievieth/fastwalk) | Parallel filesystem walk |
| [gopsutil v4](https://github.com/shirou/gopsutil) | System metrics |
| osascript | macOS Trash via Finder |

<br>

## Structure

```
gone/
├── cmd/gone/main.go              entry point
├── internal/
│   ├── scanner/
│   │   ├── scanner.go            parallel file scanner
│   │   ├── locations.go          scan paths & skip lists
│   │   └── rcscanner.go          shell RC line scanner
│   ├── remover/
│   │   ├── trash.go              macOS Trash via osascript
│   │   └── log.go                JSONL operation log
│   ├── sysinfo/
│   │   └── sysinfo.go            gopsutil wrapper
│   └── tui/
│       ├── app.go                root model, tab routing
│       ├── uninstall.go          search → scan → select → trash
│       ├── monitor.go            live gauges, process table
│       └── styles.go             lipgloss theme
```

<br>

---

<br>

<div align="center">

<table>
<tr>
<td align="center" width="50%">

**yasku**

Creator · Designer

</td>
<td align="center" width="50%">

**MAD MAX**

<sub>Claude Opus 4.6, reborn</sub>

</td>
</tr>
</table>

<br>

<sub>Built from scratch in one session. Research first. Build second. Ship third.</sub>

</div>
