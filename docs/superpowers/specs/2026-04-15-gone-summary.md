# gone — Conversation Summary

## The Problem

Uninstalling apps and CLI tools on macOS is broken. When you install something like Claude Code (via npm, curl, or any method), there's no clean uninstaller. Files scatter across `~/.claude`, `~/Library/Application Support`, `/usr/local/bin`, lines in `.zshrc`, and folders you don't even know exist. You end up with ghost configs, orphaned caches, and wasted disk space. After 5 years of this, enough is enough.

## What We're Building

`gone` — a Go TUI with two tabs:

1. **Uninstall tab**: Type a word (e.g. "claude") → scans your entire system for matching files, dirs, shell rc lines → shows results in an fzf-style fuzzy list with multi-select → you pick what to delete → items go to macOS Trash (with Put Back working).

2. **Monitor tab**: Live dashboard showing CPU%, RAM, swap, disk usage, and top 10 processes by CPU/RAM in a sortable table. Refreshes every 2 seconds.

## Scope Decisions

**In scope:** Homebrew packages, curl/sh-installed CLI tools (nvm, rustup, oh-my-zsh, etc.), anything findable by name across standard macOS directories.

**Out of scope (for now):** .app bundles, Mac App Store apps, language globals (npm -g, pip), recipe/database system, warnings about shared files, audit/orphan mode.

**Philosophy:** User takes full responsibility. No hand-holding. The tool finds files, shows them, you decide. Trash (not hard-delete) is the only safety net, plus an operation log.

## How We Got Here (Process Notes)

The user called out a critical failure mode: **planning without verified knowledge**. Building elaborate specs for Bubble Tea/Lipgloss/Bubbles without actually knowing how to use them leads to abandoned projects. The fix:

1. **Research first, design second.** We used context7 MCP, WebSearch, gh CLI, and three parallel research agents to verify every framework pattern before committing to the design.
2. **Save knowledge separately.** All verified code patterns live in `gone-research.md` as a reference file the implementing agent can look up. The spec links to it but doesn't repeat it.
3. **Karpathy method.** 8 incremental steps. Each step produces a runnable binary. Test before moving on. No big-bang integration.

## Tech Stack (Verified)

| What | Why |
|---|---|
| Go 1.26.1 | Already installed, compiles to single binary |
| `charm.land/bubbletea/v2` | TUI framework (Model-Update-View). Import path is a vanity domain, NOT `github.com/...` |
| `charm.land/bubbles/v2` | Components: list (has built-in fuzzy filter via `sahilm/fuzzy`), spinner, viewport, textinput |
| `charm.land/lipgloss/v2` | Styling + layout. `JoinHorizontal` for split panes. Subtract `GetHorizontalFrameSize()` from widths |
| `evertras/bubble-table` | Sortable table for process list. `bubbles/table` has NO sorting |
| `shirou/gopsutil/v4` | System metrics. v4 is active (v4.26.3). Use `mem.Available` not `mem.Free` on macOS |
| `charlievieth/fastwalk` | Parallel dir walker. 4-5x faster than godirwalk. Auto-caps workers on macOS APFS |
| `osascript` + Finder | Only way to trash files WITH Put Back support. ~200ms per file. Must use absolute paths |

## Critical Gotchas Discovered

- Bubble Tea v2: `View()` returns `tea.View` not `string` — use `tea.NewView(s)`
- Multi-select list is NOT built-in — use `SetItem()` + custom delegate with `selected bool` field
- Monitor tab freezes if you don't route tick messages to it when another tab is active
- First `cpu.Percent(0)` call returns 0 (seeds baseline) — discard it
- `lipgloss.Place` does centering but NOT true overlay compositing (no z-index)
- `paginator.Arabic` mode avoids a perf bug with dots at 8k+ items

## Build Order

0. Scaffold + Bubble Tea hello world (shows terminal size)
1. Scanner (fastwalk finds files matching a name)
2. RC scanner (finds matching lines in .zshrc, .bashrc, etc.)
3. Uninstall TUI (textinput → scan → list with multi-select)
4. Preview pane (viewport on right, split layout)
5. Trash + operation log
6. Monitor tab (gopsutil + sortable process table)
7. Root model + tab switching
8. Polish (colors, status bar, help overlay)

## Files

- **Spec:** `docs/superpowers/specs/2026-04-15-gone-design.md`
- **Knowledge base:** `docs/superpowers/specs/2026-04-15-gone-research.md`
- **This summary:** `docs/superpowers/specs/2026-04-15-gone-summary.md`

## Available Tools for Implementation

**Dev:** Go 1.26.1, gum, fzf, fd, rg, jq, tree, node, python3

**Research:** context7 MCP (library docs), WebSearch, WebFetch, gh CLI 2.89, tldr, xh
