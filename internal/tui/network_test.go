package tui

import (
	"testing"
	"time"

	"gone/internal/sysinfo"
)

func TestNetworkModelSetSize(t *testing.T) {
	m := NewNetworkModel()
	m = m.SetSize(120, 40)
	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestNetworkModelUpdateRefresh(t *testing.T) {
	m := NewNetworkModel()
	m.ready = true
	m.ifaces = []sysinfo.NetInterface{}

	_, cmd := m.Update(networkRefreshMsg(time.Now()))
	if cmd == nil {
		t.Fatal("networkRefreshMsg should return a refresh command")
	}
}

func TestNetworkModelViewNotReady(t *testing.T) {
	m := NewNetworkModel()
	v := m.View()
	if v.Content == "" {
		t.Error("View should return non-empty Content even when not ready")
	}
}

func TestNetworkModelViewReady(t *testing.T) {
	m := NewNetworkModel()
	m.ready = true
	m.ifaces = []sysinfo.NetInterface{}
	v := m.View()
	if v.Content == "" {
		t.Error("View should return non-empty Content when ready")
	}
}
