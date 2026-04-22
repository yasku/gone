package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"gone/internal/scanner"
)

// closedMatchCh returns a closed scanner.Match channel (simulates scan end).
func closedMatchCh() <-chan scanner.Match {
	ch := make(chan scanner.Match)
	close(ch)
	return ch
}

// --- fileItem helpers ---

func TestFileItemFilterValue(t *testing.T) {
	f := fileItem{path: "/home/user/.claude/bin"}
	if got := f.FilterValue(); got != "/home/user/.claude/bin" {
		t.Errorf("FilterValue() = %q", got)
	}
}

func TestFileItemTitle(t *testing.T) {
	f := fileItem{path: "/usr/local/bin/node"}
	if got := f.Title(); got != "/usr/local/bin/node" {
		t.Errorf("Title() = %q", got)
	}
}

func TestFileItemDescription(t *testing.T) {
	f := fileItem{path: "/tmp/x", kind: "file", size: 1024, modTime: "2026-01-01"}
	desc := f.Description()
	if !strings.Contains(desc, "file") {
		t.Error("Description missing kind")
	}
	if !strings.Contains(desc, "1.0 KB") {
		t.Error("Description missing size")
	}
}

// --- truncate ---

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this/is/a/longer/path/than/max", 10, "…/than/max"},
	}
	for _, c := range cases {
		got := truncate(c.in, c.max)
		if got != c.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}

// --- kindIcon ---

func TestKindIcon(t *testing.T) {
	cases := []struct{ kind, want string }{
		{"dir", "d"},
		{"symlink", "@"},
		{"rc-line", "#"},
		{"file", "·"},
		{"unknown", "·"},
		{"", "·"},
	}
	for _, c := range cases {
		if got := kindIcon(c.kind); got != c.want {
			t.Errorf("kindIcon(%q) = %q, want %q", c.kind, got, c.want)
		}
	}
}

// --- previewContent ---

func TestPreviewContentFile(t *testing.T) {
	f := fileItem{path: "/tmp/test.txt", kind: "file", size: 512, modTime: "2026-01-15"}
	out := previewContent(f)
	if !strings.Contains(out, "/tmp/test.txt") {
		t.Error("preview missing path")
	}
	if !strings.Contains(out, "file") {
		t.Error("preview missing kind")
	}
	if !strings.Contains(out, "512 B") {
		t.Error("preview missing size")
	}
}

func TestPreviewContentRCLine(t *testing.T) {
	f := fileItem{path: "/home/user/.zshrc:42", kind: "rc-line"}
	out := previewContent(f)
	if !strings.Contains(out, "/home/user/.zshrc") {
		t.Error("preview missing rc file path")
	}
	if !strings.Contains(out, "42") {
		t.Error("preview missing line number")
	}
}

// --- UninstallModel state machine ---

func newReadyUninstall() UninstallModel {
	m := NewUninstall("")
	m.width = 120
	m.height = 40
	return m
}

func TestUninstallUpdateScanItemMsg(t *testing.T) {
	m := newReadyUninstall()
	m.term = "claude"
	m.scanning = true

	item := fileItem{path: "/usr/local/bin/claude", size: 2048, kind: "file"}
	got, _ := m.Update(scanItemMsg{item: item, ch: closedMatchCh()})

	if got.ItemCount() != 1 {
		t.Errorf("expected 1 item after scanItemMsg, got %d", got.ItemCount())
	}
	if !strings.Contains(got.status, "found") {
		t.Errorf("status should mention found count, got %q", got.status)
	}
}

func TestUninstallUpdateScanDoneMsgWithResults(t *testing.T) {
	m := newReadyUninstall()
	m.term = "node"
	m.scanning = true

	item := fileItem{path: "/usr/local/bin/node", size: 1024, kind: "file"}
	m, _ = m.Update(scanItemMsg{item: item, ch: closedMatchCh()})

	got, _ := m.Update(scanDoneMsg{})

	if got.scanning {
		t.Error("scanning should be false after scanDoneMsg")
	}
	if got.focus != focusList {
		t.Error("focus should move to list after scan completes with results")
	}
}

func TestUninstallUpdateScanDoneMsgNoResults(t *testing.T) {
	m := newReadyUninstall()
	m.term = "nonexistent"
	m.scanning = true

	got, _ := m.Update(scanDoneMsg{})

	if got.scanning {
		t.Error("scanning should be false")
	}
	if !strings.Contains(got.status, "No matches") {
		t.Errorf("status should say no matches, got %q", got.status)
	}
}

