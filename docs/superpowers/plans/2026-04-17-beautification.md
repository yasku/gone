# gone — TUI Beautification Plan

**Goal:** Elevate `gone`'s visual quality across both tabs — animating the Monitor gauges with spring-physics progress bars, upgrading the process table to a styled lipgloss table, adding a trash confirmation modal, and applying gradient borders and severity colors throughout.

**Branch:** `develop`

**Palette:** `#9B59B6` (purple) → `#00BCD4` (cyan) — consistent throughout.

**Verify commands:** `go build ./cmd/gone/` and `go test ./...`

**Files involved:**
- `internal/tui/monitor.go`
- `internal/tui/uninstall.go`
- `internal/tui/app.go`
- `internal/tui/styles.go`

---

## Phase 1: Monitor Tab Visual Overhaul

> The Monitor tab currently shows static text boxes for CPU/RAM/Swap/Disk.
> This phase transforms them into animated, spring-physics progress bars with
> gradient fill, matching the uninstall tab's aesthetic. Biggest visual delta.

---

### Task 0: Animated Progress Bars for Monitor Gauges

**Files to modify:**
- `internal/tui/monitor.go`

The `MonitorModel` currently stores `snapshot sysinfo.Snapshot` and calls a `gauge()` helper that renders a static text box. We replace each gauge with a live `progress.Model` that animates on every `refreshMsg`.

#### Step 1 — Add four `progress.Model` fields to `MonitorModel`

```go
import "charm.land/bubbles/v2/progress"

type MonitorModel struct {
    snapshot   sysinfo.Snapshot
    styles     Styles
    width      int
    height     int
    ready      bool
    cursor     int
    sortBy     sortCol
    cpuBar     progress.Model
    ramBar     progress.Model
    swapBar    progress.Model
    diskBar    progress.Model
}
```

#### Step 2 — Initialize bars in `NewMonitorModel()`

```go
func newGaugeBar() progress.Model {
    return progress.New(
        progress.WithColors(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")),
        progress.WithoutPercentage(),
        progress.WithWidth(20),
    )
}

func NewMonitorModel() MonitorModel {
    return MonitorModel{
        styles:   DefaultStyles(),
        sortBy:   sortCPU,
        cpuBar:   newGaugeBar(),
        ramBar:   newGaugeBar(),
        swapBar:  newGaugeBar(),
        diskBar:  newGaugeBar(),
    }
}
```

#### Step 3 — Update `SetSize` to resize bars

```go
func (m MonitorModel) SetSize(w, h int) MonitorModel {
    m.width = w
    m.height = h
    barW := w/4 - 8
    if barW < 12 {
        barW = 12
    }
    m.cpuBar  = m.cpuBar.WithWidth(barW)
    m.ramBar  = m.ramBar.WithWidth(barW)
    m.swapBar = m.swapBar.WithWidth(barW)
    m.diskBar = m.diskBar.WithWidth(barW)
    return m
}
```

#### Step 4 — Trigger `SetPercent` on `refreshMsg` in `Update()`

Replace the `refreshMsg` handler:

```go
case refreshMsg:
    m.snapshot = sysinfo.TakeSnapshot(15)
    m.ready = true
    if m.cursor >= len(m.snapshot.Procs) && len(m.snapshot.Procs) > 0 {
        m.cursor = len(m.snapshot.Procs) - 1
    }
    s := m.snapshot
    // CPU: 0–100%
    cpuCmd  := m.cpuBar.SetPercent(s.CPUPercent / 100.0)
    // RAM: used/total
    var ramPct float64
    if s.MemTotal > 0 {
        ramPct = float64(s.MemUsed) / float64(s.MemTotal)
    }
    ramCmd := m.ramBar.SetPercent(ramPct)
    // Swap: used/total
    var swapPct float64
    if s.SwapTotal > 0 {
        swapPct = float64(s.SwapUsed) / float64(s.SwapTotal)
    }
    swapCmd := m.swapBar.SetPercent(swapPct)
    // Disk: used/total
    var diskPct float64
    if s.DiskTotal > 0 {
        diskPct = float64(s.DiskUsed) / float64(s.DiskTotal)
    }
    diskCmd := m.diskBar.SetPercent(diskPct)
    return m, tea.Batch(doRefresh(), cpuCmd, ramCmd, swapCmd, diskCmd)
```

