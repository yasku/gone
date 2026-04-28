# Code Review — feat/tui-beautification-phase1 vs develop

**Date:** 2026-04-17
**Reviewer:** mad-max (code-reviewer role)
**Base:** `develop` | **Head:** `feat/tui-beautification-phase1`
**Build:** `go build ./internal/... ./cmd/...` ✅ | **Tests:** `go test ./internal/...` ✅

---

## Commits reviewed

| SHA | Message |
|-----|---------|
| `3a858f3` | fix(tui): add focus trap for help overlay in AppModel |
| `443d7fc` | fix(tui): extend CursorRow background to full list width |
| `ece6bf5` | fix(tui): correct StyleFunc row index in process table CPU coloring |
| `c006e9e` | fix(remover): use osascript argv to avoid invalid AppleScript escape |
| `5eaff30` | fix(remover): fallback to $HOME when os.UserHomeDir fails in logPath |
| `357e3f5` | feat(tui): upgrade process table to lipgloss/v2/table with styled rows |
| `1b97226` | feat(tui): add full-width background highlight for cursor row |
| `e4d2846` | feat(tui): add confirmation modal before trashing items |
| `b4caedc` | feat(tui): color-code CPU% by severity in process table |
| `b4caedc` | feat(tui): animate monitor gauges with spring-physics progress bars |

---

## Critical issues (must fix)

### C-1 — `buildTable`: cursor row highlight dropped after table migration

**File:** `internal/tui/monitor.go`, `buildTable()` / `StyleFunc`
**Severity:** Critical UX regression

After migrating from a manual `for i, p := range procs` loop to `lipgloss/v2/table`,
the original cursor highlight (`m.styles.Cursor.Render(line)` when `i == m.cursor`)
was never ported to the new `StyleFunc`. The ↑/↓ and j/k keys still mutate
`m.cursor` correctly, but **no cell in the table ever reflects the selection** —
the user has no visual feedback about which process row is selected.

The `StyleFunc` closure captures both `m MonitorModel` (value) and
`procs []sysinfo.ProcInfo`, so `m.cursor` is fully accessible. The fix is a
3-line guard added before the CPU-coloring and row-parity checks:

```go
if row == m.cursor {
    return m.styles.CursorRow
}
```

`m.styles.CursorRow` was added in `styles.go` (`1b97226`) and is already used
correctly in `uninstall.go` (`443d7fc`). It just needs to be wired up here.

Note: cursor row wins over CPU severity coloring — same behaviour as the original
loop where `Cursor.Render` overrode the plain line string entirely.

**Fix commit:** `fix(tui): restore cursor row highlight in process table (review)`
→ added after report (see §Updates below)

---

## Minor issues (nice to fix)

### M-1 — `logPath`: double-failure produces a relative path

**File:** `internal/remover/log.go`

If both `os.UserHomeDir()` **and** `os.Getenv("HOME")` return empty string,
`filepath.Join("", ".config", "gone", "operations.log")` produces the relative
path `.config/gone/operations.log`, silently writing to whatever the CWD is.
Practically impossible on macOS, but a defensive `if home == ""` return/log
would make the failure loud. Out of scope for this branch; noted for a follow-up.

### M-2 — `confirmView`: `m.height - 4` without zero-guard

**File:** `internal/tui/uninstall.go`, `confirmView()`

`lipgloss.Place(m.width, m.height-4, ...)` — if `m.height` is 0 (e.g., resize
hasn't been received yet), the second argument wraps to a large uint or is
treated as 0. In practice `confirmPending` requires a prior key event, so
`m.height` will always be set first. Low risk; worth a `max(0, m.height-4)`
guard for robustness.

---

## Already-good patterns

| Pattern | Location | Notes |
|---------|----------|-------|
| `osascript argv` injection fix | `remover/trash.go` | Correct — AppleScript has no `\"` escape; argv avoids string interpolation entirely |
| `os.Getenv("HOME")` fallback | `remover/log.go` | Strict improvement over silent `_` discard |
| `progress.FrameMsg` fan-out to all 4 bars | `monitor.go:102-113` | Each bar uses its private `id` to self-select; routing all bars is correct |
| `progress.WithColors` gradient | `monitor.go:49` | Valid bubbles v2 API; confirmed against installed module |
| `BorderForegroundBlend` gradient border | `monitor.go:188`, `uninstall.go:487` | Valid lipgloss v2 API; confirmed in MemPalace |
| `table.HeaderRow` guard before `row >= 0` | `monitor.go:213` | Correct sentinel usage per lipgloss v2/table examples |
| `row < len(procs)` bounds check in StyleFunc | `monitor.go:216` | Prevents panic on snapshot race at render time |
| `confirmPending` early-return key swallow | `uninstall.go:241-253` | Clean modal focus — swallows all non-enter/esc keys |
| Help overlay focus trap | `app.go:63-65` | `?` toggles first (handled), then trap swallows rest |
| `CursorRow.Width(m.Width())` full-width | `uninstall.go:88` | Correct use of `CursorRow` for file list delegate |
| `SetSize` pointer-receiver `SetWidth` | `monitor.go:145-148` | Auto-takes `&m.cpuBar` on addressable field; mutations preserved via `return m` |
| Sorted copy in `sortedProcs` | `monitor.go:247-248` | `make+copy` prevents mutation of snapshot slice ✓ |

---

## Pre-existing issues (out of scope)

- **`backups/2026-04-16/` contains Go source files** with conflicting package names
  (`tui` and `sysinfo`). `go build ./...` fails with "found packages … in …/backups/".
  Workaround: use `go build ./internal/... ./cmd/...`. Should be cleaned up or
  added to a `.goignore` / moved outside the module root.

---

## Updates

| Issue | Fix commit |
|-------|-----------|
| C-1 cursor row highlight | `3c94832` fix(tui): restore cursor row highlight in process table (review) |
