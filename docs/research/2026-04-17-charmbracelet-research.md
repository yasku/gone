# Charmbracelet Library Research
**Date:** 2026-04-17
**Scope:** charmbracelet/bubbles v2.1.0 · charmbracelet/lipgloss v2.0.3
**Purpose:** Implementation reference for high-quality TUI design (gone — macOS uninstaller/system monitor)

---

## Table of Contents

1. [bubbles — Component Library](#1-bubbles--component-library)
   - [Import Paths (v2)](#import-paths-v2)
   - [Global v2 Patterns](#global-v2-patterns)
   - [Components](#components)
     - [list](#list)
     - [table](#table)
     - [viewport](#viewport)
     - [textinput](#textinput)
     - [textarea](#textarea)
     - [spinner](#spinner)
     - [progress](#progress)
     - [paginator](#paginator)
     - [filepicker](#filepicker)
     - [help + key](#help--key)
     - [timer](#timer)
     - [stopwatch](#stopwatch)
     - [cursor](#cursor)
2. [lipgloss — Styling & Layout](#2-lipgloss--styling--layout)
   - [Import Path (v2)](#import-path-v2)
   - [Color System](#color-system)
   - [Style API](#style-api)
   - [Block-Level Formatting](#block-level-formatting)
   - [Borders](#borders)
   - [Layout Utilities](#layout-utilities)
   - [Compositing (Canvas + Layer)](#compositing-canvas--layer)
   - [Sub-packages: table · list · tree](#sub-packages-table--list--tree)
   - [Advanced Color Features](#advanced-color-features)
3. [Power Features](#3-power-features)
4. [UX/UI Patterns for System TUIs](#4-uxui-patterns-for-system-tuis)

---

## 1. bubbles — Component Library

**Repo:** https://github.com/charmbracelet/bubbles
**Latest:** v2.1.0 (2026-03-26) · 8.2k stars
**Requires:** `charm.land/bubbletea/v2` + `charm.land/lipgloss/v2`

### Import Paths (v2)

```go
import (
    "charm.land/bubbles/v2/cursor"
    "charm.land/bubbles/v2/filepicker"
    "charm.land/bubbles/v2/help"
    "charm.land/bubbles/v2/key"
    "charm.land/bubbles/v2/list"
    "charm.land/bubbles/v2/paginator"
    "charm.land/bubbles/v2/progress"
    "charm.land/bubbles/v2/spinner"
    "charm.land/bubbles/v2/stopwatch"
    "charm.land/bubbles/v2/table"
    "charm.land/bubbles/v2/textarea"
    "charm.land/bubbles/v2/textinput"
    "charm.land/bubbles/v2/timer"
    "charm.land/bubbles/v2/viewport"
)
```

### Global v2 Patterns

| Pattern | v1 | v2 |
|---|---|---|
| Key messages | `tea.KeyMsg` | `tea.KeyPressMsg` |
| Width/Height fields | `m.Width = 40` | `m.SetWidth(40)` / `m.Width()` |
| Default keymaps | `pkg.DefaultKeyMap` (var) | `pkg.DefaultKeyMap()` (func) |
| Adaptive colors | `lipgloss.AdaptiveColor{}` | `lipgloss.LightDark(isDark)` |
| Constructor aliases | `NewModel()` | `New()` only |
| Styles for dark/light | automatic | pass `isDark bool` explicitly |

The `runeutil` and `memoization` packages are now internal — not importable.

---

### Components

#### list

Feature-rich scrollable list with built-in pagination, fuzzy filtering, activity spinner, status messages, and auto-generated help. Extrapolated from Glow.

**Key types:**
- `list.Item` interface — requires `FilterValue() string`
- `list.ItemDelegate` interface — `Render`, `Height`, `Spacing`, `Update`
- `list.DefaultDelegate` — ready-to-use delegate with title + description

**Initialization:**
```go
items := []list.Item{...}
l := list.New(items, list.NewDefaultDelegate(), width, height)
l.Title = "Apps"
l.SetShowStatusBar(true)
l.SetFilteringEnabled(true)
```

**Key methods:**
```go
l.SelectedItem() list.Item     // currently highlighted item
l.Items() []list.Item          // all items (unfiltered)
l.SetItems([]list.Item)        // replace entire list
l.InsertItem(index, item)      // insert at index
l.RemoveItem(index)            // remove by index
l.SetSize(w, h)                // resize (call on tea.WindowSizeMsg)
l.SetDelegate(d)               // swap delegate
l.FilterState()                // Unfiltered | Filtering | FilterApplied
l.SetFilteringEnabled(bool)
l.SetShowHelp(bool)
l.SetShowStatusBar(bool)
l.SetShowPagination(bool)
l.NewStatusMessage(msg)        // temporary status bar message
```

**Styles (v2):**
```go
styles := list.DefaultStyles(isDark)
// styles.Title, styles.TitleBar, styles.Spinner, styles.FilterPrompt → now in styles.Filter.Focused/Blurred
l.Styles = styles
```

**Minimal example:**
```go
type item struct{ title, desc string }
func (i item) FilterValue() string { return i.title }
func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }

l := list.New([]list.Item{
    item{"Slack", "2.1 GB"},
    item{"Xcode", "14.3 GB"},
}, list.NewDefaultDelegate(), 40, 20)
```

---

#### table

Navigable table with vertical scrolling, custom column widths, and full style control.

**Initialization:**
```go
columns := []table.Column{
    {Title: "Name", Width: 20},
    {Title: "Size", Width: 10},
}
rows := []table.Row{
    {"Slack", "2.1 GB"},
    {"Xcode", "14.3 GB"},
}
t := table.New(
    table.WithColumns(columns),
    table.WithRows(rows),
    table.WithFocused(true),
    table.WithHeight(10),
)
```

**Key methods:**
```go
t.SelectedRow() table.Row       // nil-safe, returns currently highlighted row
t.Rows() []table.Row
t.SetRows([]table.Row)
t.SetColumns([]table.Column)
t.SetWidth(w) / t.Width()
t.SetHeight(h) / t.Height()
t.GotoTop() / t.GotoBottom()
t.MoveUp(n) / t.MoveDown(n)
t.Cursor() int                  // current row index
```

**Styles:**
```go
s := table.DefaultStyles()
s.Header = s.Header.Bold(true).BorderBottom(true).BorderStyle(lipgloss.NormalBorder())
s.Selected = s.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
t.SetStyles(s)
```

---

#### viewport

Vertically (and optionally horizontally) scrollable content pane. Ideal for log output, detail views, large text rendering.

**Initialization:**
```go
vp := viewport.New(
    viewport.WithWidth(80),
    viewport.WithHeight(24),
)
vp.SetContent(longString)
vp.SoftWrap = true           // word wrap vs. horizontal scroll
vp.FillHeight = true         // pad empty lines to fill height
vp.MouseWheelEnabled = true
vp.Style = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
```

**Key methods:**
```go
vp.SetContent(string)
vp.SetWidth(w) / vp.Width()
vp.SetHeight(h) / vp.Height()
vp.ScrollToTop() / vp.ScrollToBottom()
vp.LineDown(n) / vp.LineUp(n)
vp.HalfViewDown() / vp.HalfViewUp()
vp.AtTop() bool / vp.AtBottom() bool
vp.ScrollPercent() float64
vp.YOffset int              // current scroll position (settable)
```

**Gutter support (v2 new):**
```go
// Add a fixed left gutter (e.g., line numbers) that persists during horizontal scroll
vp.LeftGutterFunc = func(idx int, after bool) string {
    return fmt.Sprintf("%3d ", idx+1)
}
```

**Highlight support (v2 new):**
```go
vp.HighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("226"))
vp.SelectedHighlightStyle = lipgloss.NewStyle().Background(lipgloss.Color("208"))
vp.SetHighlights([]viewport.Highlight{{Start: 10, End: 20}})
vp.HighlightNext() / vp.HighlightPrevious()
```

---

#### textinput

Single-line text field with scroll, unicode, paste, tab-completion, validation, and echo modes.

**Initialization:**
```go
ti := textinput.New()
ti.Placeholder = "Search apps..."
ti.CharLimit = 100
ti.SetWidth(30)
ti.EchoMode = textinput.EchoNormal  // or EchoPassword, EchoNone
ti.ShowSuggestions = true
ti.SetSuggestions([]string{"Slack", "Xcode", "Figma"})
ti.Validate = func(s string) error {
    if strings.Contains(s, "/") { return errors.New("no slashes") }
    return nil
}
```

**Key methods:**
```go
ti.Focus() / ti.Blur()
ti.Value() string
ti.SetValue(string)
ti.Reset()
ti.Focused() bool
ti.Err error    // set when Validate fails
```

**Styles (v2):**
```go
s := textinput.DefaultStyles(isDark)
s.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
s.Focused.Text = lipgloss.NewStyle()
s.Blurred.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
ti.Styles = s
```

**Key map (v2):** `textinput.DefaultKeyMap()` returns a fresh `KeyMap` struct. Includes: `CharacterForward/Backward`, `WordForward/Backward`, `DeleteWord*`, `LineStart/End`, `Paste`, `AcceptSuggestion`, `NextSuggestion`, `PrevSuggestion`.

---

#### textarea

Multi-line text area with vertical scroll, line count, virtual or real cursor, page navigation.

**Initialization:**
```go
ta := textarea.New()
ta.Placeholder = "Notes..."
ta.SetWidth(60)
ta.SetHeight(10)
ta.Focus()
```

**Key methods:**
```go
ta.Value() string
ta.SetValue(string)
ta.InsertString(string)
ta.InsertRune(rune)
ta.LineCount() int
ta.Line() int           // current line
ta.Column() int         // current column (0-indexed)
ta.ScrollYOffset() int
ta.MoveToBeginning() / ta.MoveToEnd()
```

**Styles (v2):**
```go
ta.Styles.Focused  // type textarea.StyleState
ta.Styles.Blurred
ta.Styles.Cursor   // cursor appearance
ta.VirtualCursor = true   // false = use real terminal cursor
```

---

#### spinner

Animated loading indicator with configurable frames and style.

**Built-in spinners:** `spinner.Line`, `spinner.Dot`, `spinner.MiniDot`, `spinner.Jump`, `spinner.Pulse`, `spinner.Points`, `spinner.Globe`, `spinner.Moon`, `spinner.Monkey`

**Usage:**
```go
s := spinner.New(
    spinner.WithSpinner(spinner.Dot),
    spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
)

// In Init:
return s.Tick

// In Update:
case spinner.TickMsg:
    s, cmd = s.Update(msg)
    return m, cmd

// In View:
return s.View() + " Loading..."
```

**Custom spinner:**
```go
s.Spinner = spinner.Spinner{
    Frames: []string{"◐", "◓", "◑", "◒"},
    FPS:    time.Second / 8,
}
```

---

#### progress

Animated progress bar with spring physics (via harmonica), gradient/solid fills, percentage display.

**Options:**
```go
p := progress.New(
    progress.WithDefaultBlend(),       // purple→pink gradient
    progress.WithColors(lipgloss.Color("#5A56E0"), lipgloss.Color("#EE6FF8")),  // custom blend
    progress.WithColors(lipgloss.Color("#7571F9")),  // solid fill
    progress.WithScaled(true),         // scale blend to filled portion only
    progress.WithoutPercentage(),      // hide percentage
    progress.WithWidth(40),
    progress.WithColorFunc(func(total, current float64) color.Color { ... }),
)
```

**Fill characters:**
```go
p.FullChar = progress.DefaultFullCharHalfBlock  // '▌' (default, higher resolution blending)
p.FullChar = progress.DefaultFullCharFullBlock  // '█'
p.EmptyChar = progress.DefaultEmptyCharBlock    // '░'
```

**Usage:**
```go
// Set percentage (0.0–1.0), animated via spring:
cmd = p.SetPercent(0.75)

// In Update:
case progress.FrameMsg:
    pm, cmd := p.Update(msg)
    p = pm.(progress.Model)
    return m, cmd

// In View:
return p.View()
```

---

#### paginator

Lightweight pagination state controller with optional UI rendering. Two display modes: dots or Arabic numerals.

```go
p := paginator.New()
p.Type = paginator.Dots     // or paginator.Arabic ("3/10")
p.PerPage = 5
p.SetTotalPages(len(items))
p.ActiveDot = "●"
p.InactiveDot = "○"

// Get slice bounds for current page:
start, end := p.GetSliceBounds(len(items))
currentPage := items[start:end]

// Navigate:
p.NextPage() / p.PrevPage()
p.Page int         // current page (0-indexed)
p.TotalPages int
p.OnLastPage() bool

// Render dots/numbers:
p.View()
```

**Key map (v2):** `paginator.DefaultKeyMap()` — removed `UsePgUpPgDownKeys`, `UseLeftRightKeys` etc. Customize `KeyMap` directly.

---

#### filepicker

File system navigator with directory traversal, file filtering by extension, permissions/size display.

```go
fp := filepicker.New()
fp.CurrentDirectory = "/Applications"
fp.AllowedTypes = []string{".app"}
fp.ShowHidden = false
fp.ShowPermissions = true
fp.ShowSize = true
fp.DirAllowed = false
fp.FileAllowed = true
fp.AutoHeight = true

// In Init:
return fp.Init()

// In Update:
fp, cmd = fp.Update(msg)
if didSelect, path := fp.DidSelectFile(msg); didSelect {
    // path is the selected file
}
if didSelect, path := fp.DidSelectDisabledFile(msg); didSelect {
    // user tried to select a disallowed file
}
```

**Styles:**
```go
s := filepicker.DefaultStyles()
s.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
s.File = lipgloss.NewStyle()
s.Cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
s.Symlink = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
fp.Styles = s
```

---

#### help + key

`key` manages keybindings (with help text). `help` auto-generates a compact help view from those bindings.

**Defining bindings:**
```go
type keyMap struct {
    Up   key.Binding
    Down key.Binding
    Del  key.Binding
    Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Del, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down},
        {k.Del, k.Quit},
    }
}

var keys = keyMap{
    Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
    Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
    Del:  key.NewBinding(key.WithKeys("d", "delete"), key.WithHelp("d", "uninstall")),
    Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
```

**Matching in Update:**
```go
case tea.KeyPressMsg:
    switch {
    case key.Matches(msg, keys.Up):
        // handle up
    case key.Matches(msg, keys.Del):
        // handle delete
    }
```

**Rendering help:**
```go
h := help.New()
h.Styles = help.DefaultStyles(isDark)
h.SetWidth(termWidth)

// Toggle short/full help:
h.ShowAll = !h.ShowAll

// In View:
return h.View(keys)
```

---

#### timer

Countdown timer with configurable interval and timeout.

```go
t := timer.New(5*time.Minute, timer.WithInterval(time.Second))

// In Init: return t.Init()
// In Update:
case timer.TickMsg:
    t, cmd = t.Update(msg)
case timer.TimeoutMsg:
    // timer expired

t.Timeout time.Duration   // remaining time
t.Running() bool
t.Start() / t.Stop() / t.Toggle() / t.Reset()
```

---

#### stopwatch

Counts up with configurable interval.

```go
sw := stopwatch.New(stopwatch.WithInterval(100*time.Millisecond))

// In Init: return sw.Init()
// In Update:
case stopwatch.TickMsg:
    sw, cmd = sw.Update(msg)

sw.Elapsed() time.Duration
sw.Running() bool
sw.Start() / sw.Stop() / sw.Toggle() / sw.Reset()
```

---

#### cursor

Blinking cursor component for use inside custom inputs.

```go
c := cursor.New()
c.SetMode(cursor.CursorBlink)  // CursorStatic | CursorBlink | CursorHide

// In Update:
case cursor.BlinkMsg:
    c, cmd = c.Update(msg)

c.IsBlinked() bool
c.Blink() tea.Cmd  // start blink loop (v2 — was BlinkCmd() in v1)
```

---

## 2. lipgloss — Styling & Layout

**Repo:** https://github.com/charmbracelet/lipgloss
**Latest:** v2.0.3 (2026-04-13) · 11k stars

### Import Path (v2)

```go
import "charm.land/lipgloss/v2"

// Sub-packages:
import "charm.land/lipgloss/v2/table"
import "charm.land/lipgloss/v2/list"
import "charm.land/lipgloss/v2/tree"
```

**v2 breaking change:** `Renderer` is gone. `Style` is now a plain value type. Color downsampling happens at output layer via `lipgloss.Println` / `lipgloss.Fprintln`. With Bubble Tea v2, downsampling is automatic — nothing extra needed.

---

### Color System

`lipgloss.Color(s string) color.Color` returns `image/color.Color`.

```go
// ANSI 4-bit (0–15)
lipgloss.Color("5")   // magenta
lipgloss.Color("9")   // red

// ANSI 256 (8-bit)
lipgloss.Color("86")  // aqua
lipgloss.Color("201") // hot pink

// True color (24-bit hex)
lipgloss.Color("#7D56F4")
lipgloss.Color("#04B575")

// Named ANSI constants
lipgloss.Black, lipgloss.Red, lipgloss.Green, lipgloss.Yellow
lipgloss.Blue, lipgloss.Magenta, lipgloss.Cyan, lipgloss.White
lipgloss.BrightBlack ... lipgloss.BrightWhite

// No color
lipgloss.NoColor{}
```

**Color utilities:**
```go
c := lipgloss.Color("#EB4268")
dark := lipgloss.Darken(c, 0.5)
light := lipgloss.Lighten(c, 0.35)
green := lipgloss.Complementary(c)
withAlpha := lipgloss.Alpha(c, 0.2)
```

**Adaptive colors (dark/light terminal):**
```go
// With Bubble Tea: listen for tea.BackgroundColorMsg in Update
case tea.BackgroundColorMsg:
    m.isDark = msg.IsDark()
    m.styles = newStyles(m.isDark)

// Then use LightDark:
lightDark := lipgloss.LightDark(isDark)
fg := lightDark(lipgloss.Color("#333333"), lipgloss.Color("#f1f1f1"))

// Compat package (drop-in for v1 AdaptiveColor):
import "charm.land/lipgloss/v2/compat"
color := compat.AdaptiveColor{
    Light: lipgloss.Color("#333333"),
    Dark:  lipgloss.Color("#f1f1f1"),
}
```

**Complete color (per-profile):**
```go
import "github.com/charmbracelet/colorprofile"
profile := colorprofile.Detect(os.Stdout, os.Environ())
complete := lipgloss.Complete(profile)
color := complete(lipgloss.Color("5"), lipgloss.Color("200"), lipgloss.Color("#ff00ff"))
// args: ANSI, ANSI256, TrueColor — automatically picks best available
```

---

### Style API

`lipgloss.Style` is a value type — assignment creates a true copy.

```go
base := lipgloss.NewStyle().
    Bold(true).
    Italic(true).
    Faint(true).
    Strikethrough(true).
    Underline(true).
    Reverse(true).
    Foreground(lipgloss.Color("#FAFAFA")).
    Background(lipgloss.Color("#7D56F4"))

// Render
output := base.Render("Hello")

// Copy and extend
warning := base.Background(lipgloss.Color("#FF6600"))

// SetString — bakes a string into the style; call .String() or pass to Println
s := lipgloss.NewStyle().Bold(true).SetString("default text")
```

**Inline formatting:**
```go
s := lipgloss.NewStyle().
    UnderlineStyle(lipgloss.UnderlineCurly).   // None|Single|Double|Curly|Dotted|Dashed
    UnderlineColor(lipgloss.Color("#FF0000"))
```

**Hyperlinks (v2):**
```go
s := lipgloss.NewStyle().Hyperlink("https://example.com")
// Degrades gracefully in unsupported terminals
```

**Inheritance:**
```go
// Only unset rules are inherited
child := lipgloss.NewStyle().
    Foreground(lipgloss.Color("201")).
    Inherit(parentStyle)  // inherits bg from parent, fg stays 201
```

**Unsetting rules:**
```go
s = s.UnsetBold().UnsetBackground().UnsetForeground()
```

**Enforcing constraints:**
```go
s.Inline(true).Render("force single line")
s.Inline(true).MaxWidth(20).Render("clipped")
s.MaxWidth(20).MaxHeight(5).Render("box limit")
```

**Tab handling:**
```go
s = s.TabWidth(4)                      // render tabs as N spaces
s = s.TabWidth(0)                      // remove tabs
s = s.TabWidth(lipgloss.NoTabConversion) // leave tabs as-is
```

**Wrap:**
```go
wrapped := lipgloss.Wrap(styledText, 40, " ")
```

**Get values from a style:**
```go
s.GetForeground()  // color.Color
s.GetWidth() int
s.GetBold() bool
// ... GetX() for every property
```

---

### Block-Level Formatting

**Width, Height, Alignment:**
```go
s := lipgloss.NewStyle().
    Width(40).
    Height(10).
    Align(lipgloss.Left).   // Left | Center | Right
    AlignVertical(lipgloss.Top)  // Top | Center | Bottom
```

**Padding (inside border):**
```go
s.PaddingTop(1).PaddingRight(2).PaddingBottom(1).PaddingLeft(2)
s.Padding(1)           // all sides
s.Padding(1, 2)        // top/bottom, left/right
s.Padding(1, 2, 3)     // top, sides, bottom
s.Padding(1, 2, 3, 4)  // clockwise: top, right, bottom, left
s.PaddingChar('·')     // custom fill character
```

**Margin (outside border):**
```go
s.MarginTop(1).MarginRight(2).MarginBottom(1).MarginLeft(2)
s.Margin(2, 4)
s.MarginChar('░')
s.MarginBackground(lipgloss.Color("#111"))
```

---

### Borders

**Predefined border styles:**
```go
lipgloss.NormalBorder()    // ─│┌┐└┘├┤┬┴┼
lipgloss.RoundedBorder()   // ─│╭╮╰╯
lipgloss.ThickBorder()     // ━┃┏┓┗┛
lipgloss.DoubleBorder()    // ═║╔╗╚╝
lipgloss.HiddenBorder()    // invisible (takes up space)
lipgloss.ASCIIBorder()     // +-|
lipgloss.MarkdownBorder()  // |---| markdown table style
```

**Applying borders:**
```go
// Full border
s := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#7D56F4")).
    BorderBackground(lipgloss.Color("#111")).
    Padding(1, 2)

// Selective sides (top, right, bottom, left — clockwise)
s.Border(lipgloss.NormalBorder(), true, false, true, false)  // top + bottom only
s.BorderTop(true).BorderLeft(true)

// Gradient border (v2)
s.BorderForegroundBlend(lipgloss.Color("#FF0000"), lipgloss.Color("#0000FF"))
```

**Custom border:**
```go
custom := lipgloss.Border{
    Top: "~", Bottom: "~", Left: "|", Right: "|",
    TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+",
}
s.Border(custom)
```

---

### Layout Utilities

**Measuring:**
```go
w := lipgloss.Width(renderedBlock)
h := lipgloss.Height(renderedBlock)
w, h := lipgloss.Size(renderedBlock)
// Use these instead of len() — handles ANSI sequences and wide chars correctly
```

**Joining blocks:**
```go
// Horizontal join — align along vertical axis (0.0=top, 0.5=center, 1.0=bottom)
row := lipgloss.JoinHorizontal(lipgloss.Top, blockA, blockB, blockC)
row := lipgloss.JoinHorizontal(0.2, blockA, blockB)  // 20% from top

// Vertical join — align along horizontal axis
col := lipgloss.JoinVertical(lipgloss.Center, blockA, blockB)
col := lipgloss.JoinVertical(lipgloss.Left, blockA, blockB)
```

**Placing in whitespace:**
```go
// Center in 80x24 area
block := lipgloss.Place(80, 24, lipgloss.Center, lipgloss.Center, content)

// Bottom-right corner
block := lipgloss.Place(80, 24, lipgloss.Right, lipgloss.Bottom, content)

// Styled whitespace fill
block := lipgloss.Place(80, 24, lipgloss.Center, lipgloss.Center, content,
    lipgloss.WithWhitespaceStyle(
        lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e")),
    ),
)

// Directional helpers
block := lipgloss.PlaceHorizontal(80, lipgloss.Center, content)
block := lipgloss.PlaceVertical(24, lipgloss.Bottom, content)
```

---

### Compositing (Canvas + Layer)

New in v2: cell-based compositor for overlapping/layered content.

```go
// Create layers with content and position
a := lipgloss.NewLayer(renderedBlockA).X(0).Y(0).Z(0)
b := lipgloss.NewLayer(renderedBlockB).X(20).Y(2).Z(1).ID("overlay")
c := lipgloss.NewLayer(renderedBlockC).X(10).Y(5).Z(2)

// Create a canvas and compose layers
canvas := lipgloss.NewCanvas(80, 24)
canvas.Compose(a).Compose(b).Compose(c)

// Render to string
output := canvas.Render()

// Layer can also hold child layers
parent := lipgloss.NewLayer(bg, childA, childB)
```

`Layer` API:
```go
l.X(x int) *Layer      // x position
l.Y(y int) *Layer      // y position
l.Z(z int) *Layer      // z-index (higher = on top)
l.ID(id string) *Layer // identifier for hit-testing/mouse
l.GetContent() string
l.Width() / l.Height() int
l.AddLayers(layers ...*Layer)
```

`Canvas` API:
```go
canvas.Compose(drawable)   // add a layer
canvas.Clear()
canvas.Resize(w, h)
canvas.Width() / canvas.Height()
canvas.Render() string
```

---

### Sub-packages: table · list · tree

#### lipgloss/table (static rendering)

Not interactive — for rendering styled data tables as strings.

```go
import "charm.land/lipgloss/v2/table"

t := table.New().
    Border(lipgloss.RoundedBorder()).
    BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
    StyleFunc(func(row, col int) lipgloss.Style {
        if row == table.HeaderRow {
            return headerStyle
        }
        if row%2 == 0 { return evenStyle }
        return oddStyle
    }).
    Headers("NAME", "SIZE", "STATUS").
    Rows(
        []string{"Slack", "2.1 GB", "Active"},
        []string{"Xcode", "14.3 GB", "Active"},
    )

t.Row("Figma", "1.2 GB", "Active")  // add row dynamically

lipgloss.Println(t)
```

Border variants: `lipgloss.RoundedBorder()`, `lipgloss.MarkdownBorder()`, `lipgloss.ASCIIBorder()`, `lipgloss.NormalBorder()`, `lipgloss.ThickBorder()`, `lipgloss.DoubleBorder()`

`table.HeaderRow` constant for styling the header row.

#### lipgloss/list (static rendering)

Styled static list rendering with nesting and custom enumerators.

```go
import "charm.land/lipgloss/v2/list"

l := list.New("A", "B", "C")

// Nested
l := list.New(
    "Category A", list.New("item 1", "item 2"),
    "Category B", list.New("item 3", "item 4"),
)

// Custom style and enumerator
l.Enumerator(list.Roman).  // Arabic | Alphabet | Roman | Bullet | Tree
    EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
    ItemStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("212")))

// Custom enumerator function
l.Enumerator(func(items list.Items, i int) string {
    if items.At(i).Value() == "special" { return "★" }
    return "•"
})

// Build incrementally
l := list.New()
for _, item := range items { l.Item(item) }
```

#### lipgloss/tree (static rendering)

Renders file-system-style trees.

```go
import "charm.land/lipgloss/v2/tree"

t := tree.Root(".").
    Child("Applications",
        tree.New().Root("Productivity").
            Child("Notion").Child("Obsidian"),
        tree.New().Root("Development").
            Child("Xcode").Child("VS Code"),
    )

t.Enumerator(tree.RoundedEnumerator).  // DefaultEnumerator | RoundedEnumerator
    EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("63"))).
    RootStyle(lipgloss.NewStyle().Bold(true)).
    ItemStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("212")))
```

---

### Advanced Color Features

#### 1D Gradient (Blend1D)
```go
// Generate N colors blending between stops (CIELAB color space)
colors := lipgloss.Blend1D(80, lipgloss.Color("#FF0000"), lipgloss.Color("#0000FF"))
// Use each color on consecutive characters for gradient text
for i, c := range colors {
    s := lipgloss.NewStyle().Foreground(c)
    fmt.Print(s.Render(string(text[i])))
}
```

#### 2D Gradient (Blend2D)
```go
// Returns row-major slice of colors for a 2D gradient with angle
colors := lipgloss.Blend2D(width, height, 45.0,
    lipgloss.Color("#5A56E0"),
    lipgloss.Color("#EE6FF8"),
    lipgloss.Color("#FF8C00"),
)
for y := range height {
    for x := range width {
        c := colors[y*width+x]
        fmt.Print(lipgloss.NewStyle().Foreground(c).Render("█"))
    }
    fmt.Println()
}
```

---

## 3. Power Features

### bubbles Power Features

**List: Custom ItemDelegate**
Replace the default renderer entirely to create compact single-line items, multi-column layouts, status indicators, progress bars per-item, etc.
```go
type appDelegate struct{}
func (d appDelegate) Height() int  { return 1 }
func (d appDelegate) Spacing() int { return 0 }
func (d appDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d appDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    app := item.(AppItem)
    style := normalStyle
    if index == m.Index() { style = selectedStyle }
    fmt.Fprint(w, style.Render(fmt.Sprintf("%-30s %8s", app.Name, app.Size)))
}
```

**Progress: Dynamic color function**
```go
p := progress.New(
    progress.WithColorFunc(func(total, current float64) color.Color {
        if total < 0.33 { return lipgloss.Color("#22c55e") } // green
        if total < 0.66 { return lipgloss.Color("#f59e0b") } // yellow
        return lipgloss.Color("#ef4444")                      // red
    }),
)
```

**Viewport: Line-level styling + gutter**
```go
vp.StyleLineFunc = func(idx int) lipgloss.Style {
    if isErrorLine(idx) { return errorStyle }
    return normalStyle
}
vp.LeftGutterFunc = func(idx int, afterEnd bool) string {
    return lineNumberStyle.Render(fmt.Sprintf("%4d ", idx+1))
}
```

**Key: Disabling bindings at runtime**
```go
keys.Delete.SetEnabled(!readOnly)
// Disabled keys are skipped by key.Matches and hidden from help view
```

**Multiple spinners/timers without ID collisions**
Each `spinner.Model`, `timer.Model`, `stopwatch.Model` gets a unique internal ID. Route all `TickMsg`s through all models — each rejects ticks that aren't its own.

### lipgloss Power Features

**Style inheritance chains**
Build a style hierarchy: base → section → component → state. Only unset rules are inherited, so you can override at any level without polluting others.

**BorderForegroundBlend — gradient borders**
Pass 2+ colors to `BorderForegroundBlend`; lipgloss interpolates across the border characters.

**Canvas/Layer compositing**
Draw overlapping panels (e.g., modal over list). Assign Z-indices for stacking order. Use Layer IDs for mouse hit detection.

**WithWhitespaceStyle**
Style the empty whitespace around placed content — useful for background fills, "desktop" areas.

**Complete colors**
`lipgloss.Complete(profile)` lets you specify exact colors for ANSI, ANSI256, and TrueColor profiles. The runtime picks the best available, no user input required. Critical for cross-terminal consistency.

**Inline + MaxWidth/MaxHeight**
Enforce component dimensions at render time, preventing layout overflow from long strings.

**Wrap function**
`lipgloss.Wrap(styledText, width, breakpoint)` wraps pre-styled ANSI strings without breaking escape sequences.

---

## 4. UX/UI Patterns for System TUIs

### Layout Architecture

```
┌─ Header Bar ───────────────────────────────────────────┐
│ App title + current mode + stats                       │
├─ Left Panel ──────────────┬─ Right Panel ──────────────┤
│ list.Model (app list)     │ viewport.Model (details)   │
│ + table.Model (columns)   │ + progress bars per metric │
├─ Status Bar ──────────────┴────────────────────────────┤
│ help.View(keys) + status messages                      │
└────────────────────────────────────────────────────────┘
```

Assembly with lipgloss:
```go
leftWidth := termWidth / 3
rightWidth := termWidth - leftWidth - 3  // account for borders

left := lipgloss.NewStyle().Width(leftWidth).Height(contentH).
    Border(lipgloss.RoundedBorder()).Render(m.list.View())

right := lipgloss.NewStyle().Width(rightWidth).Height(contentH).
    Border(lipgloss.RoundedBorder()).Render(m.viewport.View())

body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

view := lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)
```

### Responsive Sizing

Always handle `tea.WindowSizeMsg` to resize all components:
```go
case tea.WindowSizeMsg:
    m.termWidth = msg.Width
    m.termHeight = msg.Height
    m.list.SetSize(leftWidth, contentH)
    m.viewport.SetWidth(rightWidth)
    m.viewport.SetHeight(contentH)
    m.help.SetWidth(msg.Width)
    m.progress.SetWidth(rightWidth - 4)
```

### Modal/Overlay Pattern (Canvas)

```go
// In View:
baseView := m.renderMainLayout()
if m.showConfirmDialog {
    dialog := renderConfirmDialog(m.selectedApp)
    canvas := lipgloss.NewCanvas(m.termWidth, m.termHeight)
    canvas.Compose(lipgloss.NewLayer(baseView).X(0).Y(0).Z(0))
    canvas.Compose(lipgloss.NewLayer(dialog).
        X((m.termWidth-dialogW)/2).
        Y((m.termHeight-dialogH)/2).
        Z(1))
    return canvas.Render()
}
return baseView
```

### Focus Management Pattern

Track focused component with an enum:
```go
type focusState int
const (
    focusList focusState = iota
    focusSearch
    focusDetail
)

// Route Update messages based on focus:
switch m.focus {
case focusList:
    m.list, cmd = m.list.Update(msg)
case focusSearch:
    m.searchInput, cmd = m.searchInput.Update(msg)
}

// Style borders to indicate focus:
activeBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#7D56F4"))
inactiveBorder := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("240"))
```

### Status Indicator Pattern

Use `spinner` + state to show async operations inline:
```go
type opState int
const (opIdle opState = iota; opScanning; opUninstalling)

// In View:
switch m.op {
case opScanning:
    statusLine = m.spinner.View() + " Scanning applications..."
case opUninstalling:
    statusLine = m.spinner.View() + " Uninstalling " + m.targetApp + "..."
    statusLine += "\n" + m.progress.View()
case opIdle:
    statusLine = m.help.View(m.keys)
}
```

### Adaptive Colors Setup

```go
type styles struct {
    selected, normal, header, muted, danger lipgloss.Style
}

func newStyles(isDark bool) styles {
    ld := lipgloss.LightDark(isDark)
    return styles{
        selected: lipgloss.NewStyle().
            Background(ld(lipgloss.Color("63"), lipgloss.Color("57"))).
            Foreground(ld(lipgloss.Color("230"), lipgloss.Color("229"))).
            Bold(true),
        normal: lipgloss.NewStyle().
            Foreground(ld(lipgloss.Color("240"), lipgloss.Color("250"))),
        header: lipgloss.NewStyle().
            Bold(true).
            Foreground(ld(lipgloss.Color("#333333"), lipgloss.Color("#f1f1f1"))),
        muted: lipgloss.NewStyle().
            Faint(true).
            Foreground(ld(lipgloss.Color("244"), lipgloss.Color("243"))),
        danger: lipgloss.NewStyle().
            Foreground(lipgloss.Color("#ef4444")).Bold(true),
    }
}

// In Init + Update:
func (m model) Init() tea.Cmd {
    return tea.RequestBackgroundColor
}
case tea.BackgroundColorMsg:
    m.styles = newStyles(msg.IsDark())
```

### Progress Bar for File Sizes

Visual representation scaled to percentages:
```go
func (m model) renderAppSize(size, maxSize int64) string {
    pct := float64(size) / float64(maxSize)
    p := progress.New(
        progress.WithColors(lipgloss.Color("#5A56E0"), lipgloss.Color("#EE6FF8")),
        progress.WithoutPercentage(),
        progress.WithWidth(20),
    )
    // For static rendering (no animation):
    return p.ViewAs(pct)
}
```

### Key Binding Conventions for System TUIs

```
j/k or ↑/↓   Navigate list
enter         Confirm / enter directory / select
esc           Cancel / back / unfocus
/             Activate filter/search
d or x        Mark for deletion
u             Undo mark
space         Toggle selection (multi-select)
?             Toggle full help view
q or ctrl+c   Quit
tab           Switch panel focus
```

### Performance: Viewport with Large Content

For large log output or app detail panels, set content once and scroll — do not re-render the full string every frame:
```go
// Set once when item changes:
m.viewport.SetContent(buildDetailView(m.selectedApp))

// In View() — just call viewport.View(), no rebuilding:
return m.viewport.View()
```

Use `viewport.FillHeight = true` and `SoftWrap = true` for most detail panes. Disable `SoftWrap` for structured output (logs, package lists) to allow horizontal scrolling.
