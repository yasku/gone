package tui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"gone/internal/sysinfo"
)

func fakeProcs() []sysinfo.ProcInfo {
	return []sysinfo.ProcInfo{
		{PID: 300, Name: "c", CPU: 10.0, Mem: 5.0, RSS: 3000},
		{PID: 100, Name: "a", CPU: 90.0, Mem: 1.0, RSS: 5000},
		{PID: 200, Name: "b", CPU: 50.0, Mem: 9.0, RSS: 1000},
	}
}

func TestSortedProcsByCPU(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.sortBy = sortCPU
	got := m.sortedProcs()
	if got[0].PID != 100 || got[1].PID != 200 || got[2].PID != 300 {
		t.Fatalf("sortCPU order wrong: got %v", pidOrder(got))
	}
}

func TestSortedProcsByMem(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.sortBy = sortMem
	got := m.sortedProcs()
	if got[0].PID != 200 || got[1].PID != 300 || got[2].PID != 100 {
		t.Fatalf("sortMem order wrong: got %v", pidOrder(got))
	}
}

func TestSortedProcsByRSS(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.sortBy = sortRSS
	got := m.sortedProcs()
	if got[0].PID != 100 || got[1].PID != 300 || got[2].PID != 200 {
		t.Fatalf("sortRSS order wrong: got %v", pidOrder(got))
	}
}

func TestSortedProcsByPID(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.sortBy = sortPID
	got := m.sortedProcs()
	if got[0].PID != 100 || got[1].PID != 200 || got[2].PID != 300 {
		t.Fatalf("sortPID order wrong: got %v", pidOrder(got))
	}
}

func TestSortedProcsDoesNotMutate(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.sortBy = sortCPU
	_ = m.sortedProcs()
	if m.snapshot.Procs[0].PID != 300 {
		t.Fatal("sortedProcs mutated original snapshot")
	}
}

func TestNewGaugeBar(t *testing.T) {
	bar := newGaugeBar()
	_ = bar.SetPercent(0.5)
	if bar.View() == "" {
		t.Fatal("gauge bar rendered empty")
	}
}

func TestTruncateName(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"short", 25, "short"},
		{"exactly_twenty_five_chars", 25, "exactly_twenty_five_chars"},
		{"this_name_is_way_too_long_to_fit", 10, "this_name…"},
	}
	for _, c := range cases {
		if got := truncateName(c.in, c.max); got != c.want {
			t.Errorf("truncateName(%q, %d) = %q, want %q", c.in, c.max, got, c.want)
		}
	}
}

func pidOrder(ps []sysinfo.ProcInfo) []int32 {
	out := make([]int32, len(ps))
	for i, p := range ps {
		out[i] = p.PID
	}
	return out
}

func TestMonitorCursorDown(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.ready = true
	out, _ := m.Update(keyChar('j'))
	got := out
	if got.cursor != 1 {
		t.Errorf("cursor after j: want 1, got %d", got.cursor)
	}
}

func TestMonitorCursorUpNoWrap(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.ready = true
	m.cursor = 0
	out, _ := m.Update(keyChar('k'))
	got := out
	if got.cursor != 0 {
		t.Errorf("cursor should stay 0 at top, got %d", got.cursor)
	}
}

func TestMonitorCursorBounded(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.ready = true
	m.cursor = 2 // last index in fakeProcs (3 items)
	out, _ := m.Update(keyChar('j'))
	got := out
	if got.cursor != 2 {
		t.Errorf("cursor should stay at max, got %d", got.cursor)
	}
}

func TestMonitorSortKeyBindings(t *testing.T) {
	cases := []struct {
		key  rune
		want sortCol
	}{
		{'1', sortCPU},
		{'2', sortMem},
		{'3', sortRSS},
		{'4', sortPID},
	}
	for _, c := range cases {
		m := NewMonitorModel()
		out, _ := m.Update(keyChar(c.key))
		got := out
		if got.sortBy != c.want {
			t.Errorf("key %q: sortBy = %d, want %d", string(c.key), got.sortBy, c.want)
		}
	}
}

func TestMonitorKillPendingOnX(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.ready = true
	out, _ := m.Update(keyChar('x'))
	got := out
	if !got.killPending {
		t.Error("x key should set killPending")
	}
	if got.killTarget.PID == 0 {
		t.Error("killTarget should be set")
	}
}

