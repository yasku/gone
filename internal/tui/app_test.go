package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
)

func keyCtrlC() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: 'c', Mod: uv.ModCtrl}
}

func keyChar(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func keyTab() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyTab}
}

func TestAppUpdateSplashDone(t *testing.T) {
	m := NewApp("")
	if !m.showSplash {
		t.Fatal("expected splash visible initially")
	}
	out, _ := m.Update(splashDoneMsg{})
	got := out.(AppModel)
	if got.showSplash {
		t.Fatal("splash should be dismissed after splashDoneMsg")
	}
}

func TestAppUpdateHelpToggle(t *testing.T) {
	m := NewApp("")
	out, _ := m.Update(keyChar('?'))
	m = out.(AppModel)
	if !m.showHelp {
		t.Fatal("'?' should enable help overlay")
	}
	out, _ = m.Update(keyChar('?'))
	m = out.(AppModel)
	if m.showHelp {
		t.Fatal("'?' should toggle help back off")
	}
}

func TestAppUpdateHelpSwallowsOtherKeys(t *testing.T) {
	m := NewApp("")
	m.showSplash = false
	m.showHelp = true
	prev := m.active
	out, _ := m.Update(keyTab())
	m = out.(AppModel)
	if m.active != prev {
		t.Fatal("tab should be swallowed while help overlay is active")
	}
}

func TestAppUpdateTabCycle(t *testing.T) {
	m := NewApp("")
	m.showSplash = false
	if m.active != tabUninstall {
		t.Fatalf("expected starting tab=tabUninstall, got %d", m.active)
	}
	out, _ := m.Update(keyTab())
	m = out.(AppModel)
	if m.active != tabMonitor {
		t.Fatalf("first tab press: got %d, want tabMonitor", m.active)
	}
	out, _ = m.Update(keyTab())
	m = out.(AppModel)
	if m.active != tabUninstall {
		t.Fatalf("second tab press: got %d, want tabUninstall", m.active)
	}
}

func TestAppUpdateCtrlCQuits(t *testing.T) {
	m := NewApp("")
	_, cmd := m.Update(keyCtrlC())
	if cmd == nil {
		t.Fatal("ctrl+c should return a command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c cmd produced %T, want tea.QuitMsg", msg)
	}
}

func TestAppUpdateWindowSize(t *testing.T) {
	m := NewApp("")
	out, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = out.(AppModel)
	if m.width != 120 || m.height != 40 {
		t.Fatalf("window size not applied: %dx%d", m.width, m.height)
	}
	if !m.ready {
		t.Fatal("model should be ready after WindowSizeMsg")
	}
}

func TestAppTabSyncsKeysActive(t *testing.T) {
	m := NewApp("")
	m.showSplash = false
	out, _ := m.Update(keyTab())
	got := out.(AppModel)
	if got.keys.active != tabMonitor {
		t.Errorf("keys.active = %d after tab, want tabMonitor", got.keys.active)
	}
	out, _ = got.Update(keyTab())
	got = out.(AppModel)
	if got.keys.active != tabUninstall {
		t.Errorf("keys.active = %d after second tab, want tabUninstall", got.keys.active)
	}
}

func TestAppShortHelpHasThreeBindings(t *testing.T) {
	m := NewApp("")
	bindings := m.keys.ShortHelp()
	if len(bindings) != 3 {
		t.Errorf("ShortHelp() len = %d, want 3", len(bindings))
	}
}

func TestAppFullHelpUninstallTab(t *testing.T) {
	m := NewApp("")
	m.keys.active = tabUninstall
	groups := m.keys.FullHelp()
	if len(groups) != 2 {
		t.Fatalf("FullHelp() (uninstall) = %d groups, want 2", len(groups))
	}
	if len(groups[0]) != 3 {
		t.Errorf("global group len = %d, want 3", len(groups[0]))
	}
	if len(groups[1]) != 4 {
		t.Errorf("uninstall group len = %d, want 4", len(groups[1]))
	}
}

func TestAppFullHelpMonitorTab(t *testing.T) {
	m := NewApp("")
	m.keys.active = tabMonitor
	groups := m.keys.FullHelp()
	if len(groups) != 2 {
		t.Fatalf("FullHelp() (monitor) = %d groups, want 2", len(groups))
	}
	if len(groups[1]) != 4 {
		t.Errorf("monitor group len = %d, want 4", len(groups[1]))
	}
}

func TestAppRoutesScanMsgToUninstallRegardlessOfTab(t *testing.T) {
	m := NewApp("")
	m.showSplash = false
	m.active = tabMonitor // on monitor tab
	out, _ := m.Update(scanDoneMsg{})
	got := out.(AppModel)
	// scanDoneMsg should update uninstall (scanning=false) even while on monitor tab
	if got.uninstall.scanning {
		t.Error("scanDoneMsg should clear uninstall.scanning even when monitor tab is active")
	}
}

func TestAppRoutesRefreshToMonitor(t *testing.T) {
	m := NewApp("")
	m.showSplash = false
	m.active = tabUninstall
	m.monitor.ready = true
	// refreshMsg should reach monitor even when on uninstall tab — no panic
	out, _ := m.Update(refreshMsg{})
	_ = out
}

func TestAppViewNoPanic(t *testing.T) {
	m := NewApp("")
	out, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = out.(AppModel)
	m.showSplash = false
	v := m.View()
	// View returns a tea.View; just ensure it is non-zero and AltScreen is set
	if !v.AltScreen {
		t.Error("View() should set AltScreen=true")
	}
}
