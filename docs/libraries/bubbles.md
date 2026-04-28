# Bubbles v2 (charm.land/bubbles/v2 v2.1.0)

## Components Available
| Component | Purpose |
|-----------|---------|
| `textinput.Model` | Search bars, filter inputs |
| `spinner.Model` | Loading indicators |
| `progress.Model` | Gauge bars (CPU, RAM, RX/TX) |
| `viewport.Model` | Scrollable content |
| `list.Model` | File list, selectable items |
| `table.Model` | Tabular data display |
| `help.Model` | Help overlays, footer hints |
| `key.Binding` | Keybinding definitions |
| `textarea.Model` | Multi-line text input |
| `paginator.Model` | Pagination controls |
| `timer.Model` | Countdown timer |
| `stopwatch.Model` | Elapsed time |

## Import Path (v2)
```go
import (
    "charm.land/bubbles/v2/textinput"
    "charm.land/bubbles/v2/progress"
    "charm.land/bubbles/v2/viewport"
    "charm.land/bubbles/v2/table"
    "charm.land/bubbles/v2/list"
    "charm.land/bubbles/v2/spinner"
    "charm.land/bubbles/v2/help"
    "charm.land/bubbles/v2/key"
)
```

## progress.Model — Gauge/Progress Bar

### Creation
```go
func newGaugeBar() progress.Model {
    return progress.New(
        progress.WithColors(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")),
        progress.WithoutPercentage(),
        progress.WithWidth(20),
    )
}
```

### SetPercent() — Pointer Receiver!
```go
m.cpuBar.SetPercent(0.75)  // Mutates in place, returns nothing
```

### Animation — MUST propagate FrameMsg
```go
case progress.FrameMsg:
    var cmd tea.Cmd
    m.cpuBar, cmd = m.cpuBar.Update(msg)
    cmds = append(cmds, cmd)
    return m, tea.Batch(cmds...)
```

### Options
- `WithColors(color1, color2)` — gradient colors
- `WithoutPercentage()` — hide "45%" text
- `WithWidth(n)` — bar width in cells
- `WithGradient(color1, color2)` — use gradient fill

## table.Model — Tabular Data

### Basic Usage
```go
columns := []table.Column{
    {Title: "PID", Width: 6},
    {Title: "Name", Width: 20},
    {Title: "CPU%", Width: 8},
}

rows := []table.Row{
    {"1234", "chrome", "45.2"},
    {"5678", "gone", "12.1"},
}

t := table.New().
    Headers("PID", "Name", "CPU%")...
```

### StyleFunc for Conditional Styling
```go
StyleFunc func(row, col int) lipgloss.Style

t := table.New().
    StyleFunc(func(row, col int) lipgloss.Style {
        if row == table.HeaderRow {
            return headerStyle
        }
        if row == m.cursor {
            return cursorStyle
        }
        if col == 2 && row >= 0 {
            pct := procs[row].CPU
            switch {
            case pct >= 70.0:
                return cellStyle.Foreground(lipgloss.Color("#FF6B6B"))
            case pct >= 30.0:
                return cellStyle.Foreground(lipgloss.Color("#FFDD57"))
            }
        }
        return evenRow
    })
```

### View() returns string
```go
return baseStyle.Render(m.table.View()) + "\n" + m.table.HelpView()
```

## list.Model — Selectable List

### Item Interface
```go
type Item interface {
    Title() string
    Description() string
    FilterValue() string
}
```

### Custom Delegate
```go
type fileDelegate struct {
    styles  Styles
    maxSize int64
}

func (d fileDelegate) Height() int                             { return 1 }
func (d fileDelegate) Spacing() int                            { return 0 }
func (d fileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d fileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    f, ok := item.(fileItem)
    if !ok { return }

    cursor := "  "
    if index == m.Index() {
        cursor = "> "
    }

    line := fmt.Sprintf("%s%s %s", cursor, f.path, f.size)
    fmt.Fprint(w, d.styles.CursorRow.Width(m.Width()).Render(line))
}
```

## textinput.Model — Text Input

### Creation and Focus
```go
ti := textinput.New()
ti.Placeholder = "filter by name..."
ti.CharLimit = 40
ti.Focus()
```

### Blur and Reset
```go
m.filterInput.Blur()
m.filterInput.Reset()
```

### Virtual Cursor
```go
ti.SetVirtualCursor(false)  // hide cursor blink
```

## spinner.Model — Loading Indicator

### Creation
```go
s := spinner.New(
    spinner.WithSpinner(spinner.Globe),
    spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6"))),
)
```

### Available Spinners
- `spinner.Dot` (default)
- `spinner.Globe`
- `spinner.Moon`
- `spinner.Pipe`
- `spinner.Points`

## viewport.Model — Scrollable Content

```go
vp := viewport.New()
vp.SetWidth(w)
vp.SetHeight(h)
vp.SetContent(content)
```

## help.Model — Contextual Help

```go
// Full help overlay
hv := help.New()
hv.ShowAll = true

// Footer short help
fh := help.New()
fh.ShowAll = false

// Key bindings
hv.View(m.keys)  // returns string with formatted keybindings
```

## key.Binding — Keybindings

```go
Tab: key.NewBinding(
    key.WithKeys("tab"),
    key.WithHelp("tab", "switch tabs"),
),

// Multiple keys
NavUD: key.NewBinding(
    key.WithKeys("up", "k", "down", "j"),
    key.WithHelp("↑/↓", "navigate"),
),

// Matching keys
if key.Matches(msg, m.keys.Tab) {
    // handle
}
```

## References
- https://pkg.go.dev/charm.land/bubbles/v2
- https://github.com/charmbracelet/bubbles/tree/main/examples