package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type Styles struct {
	App             lipgloss.Style
	TabActive       lipgloss.Style
	TabInactive     lipgloss.Style
	SearchBar       lipgloss.Style
	SearchBarActive lipgloss.Style
	StatusBar       lipgloss.Style
	Preview         lipgloss.Style
	Selected        lipgloss.Style
	Cursor          lipgloss.Style
	DimText         lipgloss.Style
	SizeSmall       lipgloss.Style
	SizeMedium      lipgloss.Style
	SizeLarge       lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		App:         lipgloss.NewStyle().Padding(1, 2),
		TabActive:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Padding(0, 2),
		TabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 2),
		SearchBar:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("245")).Padding(0, 1),
		SearchBarActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
			Padding(0, 1),
		StatusBar:   lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252")).Padding(0, 1),
		Preview:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
		Cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		DimText:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		SizeSmall:   lipgloss.NewStyle().Foreground(lipgloss.Color("120")),
		SizeMedium:  lipgloss.NewStyle().Foreground(lipgloss.Color("178")),
		SizeLarge:   lipgloss.NewStyle().Foreground(lipgloss.Color("167")),
	}
}

// gradientText renders each rune of text in a gradient from #9B59B6 (purple) to #00BCD4 (cyan).
func gradientText(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	colors := lipgloss.Blend1D(len(runes), lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4"))
	var sb strings.Builder
	for i, r := range runes {
		sb.WriteString(lipgloss.NewStyle().Foreground(colors[i]).Render(string(r)))
	}
	return sb.String()
}

func HumanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
