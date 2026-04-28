# charm.land v2 Audit Report

**Date:** 2026-04-17  
**Auditor:** charmland-expert (mad-max)  
**Scope:** `internal/tui/` (5 files)  
**Baseline build:** ✅ clean (`go build ./cmd/gone/`)

---

## Scope Reviewed

| File | Lines | Status |
|------|-------|--------|
| `internal/tui/styles.go` | 78 | ✅ Clean |
| `internal/tui/app.go` | 219 | ⚠️ Focus trap gap |
| `internal/tui/splash.go` | 92 | ✅ Clean |
| `internal/tui/uninstall.go` | 505 | ⚠️ CursorRow width missing |
| `internal/tui/monitor.go` | 267 | 🔴 StyleFunc off-by-one |

---

## Critical API Misuses (Must Fix)

### 1. `monitor.go:216-217` — StyleFunc row index off-by-one

**Source:** `charm.land/lipgloss/v2/table`, `const HeaderRow int = -1`  
**MemPalace:** wing=lipgloss, room=table → `HeaderRow == -1`; data rows start at **0**.

```go
// BUGGY — current code
if col == 2 && row > 0 && row-1 < len(procs) {
    pct := procs[row-1].CPU  // procs[-1] guarded, but row 0 always skipped
```

With `HeaderRow == -1`, data rows are indexed 0, 1, 2, …  
- `row > 0` → **first process (row 0) never gets CPU color**  
- `procs[row-1]` → **all other rows show the previous process's CPU value**

```go
// FIXED
if col == 2 && row >= 0 && row < len(procs) {
    pct := procs[row].CPU
```

**Fix commit:** `ece6bf5`

---

## v1 → v2 Regressions

None found. All import paths are `charm.land/*` throughout:
- `charm.land/bubbletea/v2` ✅  
- `charm.land/lipgloss/v2` ✅  
- `charm.land/lipgloss/v2/table` ✅  
- `charm.land/bubbles/v2/spinner` ✅  
- `charm.land/bubbles/v2/progress` ✅  
- `charm.land/bubbles/v2/textinput` ✅  
- `charm.land/bubbles/v2/viewport` ✅  
- `charm.land/bubbles/v2/list` ✅  

Model signatures all conform to v2 shapes:
- `Init() tea.Cmd` ✅  
- `Update(msg tea.Msg) (T, tea.Cmd)` ✅  
- `AppModel.View() tea.View` ✅ (v2 returns `tea.View`, not `string`)  
- Sub-models `View() string` ✅  
- `tea.KeyPressMsg` (v2), not `tea.KeyMsg` (v1) ✅  

---

## State / Layout Bugs

### 2. `uninstall.go:90` — CursorRow background clipped (checklist item 14)

`fileDelegate.Render()` renders the cursor row without a fixed width:

```go
// BUGGY
fmt.Fprint(w, d.styles.CursorRow.Render(line))
```

`CursorRow` has `Background(#1A1A2E)` but no `.Width()`. The background only covers  
the text characters, not the full list width — visible as an abrupt color cutoff.

Fix: use `m list.Model` (already available as parameter) to get the width:

```go
// FIXED
fmt.Fprint(w, d.styles.CursorRow.Width(m.Width()).Render(line))
```

**Fix commit:** `443d7fc`

### 3. `app.go:64-70` — Help modal focus trap incomplete (checklist item 13)

When `m.showHelp == true`, the `Tab` key still switches the underlying active tab  
(early-returns on line 71 skip the active-tab routing, but the switch itself runs).  
All other keys (letters, arrows, etc.) fall through to the active sub-model.

```go
// CURRENT — Tab switches tabs even when help overlay is visible
if msg.String() == "tab" {
    if m.active == tabUninstall {
        m.active = tabMonitor
    } else {
        m.active = tabUninstall
    }
    return m, nil
}
// ... then active-tab routing runs for all other keys
```

Fix: add a showHelp guard before any other key handling:

```go
// FIXED — swallow all keys during help modal except ? (toggle) and ctrl+c (quit)
case tea.KeyPressMsg:
    if m.showHelp {
        if msg.String() == "?" || msg.String() == "ctrl+c" {
            // fall through to handle below
        } else {
            return m, nil  // swallow
        }
    }
    // ... existing handlers
```

**Fix commit:** `3a858f3`

---

## Minor Polish Opportunities

- `splash.go:53` — Type assertion after type switch (`msg := msg.(tea.WindowSizeMsg)`) is  
  redundant; use typed switch variable directly. Not a charm.land issue — Go style only.
- `uninstall.go:75-80` — `progress.New()` is called on every `Render()` invocation  
  (inside `fileDelegate.Render`). For lists with many items this creates N progress models  
  per frame. Consider caching in the delegate struct (or at scan time). Not a correctness  
  issue but a performance note.

---

## Out of Scope Observations (for go-expert)

- `uninstall.go:277` — `esc` in `focusSearch` calls `tea.Quit` unconditionally.  
  Pressing Escape before any search exits the whole app; may be intentional UX.
- `viewport.go SetWidth/SetHeight` are pointer receivers; `UninstallModel.SetSize`  
  is a value receiver returning `m`. Correct because `return m` propagates the mutation.  
  No bug, but worth knowing.

---

## Fix Summary

| # | File | Issue | Severity | Commit |
|---|------|-------|----------|--------|
| 1 | `monitor.go:216-217` | StyleFunc off-by-one (`row > 0` → `row >= 0`, `procs[row-1]` → `procs[row]`) | 🔴 Critical | `ece6bf5` |
| 2 | `uninstall.go:90` | CursorRow missing `.Width(m.Width())` | 🟡 Layout | `443d7fc` |
| 3 | `app.go:64-70` | Help overlay focus trap swallows wrong keys | 🟡 State | `3a858f3` |
