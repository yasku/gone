package tui

// Styles holds all Lip Gloss styles used across the gone TUI. A single
// DefaultStyles() call is made at model construction time and the result is
// stored on each model that needs rendering styles.

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
	SelectedItem    lipgloss.Style
	Cursor          lipgloss.Style
	DimText         lipgloss.Style
	SizeSmall       lipgloss.Style
	SizeMedium      lipgloss.Style
	SizeLarge       lipgloss.Style
	CursorRow       lipgloss.Style
	FooterBar       lipgloss.Style
	TabBadge        lipgloss.Style
	BadgeHigh       lipgloss.Style
	BadgeMed        lipgloss.Style
	BadgeLow        lipgloss.Style
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
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 1).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")),
		Preview:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
		Selected:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true),
		SelectedItem: lipgloss.NewStyle().Background(lipgloss.Color("#2D1B3D")).Foreground(lipgloss.Color("#E8D5F2")).Bold(true),
		Cursor:       lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		DimText:      lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		SizeSmall:    lipgloss.NewStyle().Foreground(lipgloss.Color("120")),
		SizeMedium:   lipgloss.NewStyle().Foreground(lipgloss.Color("178")),
		SizeLarge:    lipgloss.NewStyle().Foreground(lipgloss.Color("167")),
		CursorRow: lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1),
		FooterBar: lipgloss.NewStyle().
			Background(lipgloss.Color("234")).
			Foreground(lipgloss.Color("241")).
			Padding(0, 2),
		TabBadge: lipgloss.NewStyle().
			Background(lipgloss.Color("#00BCD4")).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1),
		BadgeHigh: lipgloss.NewStyle().
			Background(lipgloss.Color("#FF6B6B")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1),
		BadgeMed: lipgloss.NewStyle().
			Background(lipgloss.Color("#FFDD57")).
			Foreground(lipgloss.Color("#333333")).
			Padding(0, 1),
		BadgeLow: lipgloss.NewStyle().
			Background(lipgloss.Color("#69FF94")).
			Foreground(lipgloss.Color("#333333")).
			Padding(0, 1),
	}
}

// gradientText renders each rune of text in a smooth colour gradient from
// #9B59B6 (purple) to #00BCD4 (cyan) using Lip Gloss Blend1D.
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

// HumanSize converts b bytes to a human-readable string (e.g. "1.5 KB").
// Uses 1024-based units. Unlike sysinfo.HumanBytes, the input is int64 to
// accommodate file sizes reported by os.FileInfo.
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
