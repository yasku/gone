# gone — User Guide

## Overview

gone is a macOS TUI application with 5 tabs for uninstalling apps, monitoring system resources, tracking network traffic, reviewing operation history, and running security audits.

```
┌────────────────────────────────────────────────────────────────┐
│  ◉ Uninstall   ○ Monitor   ○ Network   ○ Logs   ○ Audit       │
│  G O N E        hunt. select. trash.                          │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│                      Tab Content                               │
│                                                                │
├────────────────────────────────────────────────────────────────┤
│  [tab] switch   [?] help   [ctrl+c] quit                      │
└────────────────────────────────────────────────────────────────┘
```

## Global Keybindings

| Key | Action |
|:----|:-------|
| `Tab` | Cycle through tabs (Uninstall → Monitor → Network → Logs → Audit) |
| `?` | Toggle help overlay |
| `Ctrl+C` | Quit application |

## Tab-Specific Keybindings

### Uninstall Tab

| Key | Action |
|:----|:-------|
| `Enter` | Start scanning for the entered term |
| `Space` | Toggle file selection |
| `Enter` (in list) | Trash selected files (shows confirmation first) |
| `/` | Filter results |
| `Esc` | Back to search / Quit |

**Workflow:**
1. Type a name (e.g., `claude`, `nvm`, `rustup`)
2. Press `Enter` to scan
3. Use `↑/↓` to navigate results
4. Press `Space` to select files
5. Press `Enter` to trash
6. Confirm with `Enter` or cancel with `Esc`

### Monitor Tab

| Key | Action |
|:----|:-------|
| `1` | Sort by CPU |
| `2` | Sort by Memory |
| `3` | Sort by RSS |
| `4` | Sort by PID |
| `↑/↓` | Navigate process list |
| `x` | Kill selected process |
| `/` | Filter processes |
| `r` | Refresh stats |

**Displayed metrics:**
- CPU usage (animated gauge)
- RAM usage (animated gauge)
- Swap usage (animated gauge)
- Disk usage (animated gauge)
- Top 50 processes by CPU

### Network Tab

| Key | Action |
|:----|:-------|
| `/` | Filter interfaces by name |
| `r` | Refresh stats |

**Displayed metrics:**
- Per-interface RX/TX gauges (bytes/second)
- Packet counts
- Auto-refreshes every 2 seconds

### Logs Tab

| Key | Action |
|:----|:-------|
| `/` | Filter log entries |
| `r` | Refresh log view |
| `c` | Clear filter |
| `↑/↓` | Scroll through logs |

**Log file location:** `~/.config/gone/operations.log`

**Entry types:**
- `TRASH` — File moved to trash
- `SCAN` — Scan operation
- `START` — Application started

### Audit Tab

| Key | Action |
|:----|:-------|
| `↑/↓` | Navigate categories |
| `r` | Refresh audit |
| `1-5` | Jump to category |

**Categories:**
1. Startup Items — Apps that run at login
2. Browser Plugins — Browser extensions
3. Network Connections — Active connections
4. Open Ports — Listening ports
5. Scheduled Tasks — Cron jobs, launch agents

**Note:** Requires `osquery` installed. If unavailable, shows install instructions.

## Features

### Uninstall Tab

- **Instant search** — Type a name, scan in seconds
- **Parallel filesystem walk** — Scans 10+ locations simultaneously
- **Shell RC detection** — Finds PATH exports, aliases, and source commands in dotfiles
- **Preview pane** — Inspect files before removing (shown when terminal > 80 chars)
- **Multi-select** — Space to toggle, Enter to trash
- **Safe removal** — Files go to macOS Trash via Finder, not `rm`
- **Per-item size bars** — Visual indicator of disk usage
- **Confirmation modal** — Review selection + total size before removal

### Monitor Tab

- **Animated gauges** — Spring-physics progress bars for CPU, RAM, swap, disk
- **Process table** — Sortable by CPU, memory, RSS, or PID
- **Color-coded severity** — High/medium/low CPU usage indicators
- **Auto-refresh** — 2 second updates
- **Process kill** — `x` key to SIGTERM selected process

