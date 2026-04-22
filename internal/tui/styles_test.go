package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestGradientTextEmpty(t *testing.T) {
	if got := gradientText(""); got != "" {
		t.Fatalf("gradientText(\"\") = %q, want empty", got)
	}
}

func TestGradientTextRenders(t *testing.T) {
	out := gradientText("GONE")
	if out == "" {
		t.Fatal("gradientText returned empty for non-empty input")
	}
	for _, r := range "GONE" {
		if !strings.ContainsRune(out, r) {
			t.Fatalf("output missing rune %q: %q", r, out)
		}
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatalf("output missing ANSI escape sequence: %q", out)
	}
}

func TestHumanSize(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, c := range cases {
		if got := HumanSize(c.in); got != c.want {
			t.Errorf("HumanSize(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDefaultStylesNonZero(t *testing.T) {
	s := DefaultStyles()
	if s.CursorRow.Render("x") == "" {
		t.Error("CursorRow style rendered empty")
	}
	if s.TabActive.Render("x") == "" {
		t.Error("TabActive style rendered empty")
	}
}

func TestAllStyleFieldsRender(t *testing.T) {
	s := DefaultStyles()
	cases := []struct {
		name  string
		style lipgloss.Style
	}{
		{"SearchBar", s.SearchBar},
		{"SearchBarActive", s.SearchBarActive},
		{"StatusBar", s.StatusBar},
		{"Preview", s.Preview},
		{"Selected", s.Selected},
		{"SelectedItem", s.SelectedItem},
		{"Cursor", s.Cursor},
		{"DimText", s.DimText},
		{"SizeSmall", s.SizeSmall},
		{"SizeMedium", s.SizeMedium},
		{"SizeLarge", s.SizeLarge},
		{"FooterBar", s.FooterBar},
		{"TabBadge", s.TabBadge},
		{"BadgeHigh", s.BadgeHigh},
		{"BadgeMed", s.BadgeMed},
		{"BadgeLow", s.BadgeLow},
	}
	for _, c := range cases {
		if got := c.style.Render("x"); got == "" {
			t.Errorf("style %s Render() returned empty", c.name)
		}
	}
}

func TestTabBadgeRendersWithContent(t *testing.T) {
	s := DefaultStyles()
	out := s.TabBadge.Render("42")
	if !strings.Contains(out, "42") {
		t.Errorf("TabBadge output missing content '42': %q", out)
	}
}

func TestFooterBarRendersWithContent(t *testing.T) {
	s := DefaultStyles()
	out := s.FooterBar.Render("tab  ?  ctrl+c")
	if !strings.Contains(out, "tab") {
		t.Errorf("FooterBar output missing content: %q", out)
	}
}

func TestSelectedItemRendersWithContent(t *testing.T) {
	s := DefaultStyles()
	out := s.SelectedItem.Render("/usr/local/bin/node")
	if !strings.Contains(out, "node") {
		t.Errorf("SelectedItem output missing content: %q", out)
	}
}
