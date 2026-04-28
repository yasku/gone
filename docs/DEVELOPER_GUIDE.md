# gone — Developer Guide

## Prerequisites

### Required

- **Go 1.26+** — The project uses Go 1.26.1
- **macOS** — Development requires macOS for `osascript` integration
- **Git** — Version control

### Optional (for enhanced functionality)

```bash
brew install fd osquery glances bmon mtr nmap
```

## Setup Development Environment

### 1. Clone the Repository

```bash
git clone https://github.com/yasku/gone.git
cd gone
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Verify Build

```bash
go build -o gone ./cmd/gone
./gone
```

### 4. Run Tests

```bash
./scripts/test-all.sh
```

## Project Structure Walkthrough

```
gone/
├── cmd/gone/main.go             Entry point
├── internal/
│   ├── cli/                     External tool wrappers
│   │   ├── runner.go            ExecJSON, ExecStream, ExecSimple
│   │   ├── tool.go               Tool discovery
│   │   ├── fd.go                fd wrapper (optional)
│   │   └── osquery.go           osquery wrapper (optional)
│   ├── scanner/                 File system scanner
│   │   ├── scanner.go           SearchStream with fastwalk
│   │   ├── locations.go         Scan paths, skip lists
│   │   └── rcscanner.go         Shell RC file scanner
│   ├── remover/                 Trash operations
│   │   ├── trash.go              MoveToTrash via osascript
│   │   └── log.go                JSONL operation log
│   ├── sysinfo/                 gopsutil wrapper
│   │   └── sysinfo.go            System metrics
│   └── tui/                     All Bubble Tea models
│       ├── app.go               Root model, tab routing
│       ├── splash.go            Startup splash
│       ├── uninstall.go         Search → scan → trash
│       ├── monitor.go           CPU/RAM gauges + process table
│       ├── network.go           Network RX/TX gauges
│       ├── logs.go              Log viewer
│       ├── audit.go             osquery security audit
│       └── styles.go            Lip Gloss styles
├── scripts/
│   ├── test-all.sh              Full test suite
│   ├── check-tools.sh            Tool availability checker
│   └── coverage.sh               Coverage reports
└── docs/
    ├── ARCHITECTURE.md           This architecture doc
    ├── SETUP.md                  Installation guide
    ├── USER_GUIDE.md             User documentation
    └── DEVELOPER_GUIDE.md        This file
```

## Running Tests

### Full Test Suite

```bash
./scripts/test-all.sh
```

Runs:
- `go test ./cmd/... ./internal/... -v`
- `go test -race ./internal/...`
- `go vet ./cmd/... ./internal/...`
- `go fmt` check

### Individual Test Commands

```bash
# Test all packages
go test ./...

# Test specific package
go test ./internal/tui/... -v

# Run with race detector
go test -race ./internal/...

# Run benchmarks
go test -bench=. ./internal/scanner/

# Vet all packages
go vet ./...

# Format all packages
go fmt ./...
```

## How to Add a New Tab

Adding a new tab requires modifying `internal/tui/app.go` to:

1. Add a new `activeTab` constant
2. Add the model field to `AppModel`
3. Initialize the model in `NewApp()`
4. Add `SetSize()` propagation in `Update()`
5. Add message routing in `Update()`
6. Add view rendering in `View()`
7. Update tab cycling in key handler

### Step-by-Step Example

**1. Define the tab constant** (app.go):

```go
const (
    tabUninstall activeTab = iota
    tabMonitor
    tabNetwork
    tabLogs
    tabAudit
    tabNewTab  // NEW
)
```

**2. Add model field to AppModel** (app.go):

```go
type AppModel struct {
    // ... existing fields
    newTab     NewTabModel  // NEW
}
```

**3. Initialize in NewApp()** (app.go):

```go
return AppModel{
    // ... existing initialization
    newTab: NewNewTabModel(),
}
```

**4. Add SetSize() propagation** (app.go, tea.WindowSizeMsg case):

```go
case tea.WindowSizeMsg:
    // ... existing SetSize calls
    m.newTab = m.newTab.SetSize(msg.Width, contentHeight)
```

**5. Add message routing** (app.go, Update()):

```go
case newTabRefreshMsg:
    var cmd tea.Cmd
    m.newTab, cmd = m.newTab.Update(msg)
    cmds = append(cmds, cmd)

// Route to active tab:
case tabNewTab:
    var cmd tea.Cmd
    m.newTab, cmd = m.newTab.Update(msg)
    cmds = append(cmds, cmd)
```

**6. Add view rendering** (app.go, View()):

```go
switch m.active {
case tabUninstall:
    tabContent = m.uninstall.View().Content
// ... existing cases
case tabNewTab:  // NEW
    tabContent = m.newTab.View().Content
}
```

**7. Update tab cycling** (app.go, tab key handler):

```go
if msg.String() == "tab" {
    m.active = (m.active + 1) % 6  // Changed from 5 to 6
    m.keys.active = m.active
    return m, nil
}
```

## How to Add a New CLI Tool Integration

### Pattern 1: Wrapper in internal/cli/

**1. Create tool wrapper** (e.g., `internal/cli/newtool.go`):

```go
package cli

