package tui

import (
	"testing"
	"time"

	"charm.land/bubbletea/v2"
)

func TestAuditModelSetSize(t *testing.T) {
	m := NewAuditModel()
	m = m.SetSize(120, 40)
	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestAuditModelInit(t *testing.T) {
	m := NewAuditModel()
	m.available = false
	m.ready = true
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init should return a command even when osquery unavailable")
	}
}

func TestAuditModelUpdateRefresh(t *testing.T) {
	m := NewAuditModel()
	m.ready = true
	m.available = false

	_, cmd := m.Update(auditRefreshMsg(time.Now()))
	if cmd == nil {
		t.Fatal("auditRefreshMsg should return a refresh command")
	}
}

func TestAuditModelRefresh(t *testing.T) {
	m := NewAuditModel()
	m.ready = true
	m.available = false
	m.loading = false

	_, _ = m.Update(tea.KeyPressMsg{Text: "r"})

	if m.loading {
		t.Error("loading should be false after refresh with osquery unavailable")
	}
}

func TestAuditModelViewNotReady(t *testing.T) {
	m := NewAuditModel()
	v := m.View()
	if v.Content == "" {
		t.Error("View should return non-empty Content even when not ready")
	}
}

func TestAuditModelViewReadyNotAvailable(t *testing.T) {
	m := NewAuditModel()
	m.ready = true
	m.available = false
	v := m.View()
	if v.Content == "" {
		t.Error("View should return non-empty Content when ready but osquery unavailable")
	}
}

func TestAuditModelViewReadyAvailable(t *testing.T) {
	m := NewAuditModel()
	m.ready = true
	m.available = true
	m.loading = false
	m.categories = []AuditCategory{
		{Name: "Apps", Count: 10, Status: "ok"},
	}
	v := m.View()
	if v.Content == "" {
		t.Error("View should return non-empty Content when ready with osquery")
	}
}

func TestAuditModelFormatDetails(t *testing.T) {
	m := NewAuditModel()
	m.ready = true

	details := []map[string]string{
		{"name": "TestApp", "path": "/Applications/TestApp.app"},
		{"name": "TestPlugin", "path": "/Library/Internet Plug-Ins/Test.plugin"},
	}

	v := m.formatDetails(details)
	if v == "" {
		t.Error("formatDetails should return non-empty string")
	}
}

func TestAuditModelFormatDetailsEmpty(t *testing.T) {
	m := NewAuditModel()
	m.ready = true

	v := m.formatDetails([]map[string]string{})
	if v != "" {
		t.Error("formatDetails should return empty string for empty details")
	}
}