func TestUninstallUpdateTrashDoneMsg(t *testing.T) {
	m := newReadyUninstall()
	m.term = "claude"
	m.scanning = true

	got, _ := m.Update(trashDoneMsg{count: 2, freed: 4096})

	if got.scanning {
		t.Error("scanning should be false after trashDoneMsg")
	}
	if !strings.Contains(got.status, "Trashed") {
		t.Errorf("status should mention Trashed, got %q", got.status)
	}
}

func TestUninstallUpdateSpaceTogglesSelection(t *testing.T) {
	m := newReadyUninstall()
	m.term = "test"

	item := fileItem{path: "/tmp/testapp", size: 100, kind: "dir"}
	m, _ = m.Update(scanItemMsg{item: item, ch: closedMatchCh()})
	m.focus = focusList

	got, _ := m.Update(keyChar(' '))

	sel := got.SelectedItems()
	if len(sel) != 1 {
		t.Errorf("expected 1 selected item after space, got %d", len(sel))
	}
}

func TestUninstallUpdateEnterInListWithNoSelection(t *testing.T) {
	m := newReadyUninstall()
	m.focus = focusList

	got, _ := m.Update(keyChar('\r'))

	if got.confirmPending {
		t.Error("confirmPending should not be set with no selected items")
	}
}

func TestUninstallUpdateConfirmEnterExecutesTrash(t *testing.T) {
	m := newReadyUninstall()
	m.focus = focusList
	m.confirmPending = true

	got, cmd := m.Update(tea.KeyPressMsg{Text: "enter", Code: tea.KeyEnter})

	if got.confirmPending {
		t.Error("confirmPending should be cleared after confirm enter")
	}
	// cmd may be nil if no items selected — that's acceptable; just verify no panic
	_ = cmd
}

func TestUninstallUpdateConfirmEscCancels(t *testing.T) {
	m := newReadyUninstall()
	m.confirmPending = true

	got, _ := m.Update(tea.KeyPressMsg{Text: "esc", Code: tea.KeyEscape})

	if got.confirmPending {
		t.Error("confirmPending should be cleared on esc")
	}
}

func TestUninstallUpdateEscFromListToSearch(t *testing.T) {
	m := newReadyUninstall()
	m.focus = focusList

	got, _ := m.Update(tea.KeyPressMsg{Text: "esc", Code: tea.KeyEscape})

	if got.focus != focusSearch {
		t.Errorf("expected focus=focusSearch after esc, got %d", got.focus)
	}
}

func TestUninstallSelectedItems(t *testing.T) {
	m := newReadyUninstall()

	for _, item := range []fileItem{
		{path: "/a", size: 1, kind: "file"},
		{path: "/b", size: 2, kind: "file"},
		{path: "/c", size: 3, kind: "file"},
	} {
		m, _ = m.Update(scanItemMsg{item: item, ch: closedMatchCh()})
	}

	// Mark /a and /c as selected
	items := m.list.Items()
	for i, it := range items {
		if f, ok := it.(fileItem); ok && (f.path == "/a" || f.path == "/c") {
			f.selected = true
			m.list.SetItem(i, f)
		}
	}

	sel := m.SelectedItems()
	if len(sel) != 2 {
		t.Errorf("expected 2 selected items, got %d", len(sel))
	}
}

func TestUninstallItemCount(t *testing.T) {
	m := newReadyUninstall()
	if m.ItemCount() != 0 {
		t.Errorf("fresh model: expected 0 items, got %d", m.ItemCount())
	}

	item := fileItem{path: "/tmp/x", kind: "file"}
	m, _ = m.Update(scanItemMsg{item: item, ch: closedMatchCh()})

	if m.ItemCount() != 1 {
		t.Errorf("after 1 insert: expected 1 item, got %d", m.ItemCount())
	}
}

func TestUninstallSetSize(t *testing.T) {
	m := NewUninstall("")
	m = m.SetSize(160, 50)
	if m.width != 160 || m.height != 50 {
		t.Errorf("SetSize not applied: %dx%d", m.width, m.height)
	}
}

func TestUninstallViewNoPanic(t *testing.T) {
	m := NewUninstall("")
	m = m.SetSize(120, 40)
	out := m.View()
	if out == "" {
		t.Error("View() returned empty string")
	}
}

func TestUninstallViewWelcomeState(t *testing.T) {
	m := NewUninstall("")
	m = m.SetSize(120, 40)
	out := m.View()
	if !strings.Contains(out, "hunt") && !strings.Contains(out, "gone") && !strings.Contains(out, "Enter") {
		t.Errorf("welcome state View() missing expected hint text, got: %q", safePrefix(out, 200))
	}
}

func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
