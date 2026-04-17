package tui

import (
	"testing"

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