Also forward `progress.FrameMsg` to each bar so the spring animation ticks:

```go
case progress.FrameMsg:
    var cmds []tea.Cmd
    var cmd tea.Cmd
    m.cpuBar,  cmd = m.cpuBar.Update(msg)
    cmds = append(cmds, cmd)
    m.ramBar,  cmd = m.ramBar.Update(msg)
    cmds = append(cmds, cmd)
    m.swapBar, cmd = m.swapBar.Update(msg)
    cmds = append(cmds, cmd)
    m.diskBar, cmd = m.diskBar.Update(msg)
    cmds = append(cmds, cmd)
    return m, tea.Batch(cmds...)
```

#### Step 5 — Replace `gauge()` helper with new `gaugeView()`

Remove the old `gauge()` method. Add:

```go
func (m MonitorModel) gaugeView(label, value string, bar progress.Model) string {
    title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00BCD4")).Render(label)
    val   := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(value)
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
        Padding(0, 2).
        Width(m.width/4 - 4).
        Render(title + "\n" + bar.View() + "\n" + val)
}
```

#### Step 6 — Update `View()` gauges section

```go
s := m.snapshot
gauges := lipgloss.JoinHorizontal(lipgloss.Top,
    m.gaugeView("CPU",  fmt.Sprintf("%.1f%%", s.CPUPercent), m.cpuBar),
    m.gaugeView("RAM",  fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.MemUsed), sysinfo.HumanBytes(s.MemTotal)), m.ramBar),
    m.gaugeView("Swap", fmt.Sprintf("%s / %s", sysinfo.HumanBytes(s.SwapUsed), sysinfo.HumanBytes(s.SwapTotal)), m.swapBar),
    m.gaugeView("Disk", fmt.Sprintf("%s free / %s", sysinfo.HumanBytes(s.DiskFree), sysinfo.HumanBytes(s.DiskTotal)), m.diskBar),
)
```

#### Step 7 — Verify and commit

```bash
go build ./cmd/gone/
go test ./...
```

Expected: Monitor tab shows 4 gauge boxes each with a gradient progress bar that animates smoothly on each 2s refresh tick. Spring physics gives a fluid fill animation.

Commit message:
```
feat(tui): animate monitor gauges with spring-physics progress bars
```

---

### Task 1: CPU% Severity Coloring in Process Table

**Files to modify:**
- `internal/tui/monitor.go`

Color each process row's CPU% value based on load severity. No structural change — just a rendering tweak in `View()`.

#### Step 1 — Add `colorCPU()` helper

```go
func colorCPU(pct float64) string {
    s := fmt.Sprintf("%8.1f", pct)
    switch {
    case pct >= 70.0:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render(s) // red
    case pct >= 30.0:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFDD57")).Render(s) // yellow
    default:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("#69FF94")).Render(s) // green
    }
}
```

#### Step 2 — Apply in the process table loop in `View()`

Replace the `CPU` column value in the row render:

```go
// Before:
line := fmt.Sprintf("  %-8d %-25s %8.1f %8.1f %12s",
    p.PID, truncateName(p.Name, 25), p.CPU, p.Mem, sysinfo.HumanBytes(p.RSS))

// After:
line := fmt.Sprintf("  %-8d %-25s %s %8.1f %12s",
    p.PID, truncateName(p.Name, 25), colorCPU(p.CPU), p.Mem, sysinfo.HumanBytes(p.RSS))
```

Also update the header to match the column widths:

```go
header := fmt.Sprintf("  %-8s %-25s %8s %8s %12s", "PID", "Name", "CPU%", "MEM%", "RSS")
```

#### Step 3 — Verify and commit

```bash
go build ./cmd/gone/
go test ./...
```

Expected: CPU% column color-codes: green for idle (<30%), yellow for moderate (30–70%), red for hot (>70%).

Commit message:
```
feat(tui): color-code CPU% by severity in process table
```

---

## Phase 2: Uninstall Tab Safety & Polish

> Adding a confirmation modal before trashing prevents accidental data loss
> and elevates the UX with a centered gradient overlay.

---

### Task 2: Trash Confirmation Modal

**Files to modify:**
- `internal/tui/uninstall.go`

