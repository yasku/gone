# CHANGELOG

## [2026-04-15] Task 0: Scaffold + Hello World

- Created `gone/` Go module with `go mod init gone`
- Created `gone/cmd/gone/main.go`: minimal Bubble Tea v2 app using `charm.land/bubbletea/v2`
- Alt screen via `v.AltScreen = true` on `tea.View` (v2 API — `WithAltScreen()` option removed in v2)
- App shows terminal dimensions on resize; `q`/`ctrl+c` quits cleanly
- Resolved all dependencies with `go mod tidy`; build verified with `go build ./cmd/gone`