import "fmt"

func RunNewTool(args []string) (string, error) {
    path, err := Which("newtool")
    if err != nil {
        return "", err
    }
    out, err := NewRunner(30 * time.Second).ExecSimple(path, args)
    if err != nil {
        return "", fmt.Errorf("newtool: %w", err)
    }
    return string(out), nil
}
```

**2. Add to tool.go discovery:**

```go
// In Which() function, add to pre-populated tools:
// "newtool",
```

### Pattern 2: Graceful Degradation

```go
func FeatureWithOptionalTool() string {
    if !IsAvailable("optionaltool") {
        return "optionaltool not installed - install with: brew install optionaltool"
    }
    // Use the tool
    result, err := RunOptionalTool()
    if err != nil {
        return fmt.Sprintf("error: %v", err)
    }
    return result
}
```

## Code Style Conventions

### Go Style

- Use `go fmt` and `go vet` before committing
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable names
- Keep functions small and focused

### Bubble Tea v2 Conventions

1. **Value receivers for all models**:
   ```go
   func (m Model) Update(msg tea.Msg) (Model, tea.Cmd)
   ```

2. **View() returns tea.View, NOT string**:
   ```go
   func (m Model) View() tea.View {
       v := tea.NewView(content)
       v.AltScreen = true
       return v
   }
   ```

3. **Always propagate progress.FrameMsg**:
   ```go
   case progress.FrameMsg:
       var cmd tea.Cmd
       m.progressBar, cmd = m.progressBar.Update(msg)
       cmds = append(cmds, cmd)
   ```

4. **Use charm.land/*/v2 import paths**:
   ```go
   import (
       "charm.land/bubbletea/v2"
       "charm.land/bubbles/v2/progress"
       "charm.land/lipgloss/v2"
   )
   ```

### Lip Gloss Conventions

1. **Block centering with lipgloss.Place**:
   ```go
   content := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Left, innerContent)
   ```

2. **Style chains**:
   ```go
   lipgloss.NewStyle().
       Bold(true).
       Foreground(lipgloss.Color("#00BCD4")).
       Padding(0, 1).
       Border(lipgloss.RoundedBorder())
   ```

3. **Color types**:
   - Hex: `lipgloss.Color("#00BCD4")`
   - ANSI 256: `lipgloss.Color("240")`
   - Gradient: `lipgloss.Blend1D(n, color1, color2)`

## Commit Message Format

```
<type>: <short description>

<longer description if needed>

<issue number if applicable>
```

### Types

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation changes
- `style:` — Formatting, no code change
- `refactor:` — Code restructuring
- `test:` — Adding/updating tests
- `chore:` — Maintenance, dependencies

### Examples

```
feat: add Network tab with RX/TX gauges

Added new Network tab that displays per-interface network I/O
with animated gauges and interface filtering.

Closes #45
```

```
fix: progress bar animation freeze on Monitor tab

Propagation of progress.FrameMsg was being swallowed by the
refreshMsg filter. Now properly propagated in all cases.

Fixes #67
```

## PR Workflow

1. **Fork the repository** on GitHub

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-feature
   ```

3. **Make changes** with proper testing:
   ```bash
   # Make changes
   ./scripts/test-all.sh  # Must pass
   ```

4. **Commit** (follow commit message format):
   ```bash
   git add .
   git commit -m "feat: add my feature"
   ```

5. **Push** to your fork:
   ```bash
   git push origin feature/my-feature
   ```

6. **Open PR** on GitHub with:
   - Clear title
   - Description of changes
   - Reference to issue if applicable

7. **Ensure CI passes** before requesting review

## Important Rules

1. **Never commit without running `./scripts/test-all.sh`**
2. **Always use value receivers** for Bubble Tea models
3. **Always propagate `progress.FrameMsg`** to all progress bars
4. **Return `tea.View` from `View()`**, NOT string
5. **Use `charm.land/*/v2`** import paths, NOT `github.com/charmbracelet/*`
6. **Test on macOS** — Linux/Windows builds may hide platform-specific bugs

## Library Versions

| Library | Version | Import Path |
|---------|---------|-------------|
| Bubble Tea | v2.0.5 | `charm.land/bubbletea/v2` |
| Bubbles | v2.1.0 | `charm.land/bubbles/v2` |
| Lip Gloss | v2.0.3 | `charm.land/lipgloss/v2` |
| gopsutil | v4.26.3 | `github.com/shirou/gopsutil/v4` |

**Critical:** Import paths use `charm.land/*/v2`, NOT `github.com/charmbracelet/*`. v1 docs are outdated.

## Additional Resources

- [Bubble Tea v2 Docs](https://pkg.go.dev/charm.land/bubbletea/v2)
- [Bubbles v2 Docs](https://pkg.go.dev/charm.land/bubbles/v2)
- [Lip Gloss v2 Docs](https://pkg.go.dev/charm.land/lipgloss/v2)
- [gopsutil v4 Docs](https://pkg.go.dev/github.com/shirou/gopsutil/v4)
