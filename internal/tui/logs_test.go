package tui

import (
	"testing"
)

func TestLogsModelSetSize(t *testing.T) {
	m := NewLogsModel()
	m = m.SetSize(120, 40)
	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestLogsModelReady(t *testing.T) {
	m := NewLogsModel()
	m.ready = true
	if !m.ready {
		t.Error("ready should be true when set directly")
	}
}

func TestLogsModelViewNotReady(t *testing.T) {
	m := NewLogsModel()
	v := m.View()
	if v.Content == "" {
		t.Error("View should return non-empty Content even when not ready")
	}
}

func TestLogsModelViewReady(t *testing.T) {
	m := NewLogsModel()
	m.ready = true
	m.entries = []LogEntry{}
	v := m.View()
	if v.Content == "" {
		t.Error("View should return non-empty Content when ready")
	}
}

func TestLogsModelFormatEntry(t *testing.T) {
	m := NewLogsModel()
	m.ready = true
	m.width = 120

	entry := LogEntry{
		Timestamp: "2026-04-27T10:00:00Z",
		Operation: "TRASH",
		Path:      "/Users/test/app",
		Size:      1024,
		Kind:      "directory",
	}

	v := m.formatEntry(entry)
	if v == "" {
		t.Error("formatEntry should return non-empty string")
	}
}

func TestLogsModelFormatEntryUnknownOp(t *testing.T) {
	m := NewLogsModel()
	m.ready = true
	m.width = 120

	entry := LogEntry{
		Timestamp: "2026-04-27T10:00:00Z",
		Operation: "UNKNOWN",
		Path:      "/Users/test/app",
		Size:      1024,
		Kind:      "file",
	}

	v := m.formatEntry(entry)
	if v == "" {
		t.Error("formatEntry should return non-empty string for unknown operation")
	}
}

func TestLogsModelFormatLogsEmpty(t *testing.T) {
	m := NewLogsModel()
	m.ready = true
	m.entries = []LogEntry{}

	v := m.formatLogs()
	if v != "" {
		t.Errorf("formatLogs for empty entries = %q, want empty string", v)
	}
}
