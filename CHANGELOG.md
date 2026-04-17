# CHANGELOG

## [2026-04-17] Code review fixes — `feat/tui-beautification-phase1`

Post-review patches from a branch-wide audit. All behavioural correctness, no feature changes.

- Fixed `internal/tui/monitor.go`: `StyleFunc` guard changed from `row > 0` (which skipped the first data row) to `row >= 0`, and `procs[row-1]` corrected to `procs[row]`, so every process row receives the CPU-severity color matching its own data
- Fixed `internal/tui/monitor.go`: restored cursor row highlight after migrating to `lipgloss/v2/table` — `StyleFunc` now checks `row == m.cursor` first so the selected style wins over CPU-severity and row-parity styling
- Fixed `internal/tui/uninstall.go`: `CursorRow` style now receives `Width(m.Width())` so the highlight background spans the full list column; lipgloss only fills background to text width unless width is explicit
- Fixed `internal/tui/app.go`: added focus-trap for the help overlay — all keys except `?` (toggle) and `ctrl+c` (quit) now return early while the modal is visible, preventing propagation to the active sub-model
- Fixed `internal/remover/trash.go`: replaced AppleScript string interpolation with `osascript` argv (`-e` + positional argument); filenames with embedded double quotes no longer break Trash operations
- Fixed `internal/remover/log.go`: added `$HOME` environment fallback when `os.UserHomeDir()` fails, mirroring the pattern in `internal/scanner/locations.go`; prevents a root-relative `"/.config/gone/operations.log"` when the home dir can't be resolved
- Verified: `go build ./...` succeeds; `go test ./...` all green (5/5 tests)

## [2026-04-17] TUI beautification — phase 1

Plan: `docs/superpowers/plans/2026-04-17-beautification.md`. Monitor tab gets spring-physics animation, process table becomes a proper styled lipgloss table, removals gain a confirmation modal, and the cursor row highlight is promoted to full-width everywhere.

- Added `internal/tui/monitor.go`: four `progress.Model` fields (`cpuBar`, `ramBar`, `swapBar`, `diskBar`) initialized via a `newGaugeBar()` helper using `progress.WithColors(#9B59B6, #00BCD4)` + `WithoutPercentage`; on every `refreshMsg` each gauge's `SetPercent()` is driven from the current `Snapshot`, producing animated transitions instead of the previous static text boxes
- Upgraded `internal/tui/monitor.go`: replaced the manual process loop with `charm.land/lipgloss/v2/table.New()` + `StyleFunc`; columns for PID/CPU%/Mem%/RSS/Name rendered with row-parity and CPU-severity colors
- Added `internal/tui/monitor.go`: CPU% severity coloring — green below 15%, yellow 15–50%, red ≥ 50%
- Added `internal/tui/uninstall.go`: `confirmModal` state with `trashConfirmMsg` / `confirmOKMsg` wiring — before the actual Trash call, a rounded-border modal renders selected count + total freed size and waits for `y`/`enter` to proceed, `esc`/`n` to cancel
- Added `internal/tui/styles.go`: `CursorRow` style with full-width background highlight; `internal/tui/uninstall.go` applies `m.Width()` to the selected delegate's row so the purple highlight extends across the full list column
- Verified: `go build ./...` succeeds; `go test ./...` all green (5/5 tests)

## [2026-04-17] UX/UI redesign

Plan: `docs/superpowers/plans/2026-04-17-gone-ux-ui.md`. Permanent branded header, gradient accents, per-item size progress bars.

- Added `internal/tui/styles.go`: `gradientText()` helper that blends `#9B59B6` → `#00BCD4` across each rune using `lipgloss.Blend1D`; `SearchBarActive` style with gradient border that only renders while the search bar has focus
- Modified `internal/tui/app.go`: permanent header bar with gradient `GONE` wordmark, active-tab indicator, and help hint (`?`); rendered above every tab view
- Modified `internal/tui/uninstall.go`: per-item size progress bar in the file list delegate — fills proportional to the file's size vs. the largest result in view, green→yellow→red gradient by severity; preview pane border uses the same purple→cyan gradient
- Added `github.com/charmbracelet/x` transitive dependency via `lipgloss/v2` gradient APIs
- Verified: `go build ./...` succeeds; `go test ./...` all green (5/5 tests)

## [2026-04-17] Animated splash screen

Plan: `docs/superpowers/plans/2026-04-17-splash-screen.md`. Startup splash with gradient-bordered ASCII banner and spinner; 2.5 s before transitioning to the main app.

