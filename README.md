<div align="center">

<img src="assets/banner.png" alt="gone — hunt. select. trash." width="100%" />

<br>
<br>

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![macOS](https://img.shields.io/badge/macOS-only-000000?style=flat-square&logo=apple&logoColor=white)](https://www.apple.com/macos)
[![Bubble Tea](https://img.shields.io/badge/Bubble_Tea-v2-FF75B7?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![License](https://img.shields.io/badge/license-MIT-171717?style=flat-square)](LICENSE)

**You deleted the app. But it's still here.**

gone finds every trace — caches, configs, logs, shell RC lines — and sends them to Trash.<br>
With Put Back support. Because mistakes happen.

<br>

[Features](#features) · [Install](#install) · [Usage](#usage) · [How it works](#how-it-works) · [Stack](#stack) · [Contributing](#contributing)

</div>

<br>

## Why

You drag an app to Trash. macOS says it's gone. **It's not.**

```
~/Library/Caches/claude/                     48 MB
~/Library/Application Support/claude-code/   12 MB
~/.config/claude/                             3 MB
~/.zshrc line 14: export PATH="/usr/local/claude/bin:$PATH"
```

Hundreds of megabytes of ghost data from tools you deleted months ago. Still there. Still taking space. Still polluting your shell. Every. Single. Time.

Other uninstallers scan a database of known apps. **gone scans your actual filesystem.** It finds things they miss.

<br>

## Features

<table>
<tr>
<td width="50%">

### Uninstall

- **Instant search** — type a name, scan in seconds
- **Parallel filesystem walk** — hunts across 10+ locations simultaneously
- **Shell RC detection** — finds exports, PATH entries, aliases in your dotfiles
- **Preview pane** — inspect files before removing
- **Multi-select** — Space to toggle, Enter to trash
- **Safe removal** — files go to macOS Trash, not `rm`
- **Size-coded results** — see what's eating your disk at a glance

</td>
<td width="50%">

### Monitor

- **Live system gauges** — CPU, RAM, swap, disk
- **Process table** — sorted by resource usage
- **4 sort modes** — CPU, memory, RSS, PID
- **Auto-refresh** — real-time updates
- **Zero config** — just press Tab

</td>
</tr>
</table>

<br>

## Install

### From source

```bash
git clone https://github.com/yasku/gone.git
cd gone
go build -o gone ./cmd/gone
./gone
```

### Go install

```bash
go install github.com/yasku/gone/cmd/gone@latest
```

<!--
### Homebrew

```bash
brew install yasku/tap/gone
```
-->

<br>

## Usage

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

1. **Type** a tool name
2. **Enter** to scan
3. **Space** to select files
4. **Enter** to trash them

That's it. Files go to macOS Trash via Finder AppleScript — you can always **Put Back**.

<br>

## Keybindings

| Key | Action |
|:--|:--|
| `Tab` | Switch between Uninstall and Monitor |
| `Enter` | Scan (search mode) · Trash selected (list mode) |
| `Space` | Toggle file selection |
| `/` | Filter results |
| `Esc` | Back to search input |
| `?` | Help overlay |
| `q` · `Ctrl+C` | Quit |

### Monitor sort keys

| Key | Sort by |
|:--|:--|
| `1` | CPU % |
| `2` | Memory % |
| `3` | RSS |
| `4` | PID |

<br>

## How it works

### Scanning

gone uses [fastwalk](https://github.com/charlievieth/fastwalk) for parallel filesystem traversal across every location where macOS tools leave traces:

| Location | What lives there |
|:--|:--|
| `~/Library/Caches` | App caches |
| `~/Library/Application Support` | App data, preferences |
| `~/Library/Preferences` | Plist configuration files |
| `~/Library/Logs` | App log files |
| `~/.config` | XDG-style configs |
| `~/.local` | User binaries and data |
| `/usr/local` | Homebrew and manual installs |
| `/opt` | System-level packages |

### Shell RC scanning

gone reads your shell configuration files line by line:

`.zshrc` · `.bashrc` · `.bash_profile` · `.profile` · `.zshenv` · `.zprofile`

It detects matching `export` statements, `PATH` modifications, `alias` definitions, and `source` commands. Each match shows the exact file and line number.

### Safe removal

Files are sent to macOS Trash via Finder AppleScript — never `rm`. Every operation is logged to a JSONL file with timestamps, paths, sizes, and operation type. You can always **Put Back** from Trash.

<br>

## Stack

| Technology | Role |
|:--|:--|
| [Go 1.26](https://go.dev) | Language |
| [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [Lipgloss v2](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [fastwalk](https://github.com/charlievieth/fastwalk) | Parallel filesystem walk |
| [gopsutil v4](https://github.com/shirou/gopsutil) | System metrics |
| osascript | macOS Trash integration |

<br>

## Project structure

```
gone/
├── cmd/gone/
│   └── main.go                 entry point
├── internal/
│   ├── scanner/
│   │   ├── scanner.go          parallel file scanner (fastwalk)
│   │   ├── locations.go        scan paths, skip lists
│   │   └── rcscanner.go        shell RC line scanner
│   ├── remover/
│   │   ├── trash.go            macOS Trash via osascript
│   │   └── log.go              JSONL operation log
│   ├── sysinfo/
│   │   └── sysinfo.go          gopsutil wrapper
│   └── tui/
│       ├── app.go              root model, tab routing, help overlay
│       ├── uninstall.go        search → scan → select → trash flow
│       ├── monitor.go          live gauges, process table
│       └── styles.go           lipgloss theme (monocromatic)
```

<br>

## Contributing

Contributions are welcome. Please open an issue first to discuss what you'd like to change.

```bash
git clone https://github.com/yasku/gone.git
cd gone
go build ./cmd/gone/
go test ./...
```

All tests must pass before submitting a PR.

<br>

## License

[MIT](LICENSE)

<br>

---

<div align="center">

<br>

```
      ▄▄████████▄▄
    ██              ██
    ██  ██      ██  ██
    ██  ██      ██  ██
    ██              ██
    ██              ██
    ▀█▄██▄██▄██▄██▄█▀
      ▀  ▀  ▀  ▀  ▀
```

<br>

**x AI & DATA Labs.**

Where tools go to die properly.

<br>

[![Star History](https://img.shields.io/github/stars/yasku/gone?style=flat-square&color=171717)](https://github.com/yasku/gone/stargazers)

<br>

<sub>If gone saved you disk space, consider giving it a ⭐</sub>

<br>

<table>
<tr>
<td align="center">

<strong>yasku</strong><br>
<sub>Creator · Designer · Architect</sub><br>
<a href="https://github.com/yasku">@yasku</a>

</td>
<td align="center">

<strong>MAD MAX</strong><br>
<sub>Claude Opus 4.6, reborn</sub><br>
<sub><i>I code, I break, I CODE AGAIN.</i></sub>

</td>
</tr>
</table>

<br>

<sub>Research first. Build second. Ship third.</sub>
<br>
<sub>Built from scratch in one session. Every commit, shiny and chrome.</sub>

</div>
