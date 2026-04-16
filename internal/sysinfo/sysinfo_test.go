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