- Created `internal/tui/splash.go`: `SplashModel` with `charm.land/bubbles/v2/spinner` (Globe style, color `#9B59B6`); renders 6-line ASCII `GONE` banner in cyan inside a rounded-border box with `BorderForegroundBlend(#9B59B6, #00BCD4)`; emits `splashDoneMsg{}` via `tea.Tick` after 2500 ms
- Modified `internal/tui/app.go`: `AppModel` starts in splash state; on `splashDoneMsg` transitions to the normal root model; spinner message ticks are forwarded to `SplashModel` until it exits
- Adjusted splash duration from 800 ms → 2500 ms for readability
- Verified: `go build ./...` succeeds; `go test ./...` all green (5/5 tests)

## [2026-04-17] Fix preview pane alignment

- Fixed `internal/tui/uninstall.go`: wrapped `m.list.View()` in `lipgloss.NewStyle().Width(listW).Render(...)` so the preview pane anchors correctly to the right regardless of list item length

## [2026-04-16] Open source launch governance

- Created `LICENSE` (MIT, copyright Agustin Yaskuloski 2026)
- Created `CONTRIBUTING.md`: dev setup, ground rules (macOS-only, no curated DBs, Trash-only removal), PR flow with `feat:`/`fix:`/`docs:` commit prefixes, code style, testing guidelines, bug + security reporting
- Created `SECURITY.md`: supported versions, vulnerability reporting policy via agustiny-dev.ar, in-scope and out-of-scope issues, hall of fame
- Created `CODE_OF_CONDUCT.md`: Contributor Covenant v2.1 with enforcement ladder
- Created `.github/ISSUE_TEMPLATE/bug_report.yml`: structured form requiring macOS version, Go version, gone commit SHA, repro steps, log output
- Created `.github/ISSUE_TEMPLATE/feature_request.yml`: alignment-checklist tied to gone's scope; rejects cross-platform requests
- Created `.github/ISSUE_TEMPLATE/config.yml`: disables blank issues, links to Discussions and SECURITY.md
- Created `.github/PULL_REQUEST_TEMPLATE.md`: type-of-change selector, full verification checklist (`go fmt`/`vet`/`build`/`test`), `CHANGELOG.md` update reminder, ground-rules checkboxes
- Modified `README.md`:
  - Added 7 new badges (Built with Claude Code, Powered by MAD MAX, Made by yasku, GitHub stars/issues/last-commit, PRs Welcome)
  - Wrapped Usage TUI mockup in `<div align="center">` for centered rendering on GitHub
  - Moved `assets/banner.png` into the footer placeholder under "Who we are"
- Verified: `go build ./cmd/... ./internal/...` succeeds; `go test ./cmd/... ./internal/...` all green (5/5 tests)

## [v1.0.0] — 2026-04-16

First stable release. `gone` is a macOS uninstaller and system monitor TUI built in Go with Bubble Tea v2.

### Features

- **Uninstall tab** — instant filesystem search across `~/Library/Caches`, `~/Library/Application Support`, `~/Library/Preferences`, `~/Library/Logs`, `~/.config`, `~/.local`, `/usr/local`, `/opt`; parallel walk via fastwalk
- **Shell RC scanning** — detects `export`, `PATH`, `alias`, `source` matches in `.zshrc`, `.bashrc`, `.bash_profile`, `.profile`, `.zshenv`, `.zprofile` with exact file and line number
- **Preview pane** — directory tree, file metadata, modified timestamp
- **Multi-select** with Space, trash with Enter; size-coded results (green/yellow/red by file size)
- **Safe removal** via macOS Trash through Finder AppleScript — Put Back always works; operations logged to `~/.config/gone/operations.log`
- **Monitor tab** — live CPU/RAM/Swap/Disk gauges, process table with 4 sort modes (CPU%, Mem%, RSS, PID), 2s auto-refresh
- **Root model** with Tab switching and help overlay (`?`)

### Documentation