func TestMonitorKillPendingEscCancels(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.killPending = true
	m.killTarget = fakeProcs()[0]
	out, _ := m.Update(tea.KeyPressMsg{Text: "esc", Code: tea.KeyEscape})
	got := out
	if got.killPending {
		t.Error("esc should clear killPending")
	}
}

func TestMonitorKillPendingEnterEmitsCmd(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.killPending = true
	m.killTarget = fakeProcs()[0]
	_, cmd := m.Update(tea.KeyPressMsg{Text: "enter", Code: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter while killPending should emit kill command")
	}
}

func TestMonitorFilterSlashEnablesFilter(t *testing.T) {
	m := NewMonitorModel()
	out, _ := m.Update(keyChar('/'))
	got := out
	if !got.filtering {
		t.Error("/ key should enable filtering mode")
	}
}

func TestMonitorFilterEscDisablesFilter(t *testing.T) {
	m := NewMonitorModel()
	m.filtering = true
	cmd := m.filterInput.Focus()
	_ = cmd
	out, _ := m.Update(tea.KeyPressMsg{Text: "esc", Code: tea.KeyEscape})
	got := out
	if got.filtering {
		t.Error("esc should disable filtering mode")
	}
	if got.filterInput.Value() != "" {
		t.Error("filter input should be cleared on esc")
	}
}

func TestSortedProcsWithFilter(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs() // names: "a", "b", "c"
	m.filtering = true
	m.filterInput.SetValue("a")

	got := m.sortedProcs()
	if len(got) == 0 {
		t.Fatal("filter 'a' should return at least 1 proc")
	}
	found := false
	for _, p := range got {
		if p.Name == "a" {
			found = true
		}
	}
	if !found {
		t.Errorf("proc 'a' should be in fuzzy results for term 'a'")
	}
}

func TestSortedProcsWithFuzzyFilter(t *testing.T) {
	m := NewMonitorModel()
	// Use names that allow fuzzy matching: "claude" should match "clde" (substring fuzzy)
	m.snapshot.Procs = []sysinfo.ProcInfo{
		{PID: 1, Name: "claude", CPU: 1.0},
		{PID: 2, Name: "node", CPU: 1.0},
		{PID: 3, Name: "python3", CPU: 1.0},
	}
	m.filtering = true
	m.filterInput.SetValue("claude")

	got := m.sortedProcs()
	if len(got) == 0 {
		t.Fatal("fuzzy filter 'claude' should match 'claude'")
	}
	if got[0].Name != "claude" {
		t.Errorf("top fuzzy result for 'claude' should be 'claude', got %q", got[0].Name)
	}
}

func TestSortedProcsEmptyFilterReturnsAll(t *testing.T) {
	m := NewMonitorModel()
	m.snapshot.Procs = fakeProcs()
	m.filtering = true
	m.filterInput.SetValue("") // empty filter = no filtering

	got := m.sortedProcs()
	if len(got) != len(fakeProcs()) {
		t.Errorf("empty filter should return all procs, got %d", len(got))
	}
}

func TestMonitorProcCount(t *testing.T) {
	m := NewMonitorModel()
	if m.ProcCount() != 0 {
		t.Errorf("fresh model: expected 0, got %d", m.ProcCount())
	}
	m.snapshot.Procs = fakeProcs()
	if m.ProcCount() != 3 {
		t.Errorf("expected 3, got %d", m.ProcCount())
	}
}

func TestMonitorKillDoneMsgSetsError(t *testing.T) {
	m := NewMonitorModel()
	m, _ = m.Update(killDoneMsg{pid: 999, err: fmt.Errorf("permission denied")})
	if m.killErr == "" {
		t.Error("killErr should be set on kill failure")
	}
}

func TestMonitorKillDoneMsgClearsError(t *testing.T) {
	m := NewMonitorModel()
	m.killErr = "previous error"
	out, _ := m.Update(killDoneMsg{pid: 999, err: nil})
	got := out
	if got.killErr != "" {
		t.Errorf("killErr should be cleared on success, got %q", got.killErr)
	}
}

func TestMonitorSetSize(t *testing.T) {
	m := NewMonitorModel()
	m = m.SetSize(200, 50)
	if m.width != 200 || m.height != 50 {
		t.Errorf("SetSize not applied: %dx%d", m.width, m.height)
	}
}

func TestMonitorViewNoPanic(t *testing.T) {
	m := NewMonitorModel()
	m.ready = true
	m.snapshot.Procs = fakeProcs()
	m = m.SetSize(120, 40)
	out := m.View()
	if out.Content == "" {
		t.Error("View() returned empty Content")
	}
}
