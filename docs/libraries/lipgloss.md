# Lipgloss v2 (charm.land/lipgloss/v2 v2.0.3)

## NewStyle() Chain
```go
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#00BCD4")).
    Padding(0, 1).
    Border(lipgloss.RoundedBorder())
```

## Color Types
```go
// Hex (24-bit true color)
lipgloss.Color("#00BCD4")   // cyan
lipgloss.Color("#9B59B6")   // purple

// ANSI 256
lipgloss.Color("240")       // dark gray
lipgloss.Color("252")       // light gray

// ANSI 16
lipgloss.Color("5")         // magenta
lipgloss.Color("9")         // red

// Named constants
lipgloss.Black
lipgloss.Red
lipgloss.Green
lipgloss.Yellow
lipgloss.Blue
lipgloss.Magenta
lipgloss.Cyan
lipgloss.White
```

## Color Utilities
```go
c := lipgloss.Color("#EB4268")
dark := lipgloss.Darken(c, 0.5)      // 50% darker
light := lipgloss.Lighten(c, 0.35)   // 35% lighter
complement := lipgloss.Complementary(c)
withAlpha := lipgloss.Alpha(c, 0.2)
```

## Blend/Gradient
```go
// 1D blend for gradient text
colors := lipgloss.Blend1D(len(runes), lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4"))
for i, r := range runes {
    sb.WriteString(lipgloss.NewStyle().Foreground(colors[i]).Render(string(r)))
}
```

## Borders
```go
lipgloss.RoundedBorder()
lipgloss.NormalBorder()
lipgloss.HiddenBorder()

// Border foreground
lipgloss.BorderForeground(lipgloss.Color("240"))

// Gradient border (2 colors)
lipgloss.BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4"))
```

## Layout Helpers

### Join Horizontal
```go
lipgloss.JoinHorizontal(lipgloss.Center, elem1, elem2, elem3)
```

### Join Vertical
```go
lipgloss.JoinVertical(lipgloss.Center, elem1, elem2)
```

### Place (Center Content)
```go
lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
```

### Available Alignments
- `lipgloss.Center`
- `lipgloss.Left`
- `lipgloss.Right`
- `lipgloss.Top`
- `lipgloss.Bottom`

## CENTERING BLOCKS (Critical Pattern)
When you need to center multiple elements as a single block (e.g., 4 gauges centered together), use `lipgloss.Place`:

```go
// ❌ WRONG - gauges are left-aligned within their container
gauges := lipgloss.JoinHorizontal(lipgloss.Center, gauge1, gauge2, gauge3, gauge4)
b.WriteString(gauges)

// ✅ CORRECT - the entire block is centered on the screen
gauges := lipgloss.JoinHorizontal(lipgloss.Center, gauge1, gauge2, gauge3, gauge4)
b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Left, gauges))
```

**Syntax:** `lipgloss.Place(width, height, horizontalAlign, verticalAlign, content)`
- `width, height` — dimensions of the container area
- `horizontalAlign` — where to place content horizontally (Center/Left/Right)
- `verticalAlign` — where to place content vertically (Top/Center/Bottom)
- `content` — the content to place

**Rule:** When joining multiple elements and they need to be centered as a group, ALWAYS wrap with `lipgloss.Place(m.width, ..., lipgloss.Center, ...)`.

## Block Formatting
```go
// Padding
lipgloss.NewStyle().Padding(2)           // all sides
lipgloss.NewStyle().Padding(1, 4)        // vertical, horizontal
lipgloss.NewStyle().Padding(1, 4, 2)    // top, horizontal, bottom
lipgloss.NewStyle().Padding(1, 4, 2, 3) // top, right, bottom, left

// Margin (outside border)
lipgloss.NewStyle().Margin(2)

// Width and Height
lipgloss.NewStyle().Width(50).Height(20)
```

## Width Calculation
```go
lipgloss.Width(style.Render(content))  // get rendered width
lipgloss.Height(style.Render(content)) // get rendered height
```

## Inline Formatting
```go
lipgloss.NewStyle().
    Bold(true).
    Italic(true).
    Faint(true).
    Blink(true).
    Strikethrough(true).
    Underline(true).
    Reverse(true)
```

## Adaptive Colors (Light/Dark)
```go
import "charm.land/lipgloss/v2/compat"

// v1 style
color := lipgloss.AdaptiveColor{Light: "#f1f1f1", Dark: "#cccccc"}

// v2 style
color := compat.AdaptiveColor{
    Light: lipgloss.Color("#f1f1f1"),
    Dark: lipgloss.Color("#cccccc"),
}
```

## Render Returns String
```go
rendered := style.Render("text")  // returns string
```

## Style Composition
```go
// Compose styles
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#00BCD4")).
    Padding(0, 1)

lipgloss.Println(style.Render("text"))  // print directly
```

## Styles Struct Pattern (gone uses this)
Centralize styles to avoid duplication:
```go
type Styles struct {
    App         lipgloss.Style
    TabActive   lipgloss.Style
    SearchBar   lipgloss.Style
    CursorRow   lipgloss.Style
    DimText     lipgloss.Style
}

func DefaultStyles() Styles {
    return Styles{
        CursorRow: lipgloss.NewStyle().
            Background(lipgloss.Color("#1A1A2E")).
            Foreground(lipgloss.Color("#00BCD4")).
            Bold(true),
        ...
    }
}
```

## tea.NewView() with Lipgloss
```go
func (m model) View() tea.View {
    content := lipgloss.JoinVertical(
        lipgloss.Center,
        titleStyle.Render("Title"),
        bodyStyle.Render("Body"),
    )
    return tea.NewView(content)
}

// Or with border
box := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
    Padding(1, 3).
    Width(50).
    Render(content)
```

## Common Mistakes
1. `Color` is a function returning `color.Color`, not a string
2. `Render()` returns `string`, not a styled type
3. Use `charm.land/lipgloss/v2`, NOT `github.com/charmbracelet/lipgloss`

## References
- https://pkg.go.dev/charm.land/lipgloss/v2
- https://github.com/charmbracelet/lipgloss