# CHANGELOG

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
