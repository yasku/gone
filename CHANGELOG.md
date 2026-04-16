# CHANGELOG

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