Before trashing selected items on `Enter`, show a centered overlay modal: `"Trash N items (X GB)? [Enter] Confirm  [Esc] Cancel"`. If the user presses `Enter` again, proceed. `Esc` dismisses.

#### Step 1 — Add `confirmPending bool` state to `UninstallModel`

```go
type UninstallModel struct {
    // ... existing fields ...
    confirmPending bool
}
```

#### Step 2 — Replace the `enter` handler in `focusList`

```go
case "enter":
    sel := m.SelectedItems()
    if len(sel) == 0 {
        return m, nil
    }
    m.confirmPending = true
    return m, nil
```

#### Step 3 — Add confirm/cancel handling at the top of `Update()`

Add this before the main `switch msg := msg.(type)` block:

```go
if m.confirmPending {
    if key, ok := msg.(tea.KeyPressMsg); ok {
        switch key.String() {
        case "enter":
            m.confirmPending = false
            sel := m.SelectedItems()
            m.status = fmt.Sprintf("Trashing %d items…", len(sel))
            m.scanning = true
            return m, trashSelected(sel, m.term)
        case "esc":
            m.confirmPending = false
            return m, nil
        }
    }
    return m, nil // swallow all other input while confirm is up
}
```

#### Step 4 — Add `confirmView()` helper

```go
func (m UninstallModel) confirmView() string {
    sel := m.SelectedItems()
    var total int64
    for _, s := range sel {
        total += s.size
    }
    msg := fmt.Sprintf(
        "  Trash %d item(s) (%s)?\n\n  [Enter] Confirm    [Esc] Cancel",
        len(sel), HumanSize(total),
    )
    box := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
        Padding(1, 3).
        Width(44).
        Render(msg)
    return lipgloss.Place(m.width, m.height-4, lipgloss.Center, lipgloss.Center, box)
}
```

#### Step 5 — Render modal in `View()` when `confirmPending`

At the very start of `View()`, before the normal rendering:

```go
func (m UninstallModel) View() string {
    if m.confirmPending {
        return m.confirmView()
    }
    // ... rest of View unchanged ...
}
```

#### Step 6 — Verify and commit

```bash
go build ./cmd/gone/
go test ./...
```

Expected: select items → press Enter → modal appears centered with gradient border showing count + size → Enter confirms and trashes → Esc dismisses without action.

Commit message:
```
feat(tui): add confirmation modal before trashing items
```

---

### Task 3: Full-Width Cursor Highlight in Uninstall List

**Files to modify:**
- `internal/tui/uninstall.go`
- `internal/tui/styles.go`

The cursor row currently shows only a `>` prefix in white. A full-width background highlight makes the selection immediately obvious.

#### Step 1 — Add `CursorRow` style to `Styles`

In `styles.go`, add the field:

```go
type Styles struct {
    // ... existing fields ...
    CursorRow lipgloss.Style
}
```

In `DefaultStyles()`, add:

```go
CursorRow: lipgloss.NewStyle().
    Background(lipgloss.Color("#1A1A2E")).
    Foreground(lipgloss.Color("#00BCD4")).
    Bold(true),
```

#### Step 2 — Apply `CursorRow` in `fileDelegate.Render()`

Replace the final render block:

```go
// Before:
if index == m.Index() {
    fmt.Fprint(w, d.styles.Cursor.Render(line))
} else if f.selected {
    fmt.Fprint(w, d.styles.Selected.Render(line))
} else {
    fmt.Fprint(w, line)
}

// After:
if index == m.Index() {
    fmt.Fprint(w, d.styles.CursorRow.Render(line))
} else if f.selected {
    fmt.Fprint(w, d.styles.Selected.Render(line))
} else {
    fmt.Fprint(w, line)
}
```

#### Step 3 — Verify and commit

```bash
go build ./cmd/gone/
go test ./...
```

Expected: cursor row has a dark navy background with cyan text, clearly distinct from unselected and selected rows.

Commit message:
```
feat(tui): add full-width background highlight for cursor row
```

---

## Phase 3: Process Table Upgrade

> Replaces the manual `fmt.Sprintf` process table with `lipgloss/v2/table`
> for styled headers, alternating row colors, and a thick gradient border.
> Highest code complexity — do Phase 1 & 2 first.

---

### Task 4: lipgloss/v2/table for Process List