### Network Tab

- **Per-interface gauges** — RX (receive) and TX (transmit) animated bars
- **Interface filtering** — `/` to filter by interface name (e.g., `en0`, `lo0`)
- **Auto-refresh** — 2 second updates

### Logs Tab

- **Operation history** — All trash operations logged with timestamps
- **Color-coded entries** — Different colors for trash/scan/start operations
- **Filtering** — Find specific paths or operations
- **Viewport scrolling** — Navigate through long logs

### Audit Tab

- **Security audit** — Query system state via osquery
- **Category navigation** — Arrow keys to explore
- **Status indicators** — ✓ (clean) / ⚠ (warning) per category
- **Graceful degradation** — Works without osquery (shows install instructions)

## Configuration

gone requires no configuration files. Optional tools can be installed for enhanced functionality.

### Optional Tools

| Tool | Purpose | Install |
|:-----|:----|:-------|
| `fd` | Faster file scanning | `brew install fd` |
| `osquery` | Security audit queries | `brew install osquery` |

## FAQ

### Q: Where do trashed files go?
A: Files are sent to macOS Trash via Finder AppleScript. You can restore them via Finder's "Put Back" feature.

### Q: Does gone delete shell RC lines?
A: gone **detects** RC lines (PATH exports, aliases, sources) but does not delete them automatically. You must edit your `.zshrc` manually to remove lines.

### Q: How do I remove an RC line?
A: The uninstall tab shows RC lines with format `~/.zshrc:14` (file:line). Edit the file manually:
```bash
nano ~/.zshrc  # or vim, code, etc.
```

### Q: What locations does gone scan?
A:
- `~/Library/Caches` — App caches
- `~/Library/Application Support` — App data
- `~/Library/Preferences` — Plist files
- `~/Library/Logs` — Log files
- `~/.config` — XDG-style configs
- `~/.local` — User binaries
- `/usr/local` — Homebrew/manual installs
- `/opt` — System packages

### Q: Is gone safe to use?
A: Yes. Files are moved to Trash, not deleted. You can always "Put Back" from Trash.

### Q: Can I run gone on Linux/Windows?
A: No. gone uses macOS-specific APIs (osascript for Trash) and is not cross-platform.

### Q: How do I report a bug?
A: Open an issue at https://github.com/yasku/gone/issues

## Keyboard Quick Reference

```
┌─────────────────────────────────────────────────────────────────┐
│ Global                                                           │
├──────┬──────────────────────────────────────────────────────────┤
│ Tab  │ Switch tabs                                               │
│ ?    │ Help overlay                                              │
│ Ctrl+C │ Quit                                                    │
└──────┴──────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Uninstall Tab                                                    │
├──────┬──────────────────────────────────────────────────────────┤
│ Enter│ Scan / Trash selected                                     │
│ Space│ Toggle selection                                           │
│ Esc  │ Back / Quit                                               │
└──────┴──────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Monitor Tab                                                       │
├──────┬──────────────────────────────────────────────────────────┤
│ 1-4  │ Sort by CPU / Memory / RSS / PID                         │
│ ↑/↓  │ Navigate process list                                     │
│ x    │ Kill process                                              │
│ /    │ Filter processes                                          │
│ r    │ Refresh                                                   │
└──────┴──────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Network Tab                                                       │
├──────┬──────────────────────────────────────────────────────────┤
│ /    │ Filter interfaces                                         │
│ r    │ Refresh stats                                             │
└──────┴──────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Logs Tab                                                          │
├──────┬──────────────────────────────────────────────────────────┤
│ /    │ Filter logs                                               │
│ r    │ Refresh                                                   │
│ c    │ Clear filter                                              │
│ ↑/↓  │ Scroll                                                    │
└──────┴──────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Audit Tab                                                         │
├──────┬──────────────────────────────────────────────────────────┤
│ ↑/↓  │ Navigate categories                                       │
│ r    │ Refresh                                                   │
│ 1-5  │ Jump to category                                          │
└──────┴──────────────────────────────────────────────────────────┘
```