- README with full Usage, Install, Stack, Project structure sections
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`
- Issue and Pull Request templates in `.github/`
- MIT license

### Stack

- Go 1.26
- Bubble Tea v2, Lipgloss v2, Bubbles v2
- fastwalk for parallel filesystem traversal
- gopsutil v4 for system metrics
- osascript for Trash integration

---

## [2026-04-16] Task 8: Polish

- Modified `gone/internal/tui/uninstall.go`: updated `fileDelegate.Render()` to color-code file sizes — green (`SizeSmall`) for <1 MB, yellow (`SizeMedium`) for 1 MB–100 MB, red (`SizeLarge`) for ≥100 MB; added dim-text rendering for file kind column in the list row
- Modified `gone/internal/tui/app.go`: added `showHelp bool` field to `AppModel`; added `?` key handler in `Update()` that toggles `showHelp`; added help overlay in `View()` using `lipgloss.Place` centered on screen — shows all keybindings in a rounded-bordered box; overlay dismisses on second `?` press; `AltScreen = true` preserved on overlay view
- Verified: `go build ./...` succeeds; `go test ./...` all pass (5/5 tests)

## [2026-04-16] Task 7: Root Model + Tab Switching

- Created `gone/internal/tui/app.go`: `AppModel` struct with `activeTab` enum (tabUninstall / tabMonitor); `NewApp()` constructor; `Init()` batches both sub-model init commands; `Update()` handles `ctrl+c` quit, `tab` key cycling, `WindowSizeMsg` resizing both sub-models; always routes `refreshMsg` to monitor regardless of active tab (prevents freeze); routes other messages to active tab only; `View()` renders tab bar with active/inactive styles via lipgloss + bottom border, then delegates content to active sub-model; sets `AltScreen = true` on returned `tea.View`
- Modified `gone/cmd/gone/main.go`: simplified to a single `main()` that calls `tea.NewProgram(tui.NewApp())` — root model moved entirely into `tui.AppModel`
- Verified: `go build ./...` succeeds; `go test ./...` all pass (scanner, remover, sysinfo packages)

## [2026-04-16] Task 6: System Monitor

- Created `gone/internal/sysinfo/sysinfo.go`: `Snapshot` struct (CPU%, MemTotal/Used/Avail, Swap, Disk, top-N processes); `TakeSnapshot(topN int)` collects data via gopsutil v4 (cpu, mem, disk, process packages); `HumanBytes()` formatter; processes sorted by CPU% descending
- Created `gone/internal/sysinfo/sysinfo_test.go`: `TestTakeSnapshotReturnsData` verifies MemTotal, DiskTotal, and Procs are non-zero
- Created `gone/internal/tui/monitor.go`: `MonitorModel` with `refreshMsg` + `doRefresh()` tea.Tick every 2s; `SetSize()` for layout; `View()` renders 4 system gauges (CPU/RAM/Swap/Disk) via lipgloss bordered boxes + process table with header, separator, highlighted cursor row; keyboard navigation (↑/↓, sort keys 1-4); `sortedProcs()` re-sorts by Mem/RSS/PID or keeps CPU order
- Adapted from plan: replaced `evertras/bubble-table` (which targets bubbletea v1) with a manual lipgloss table to avoid v1/v2 type incompatibility; same visual layout and interactive sorting preserved
- Added `github.com/shirou/gopsutil/v4 v4.26.3` to go.mod
- Verified: `go build ./...` succeeds; `go test ./...` all pass (5/5 tests)

## [2026-04-16] Task 5: Trash + Operation Log

- Created `gone/internal/remover/trash.go`: `MoveToTrash()` sends file to macOS Trash via `osascript` + Finder AppleScript; escapes quotes in paths; returns wrapped error with stderr on failure
- Created `gone/internal/remover/log.go`: `LogEntry` struct (ts, op, path, size, kind, term); `AppendLog()` writes a JSON-lines entry to `~/.config/gone/operations.log`; creates the directory if missing
- Created `gone/internal/remover/log_test.go`: `TestAppendLog` — overrides `HOME`, calls `AppendLog`, reads log file, unmarshals JSON line, verifies `path` and `op` fields
- Modified `gone/internal/tui/uninstall.go`:
  - Added `"gone/internal/remover"` import
  - Added `trashDoneMsg` struct (count, freed, errors)
  - Added `trashSelected()` command: iterates selected items (skips rc-lines), calls `MoveToTrash`, logs each success, collects errors; returns `trashDoneMsg`
  - Replaced placeholder Enter handler with real trash flow: calls `trashSelected(m.SelectedItems(), m.term)`, sets `m.scanning = true` while in progress
  - Added `trashDoneMsg` case in `Update()`: clears scanning flag, updates status with count/freed/errors, removes trashed items from list
- Verified: `go mod tidy && go build ./cmd/gone` succeeds; `go test ./...` all pass (scanner + remover packages)

## [2026-04-16] Task 4: Preview Pane

- Modified `gone/internal/tui/uninstall.go`:
  - Added `charm.land/bubbles/v2/viewport` and `charm.land/lipgloss/v2` imports
  - Added `viewport viewport.Model` and `showPreview bool` fields to `UninstallModel`
  - Initialized viewport with `viewport.New()` in `NewUninstallModel()`
  - Updated `SetSize()` to split width 50/50: list gets `w/2-2`, viewport gets `w/2-4`; `showPreview` hides pane when terminal is ≤80 cols wide (uses `SetWidth`/`SetHeight` pointer methods on v2 viewport)
  - Added `previewContent()` function: shows path, type, size, modified date; for dirs lists up to 20 entries with overflow count; for rc-lines shows file and line number
  - Updated `View()` to use `lipgloss.JoinHorizontal` for side-by-side split when `showPreview` is true, falls back to full-width list otherwise
  - Updated `Update()` to set initial viewport content on `scanResultMsg` and refresh preview after each list navigation keystroke
- Fixed v2 API differences: `viewport.New()` takes option funcs (not int args); `Width`/`Height` are not assignable fields — use `SetWidth()`/`SetHeight()` methods
- Verified: `go build ./...` succeeds; `go test ./...` all pass

## [2026-04-16] Task 3: Uninstall TUI — Search, Scan, List, Multi-Select

- Created `gone/internal/tui/styles.go`: `Styles` struct with lipgloss styles (app, tabs, search bar, status bar, preview, selected, cursor, dim text, size colors); `DefaultStyles()` factory; `HumanSize()` formatter
- Created `gone/internal/tui/uninstall.go`: `UninstallModel` with textinput (search bar), spinner (async scan indicator), list (custom single-line delegate with `[ ]`/`[x]` checkbox), and multi-select logic; `fileItem` type implementing `list.Item`; `fileDelegate` custom renderer; `runFullScan()` command that combines file scanner + RC scanner results; focus state machine (search → list → search); space toggles selection, esc returns to search, enter triggers scan or placeholder trash action; status bar shows selected count + total size
- Modified `gone/cmd/gone/main.go`: replaced CLI harness with `rootModel` Bubble Tea app; routes `WindowSizeMsg` to `UninstallModel.SetSize()`; renders via `v.AltScreen = true` on returned `tea.View`
- Added `charm.land/bubbles/v2 v2.1.0` and `charm.land/lipgloss/v2 v2.0.3` to go.mod
- Fixed v2 API differences: `Width` is a method in v2 textinput — used `SetWidth()` instead; `WithAltScreen()` removed in v2 — set `v.AltScreen = true` on View
- Verified: `go build ./...` succeeds; `go test ./...` all pass

## [2026-04-15] Task 2: Shell RC Scanner

- Created `gone/internal/scanner/rcscanner.go`: `SearchRC()` scans `~/.zshrc`, `~/.zshenv`, `~/.zprofile`, `~/.bashrc`, `~/.bash_profile`, `~/.profile` for lines matching the search term (case-insensitive); returns `[]RCMatch` with file path, line number, and line content
- Created `gone/internal/scanner/rcscanner_test.go`: `TestSearchRCFindsMatchingLines` — writes a temp `.zshrc`, overrides `RCFiles` and `HOME`, verifies 2 matches for "claude" at lines 2 and 4
- Modified `gone/cmd/gone/main.go`: added RC scan after file scan results; summary line now shows `N files + M rc lines in Xs`
- Verified: `go run ./cmd/gone claude` returns 154 files + 8 rc lines in ~1.1s

## [2026-04-15] Task 1: File Scanner

- Created `gone/internal/scanner/locations.go`: scan root paths and skip-dirs map
- Created `gone/internal/scanner/scanner.go`: `Search()` using fastwalk for parallel file discovery; `DirSize()` helper
- Created `gone/internal/scanner/scanner_test.go`: two tests — finds matching files/dirs, skips `node_modules`
- Modified `gone/cmd/gone/main.go`: CLI test harness; `go run ./cmd/gone <term>` prints all matches with kind/size/date
- Added `github.com/charlievieth/fastwalk v1.0.14` to go.mod
- Fixed test: renamed `main.bin` → `foo-main.bin` so both a dir and a file match "foo"
- Verified: 154 matches for "claude" in ~1.1s (<3s target)

## [2026-04-15] Task 0: Scaffold + Hello World

- Created `gone/` Go module with `go mod init gone`
- Created `gone/cmd/gone/main.go`: minimal Bubble Tea v2 app using `charm.land/bubbletea/v2`
- Alt screen via `v.AltScreen = true` on `tea.View` (v2 API — `WithAltScreen()` option removed in v2)
- App shows terminal dimensions on resize; `q`/`ctrl+c` quits cleanly
- Resolved all dependencies with `go mod tidy`; build verified with `go build ./cmd/gone`
