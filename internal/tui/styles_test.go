package tui

import (
	"strings"
	"testing"
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
