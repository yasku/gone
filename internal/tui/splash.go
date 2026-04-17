package tui

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type splashDoneMsg struct{}

func splashDone() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		return splashDoneMsg{}
	})
}

const goneLogo = `
  ░██████╗░░█████╗░███╗░░██╗███████╗
  ██╔════╝░██╔══██╗████╗░██║██╔════╝
  ██║░░██╗░██║░░██║██╔██╗██║█████╗░░
  ██║░░╚██╗██║░░██║██║╚████║██╔══╝░░
  ╚██████╔╝╚█████╔╝██║░╚███║███████╗
  ░╚═════╝░░╚════╝░╚═╝░░╚══╝╚══════╝`

type SplashModel struct {
	spinner spinner.Model
	width   int
	height  int
	done    bool
}

func NewSplashModel() SplashModel {
	s := spinner.New(
		spinner.WithSpinner(spinner.Globe),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6"))),
	)
	return SplashModel{spinner: s}
}

func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, splashDone())
}

func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg.(type) {
	case splashDoneMsg:
		m.done = true
		return m, nil
	case tea.WindowSizeMsg:
		msg := msg.(tea.WindowSizeMsg)
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SplashModel) View() string {
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BCD4")).
		Bold(true)

	taglineStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(
			lipgloss.Color("#9B59B6"),
			lipgloss.Color("#00BCD4"),
		).
		Padding(1, 4)

	content := strings.Join([]string{
		logoStyle.Render(goneLogo),
		"",
		taglineStyle.Render("hunt. select. trash."),
		"",
		m.spinner.View() + "  initializing...",
	}, "\n")

	box := boxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
