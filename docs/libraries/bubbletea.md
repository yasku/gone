# Bubble Tea v2 (charm.land/bubbletea/v2 v2.0.5)

## Key Concepts

### Model-View-Update Pattern
```go
type Model struct {
    field type
}

func (m Model) Init() tea.Cmd           // Return initial command(s)
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)  // Handle messages
func (m Model) View() tea.View          // Return tea.View, NOT string
```

### View() Must Return tea.View
In v2, `View()` returns `tea.View`, not `string`:
```go
// CORRECT
func (m model) View() tea.View {
    return tea.NewView("Hello!")
}

// CORRECT with fields
func (m model) View() tea.View {
    v := tea.NewView("")
    v.AltScreen = true
    v.SetContent(content)
    return v
}
```

### tea.NewView() vs tea.View Struct
`tea.NewView()` creates a View with content. `tea.View` is the struct with fields:
- `AltScreen bool` — use alternate screen buffer
- `MouseMode tea.MouseMode` — enable mouse
- `Content string` — Set via `SetContent()` or `NewView(content)`
- `Cursor *Cursor` — cursor position for textinput

### Value Receivers
Models use value receivers. Mutations via reassignment:
```go
model, cmd := model.Update(msg)  // reassign!
```

### tea.Cmd = func() tea.Msg
```go
type Cmd func() Msg

// Timer pattern
func doRefresh() tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
}

// Async work
func killProc(pid int32) tea.Cmd {
    return func() tea.Msg {
        err := syscall.Kill(int(pid), syscall.SIGTERM)
        return killDoneMsg{pid: pid, err: err}
    }
}
```

### tea.Batch()
Combine multiple commands:
```go
return m, tea.Batch(cmd1, cmd2, cmd3)
```

### tea.Tick()
```go
tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return refreshMsg(t) })
```

### tea.WindowSizeMsg
Sent when terminal resizes. Parent propagates to children via SetSize().

### Custom Messages
```go
type refreshMsg time.Time
type scanItemMsg struct { item fileItem }
```

### KeyPressMsg (not KeyMsg!)
v2 uses `tea.KeyPressMsg`, not `tea.KeyMsg`:
```go
case tea.KeyPressMsg:
    switch msg.String() {
    case "q", "ctrl+c":
        return m, tea.Quit
    case "enter":
        // handle
    }
```

## Program Lifecycle
1. `tea.NewProgram(model)` creates program
2. `p.Run()` starts event loop
3. `Init()` called once → returns initial commands
4. `Update()` called on every message
5. `View()` called after Update → renders UI

## tea.NewProgram Options (v1 vs v2)
In v2, program options moved to View fields:
```go
// v1
p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

// v2 — declarative in View()
func (m model) View() tea.View {
    v := tea.NewView("")
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    return v
}
```

## View Fields (v2)
| Field | Purpose |
|-------|---------|
| `Content` | Rendered string (set via `SetContent()`) |
| `AltScreen` | Enter/exit alternate screen buffer |
| `MouseMode` | `MouseModeNone`, `MouseModeCellMotion`, `MouseModeAllMotion` |
| `ReportFocus` | Enable focus/blur events |
| `Cursor` | Control cursor position, shape, blink |

## References
- https://pkg.go.dev/charm.land/bubbletea/v2
- https://github.com/charmbracelet/bubbletea/tree/main/examples