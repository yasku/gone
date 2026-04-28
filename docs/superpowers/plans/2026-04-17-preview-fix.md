# Task 0: Fix Preview Pane Alignment

## Problem

In `internal/tui/uninstall.go`, `lipgloss.JoinHorizontal` positions the preview pane
relative to the actual rendered width of `m.list.View()`, not the allocated `listW`.
When list items are shorter than `listW`, the preview drifts left instead of anchoring
to the right side of the terminal aligned with the search bar.

## Fix

In the `View()` method of `UninstallModel`, replace the existing `showPreview` block
(currently inside the `} else if len(m.list.Items()) > 0 {` branch) with this:

```go
} else if len(m.list.Items()) > 0 {
    if m.showPreview {
        total := m.width - 2
        listW := total * 3 / 5
        previewContentW := total - listW - 4
        listView := lipgloss.NewStyle().Width(listW).Render(m.list.View())
        previewView := m.styles.Preview.
            Width(previewContentW).
            Height(m.height - 8).
            Render(m.viewport.View())
        b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listView, previewView))
    } else {
        b.WriteString(m.list.View())
    }
```

Key change: `lipgloss.NewStyle().Width(listW).Render(m.list.View())` forces the list
to exactly `listW` characters wide so the preview always starts at the correct position.

## File to edit

`internal/tui/uninstall.go`

## Verify

Run both commands and ensure they succeed with zero errors:

```bash
go test ./...
go build ./cmd/gone/
```

## Commit message

```
fix(tui): force list fixed width so preview pane aligns correctly
```
