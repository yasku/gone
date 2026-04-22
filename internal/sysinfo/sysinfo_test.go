package sysinfo_test

import (
	"testing"

	"gone/internal/sysinfo"
)

func TestTakeSnapshotReturnsData(t *testing.T) {
	s := sysinfo.TakeSnapshot(5)
	if s.MemTotal == 0 {
		t.Error("expected non-zero MemTotal")
	}
	if s.DiskTotal == 0 {
		t.Error("expected non-zero DiskTotal")
	}
	if len(s.Procs) == 0 {
		t.Error("expected at least one process")
	}
}

func TestTakeSnapshotCPUBounded(t *testing.T) {
	s := sysinfo.TakeSnapshot(5)
	if s.CPUPercent < 0 || s.CPUPercent > 100 {
		t.Errorf("CPUPercent out of bounds: %.2f", s.CPUPercent)
	}
}

func TestTakeSnapshotMemUsedLessOrEqualTotal(t *testing.T) {
	s := sysinfo.TakeSnapshot(5)
	if s.MemUsed > s.MemTotal {
		t.Errorf("MemUsed (%d) > MemTotal (%d)", s.MemUsed, s.MemTotal)
	}
}

func TestTakeSnapshotProcCountCapped(t *testing.T) {
	const cap = 3
	s := sysinfo.TakeSnapshot(cap)
	if len(s.Procs) > cap {
		t.Errorf("expected ≤%d procs, got %d", cap, len(s.Procs))
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}
	for _, c := range cases {
		got := sysinfo.HumanBytes(c.in)
		if got != c.want {
			t.Errorf("HumanBytes(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