**Files to modify:**
- `internal/tui/monitor.go`

#### Step 1 — Add import

```go
import "charm.land/lipgloss/v2/table"
```

Verify `charm.land/lipgloss/v2/table` is available:

```bash
go get charm.land/lipgloss/v2
```

The `table` subpackage ships with lipgloss v2 — no extra dependency needed.

#### Step 2 — Add `buildTable()` helper

```go
func (m MonitorModel) buildTable(procs []sysinfo.ProcInfo) string {
    headerStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("#9B59B6")).
        Align(lipgloss.Center)

    cellStyle := lipgloss.NewStyle().Padding(0, 1)

    oddRow  := cellStyle.Foreground(lipgloss.Color("252"))
    evenRow := cellStyle.Foreground(lipgloss.Color("245"))

    borderStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("240"))

    t := table.New().
        Border(lipgloss.NormalBorder()).
        BorderStyle(borderStyle).
        Headers("PID", "Name", "CPU%", "MEM%", "RSS").
        StyleFunc(func(row, col int) lipgloss.Style {
            if row == table.HeaderRow {
                return headerStyle
            }
            if row%2 == 0 {
                return evenRow
            }
            return oddRow
        })

    for _, p := range procs {
        t.Row(
            fmt.Sprintf("%d", p.PID),
            truncateName(p.Name, 25),
            fmt.Sprintf("%.1f", p.CPU),
            fmt.Sprintf("%.1f", p.Mem),
            sysinfo.HumanBytes(p.RSS),
        )
    }

    return t.Render()
}
```

#### Step 3 — Apply severity color on CPU column

After `buildTable`, wrap the CPU column with `colorCPU()`. Since `lipgloss/table` uses `StyleFunc` per cell (row/col), add column-aware coloring:

```go
StyleFunc(func(row, col int) lipgloss.Style {
    if row == table.HeaderRow {
        return headerStyle
    }
    if col == 2 && row > 0 && row-1 < len(procs) { // CPU% column
        pct := procs[row-1].CPU
        switch {
        case pct >= 70.0:
            return cellStyle.Foreground(lipgloss.Color("#FF6B6B"))
        case pct >= 30.0:
            return cellStyle.Foreground(lipgloss.Color("#FFDD57"))
        default:
            return cellStyle.Foreground(lipgloss.Color("#69FF94"))
        }
    }
    if row%2 == 0 {
        return evenRow
    }
    return oddRow
})
```

> Note: `StyleFunc` replaces the previous one — combine both into a single func.

#### Step 4 — Replace old process table rendering in `View()`

Remove the manual `for i, p := range procs` loop and the header/separator lines. Replace with:

```go
b.WriteString(m.buildTable(procs))
```

Keep the sort hint line above.

#### Step 5 — Remove `colorCPU()` helper (now handled inline in StyleFunc)

If `colorCPU()` was added in Task 1 as a standalone helper, it's now superseded. Remove it to avoid dead code. The coloring lives entirely in `StyleFunc`.

#### Step 6 — Verify and commit

```bash
go build ./cmd/gone/
go test ./...
```

Expected: process list renders as a bordered table with styled header row, alternating row shades, and CPU% column color-coded by severity.

Commit message:
```
feat(tui): upgrade process table to lipgloss/v2/table with styled rows
```

---

## Post-Phase Verification

After all 4 tasks across 3 phases:

```bash
go build ./cmd/gone/
go test ./...
./gone
```

Manual smoke test checklist:
- [ ] Monitor tab: 4 gauges animate smoothly on each refresh (spring fill)
- [ ] Monitor tab: gauge boxes have gradient purple→cyan borders
- [ ] Monitor tab: CPU% values color-coded (green/yellow/red)
- [ ] Monitor tab: process table renders with styled header + alternating rows
- [ ] Uninstall tab: select items → Enter → confirmation modal appears
- [ ] Uninstall tab: [Enter] on modal trashes, [Esc] cancels cleanly
- [ ] Uninstall tab: cursor row has full-width navy background highlight
- [ ] Help overlay (`?`): still works
- [ ] Tab switching: no freezes, refresh continues in background
- [ ] `go test ./...`: all green

Final commit (if needed):
```
chore: bump CHANGELOG for beautification phase
```
