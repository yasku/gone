package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSplashInitReturnsCmd(t *testing.T) {
	m := NewSplashModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("SplashModel.Init() should return a non-nil command")
	}
}

func TestSplashDoneMsgSetsDone(t *testing.T) {
	m := NewSplashModel()
	got, _ := m.Update(splashDoneMsg{})
	if !got.done {
		t.Error("splashDoneMsg should set done=true")
	}
}

func TestSplashWindowSizeMsgAppliesDimensions(t *testing.T) {
	m := NewSplashModel()
	got, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if got.width != 100 || got.height != 50 {
		t.Errorf("WindowSizeMsg not applied: %dx%d", got.width, got.height)
	}
}

func TestSplashViewNoPanic(t *testing.T) {
	m := NewSplashModel()
	m.width = 80
	m.height = 24
	out := m.View()
	if out.Content == "" {
		t.Error("SplashModel.View() returned empty Content")
	}
}
