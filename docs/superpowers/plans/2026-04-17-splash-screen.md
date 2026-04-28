# Task 0: Splash Screen with Animated Banner

## Goal

Add a gorgeous startup splash screen to `gone` that displays for ~800ms before
transitioning to the main app. Uses only libraries already in go.mod:
`charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`.

## Design

```
╔══════════════════════════════════════════╗   ← gradient border purple→cyan
║                                          ║
║   ░██████╗░░█████╗░███╗░░██╗███████╗    ║
║   ██╔════╝░██╔══██╗████╗░██║██╔════╝    ║
║   ██║░░██╗░██║░░██║██╔██╗██║█████╗░░    ║
║   ██║░░╚██╗██║░░██║██║╚████║██╔══╝░░    ║
║   ╚██████╔╝╚█████╔╝██║░╚███║███████╗    ║
║   ░╚═════╝░░╚════╝░╚═╝░░╚══╝╚══════╝    ║
║                                          ║
║         hunt. select. trash.             ║
║                                          ║
║   ⣾  initializing...                    ║   ← spinner (Globe style)
╚══════════════════════════════════════════╝
```

- Border: `lipgloss.RoundedBorder()` with `BorderForegroundBlend` from `#9B59B6` → `#00BCD4`
- ASCII logo: hardcoded string, rendered with `lipgloss.NewStyle().Foreground(lipgloss.Color("#00BCD4"))`
- Tagline: `"hunt. select. trash."` dimmed gray `Color("245")`
- Spinner: `bubbles/v2/spinner` with `spinner.Globe` style, color `Color("#9B59B6")`
- Duration: 800ms via `tea.Tick`, then emit `splashDoneMsg{}` to transition

## Critical API Patterns (use verbatim — these are verified for v2)

### spinner bubbles/v2 — WithSpinner/WithStyle options API
```go
import "charm.land/bubbles/v2/spinner"

s := spinner.New(
    spinner.WithSpinner(spinner.Globe),   // Globe, Dot, Line, Pulse, MiniDot, Hamburger
    spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6"))),
)
// In Init(): return s.Tick
// In Update(): case spinner.TickMsg: m.spinner, cmd = m.spinner.Update(msg)
// In View(): m.spinner.View()
```

### lipgloss/v2 — Place (centering in terminal)
```go
// Centers content in the terminal
centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
```

### lipgloss/v2 — BorderForegroundBlend (gradient border)
```go
style := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForegroundBlend(
        lipgloss.Color("#9B59B6"),  // start color
        lipgloss.Color("#00BCD4"),  // end color
    ).
    Padding(1, 4)
```

### bubbletea/v2 — AltScreen via tea.View
```go
func (m Model) View() tea.View {
    v := tea.NewView(content)
    v.AltScreen = true
    return v
}
```

### bubbletea/v2 — WindowSizeMsg (always use type assertion)
```go
case tea.WindowSizeMsg:
    msg := msg.(tea.WindowSizeMsg)
    m.width = msg.Width
    m.height = msg.Height
```

## Files to create/modify

### New file: `internal/tui/splash.go`

```go
package tui

import (
    "time"
    "strings"

    "charm.land/bubbles/v2/spinner"
    tea "charm.land/bubbletea/v2"
    "charm.land/lipgloss/v2"
)

type splashDoneMsg struct{}

func splashDone() tea.Cmd {
    return tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
        return splashDoneMsg{}
    })
}

const goneLogo = `
  ░██████╗░░█████╗░███╗░░██╗███████╗
  ██╔════╝░██╔══██╗████╗░██║██╔════╝
  ██║░░██╗░██║░░██║██╔██╗██║█████╗░░
  ██║░░╚██╗██║░░██║██║╚████║██╔══╝░░
  ╚██████╔╝╚█████╔╝██║░╚███║███████╗
  ░╚═════╝░░╚════╝░╚═╝░░╚══╝╚══════╝`

type SplashModel struct {
    spinner spinner.Model
    width   int
    height  int
    done    bool
}

func NewSplashModel() SplashModel {
    s := spinner.New(
        spinner.WithSpinner(spinner.Globe),
        spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6"))),
    )
    return SplashModel{spinner: s}
}

func (m SplashModel) Init() tea.Cmd {
    return tea.Batch(m.spinner.Tick, splashDone())
}

func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
    switch msg.(type) {
    case splashDoneMsg:
        m.done = true
        return m, nil
    case tea.WindowSizeMsg:
        msg := msg.(tea.WindowSizeMsg)
        m.width = msg.Width
        m.height = msg.Height
        return m, nil
    case spinner.TickMsg:
        var cmd tea.Cmd
        m.spinner, cmd = m.spinner.Update(msg)
        return m, cmd
    }
    return m, nil
}

func (m SplashModel) View() string {
    logoStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00BCD4")).
        Bold(true)

    taglineStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("245")).
        Italic(true)

    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForegroundBlend(
            lipgloss.Color("#9B59B6"),
            lipgloss.Color("#00BCD4"),
        ).
        Padding(1, 4)

    content := strings.Join([]string{
        logoStyle.Render(goneLogo),
        "",
        taglineStyle.Render("hunt. select. trash."),
        "",
        m.spinner.View() + "  initializing...",
    }, "\n")

    box := boxStyle.Render(content)
    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
```

### Modify: `internal/tui/app.go`

Add `showSplash bool` and `splash SplashModel` fields to `AppModel`:

```go
type AppModel struct {
    active     activeTab
    uninstall  UninstallModel
    monitor    MonitorModel
    splash     SplashModel
    styles     Styles
    width      int
    height     int
    ready      bool
    showSplash bool
    showHelp   bool
}
```

Update `NewApp()` to enable splash on start:

```go
func NewApp() AppModel {
    return AppModel{
        active:     tabUninstall,
        uninstall:  NewUninstallModel(),
        monitor:    NewMonitorModel(),
        splash:     NewSplashModel(),
        styles:     DefaultStyles(),
        showSplash: true,
    }
}
```

Update `Init()` to include splash:

```go
func (m AppModel) Init() tea.Cmd {
    return tea.Batch(
        m.splash.Init(),
        m.uninstall.Init(),
        m.monitor.Init(),
    )
}
```

Update `Update()` — handle splash done and route msgs while splashing:

```go
// At the top of the switch in Update():
case splashDoneMsg:
    m.showSplash = false
    return m, nil
```

Also forward `spinner.TickMsg` and `tea.WindowSizeMsg` to splash while active:

```go
// Before the existing WindowSizeMsg handler, add:
if m.showSplash {
    var cmd tea.Cmd
    m.splash, cmd = m.splash.Update(msg)
    cmds = append(cmds, cmd)
}
```

Update `View()` — if `showSplash`, return splash view instead of normal UI:

```go
func (m AppModel) View() tea.View {
    if !m.ready {
        v := tea.NewView("loading...")
        v.AltScreen = true
        return v
    }
    if m.showSplash {
        v := tea.NewView(m.splash.View())
        v.AltScreen = true
        return v
    }
    // ... rest of existing View() unchanged
```

## Verify

```bash
go build ./cmd/gone/
go test ./...
```

Both must pass with zero errors.

## Commit message

```
feat(tui): add animated splash screen with gradient banner and spinner
```
