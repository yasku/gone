package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"gone/internal/sysinfo"
)

type networkRefreshMsg time.Time

func doNetworkRefresh() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return networkRefreshMsg(t) })
}

type NetworkModel struct {
	ifaces      []sysinfo.NetInterface
	filter      string
	filtering   bool
	filterInput textinput.Model
	styles      Styles
	width       int
	height      int
	ready       bool
	rxBars      map[string]progress.Model
	txBars      map[string]progress.Model
}

func newNetworkGaugeBar() progress.Model {
	return progress.New(
		progress.WithColors(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")),
		progress.WithoutPercentage(),
		progress.WithWidth(10),
	)
}

func NewNetworkModel() NetworkModel {
	fi := textinput.New()
	fi.Placeholder = "filter interfaces..."
	fi.CharLimit = 40

	return NetworkModel{
		filterInput: fi,
		styles:      DefaultStyles(),
		ifaces:      sysinfo.GetNetInterfaces(),
		ready:       true,
		rxBars:      make(map[string]progress.Model),
		txBars:      make(map[string]progress.Model),
	}
}

func (m *NetworkModel) ensureBars(iface sysinfo.NetInterface) {
	if _, ok := m.rxBars[iface.Name]; !ok {
		m.rxBars[iface.Name] = newNetworkGaugeBar()
		m.txBars[iface.Name] = newNetworkGaugeBar()
	}
}

func (m NetworkModel) Init() tea.Cmd {
	return doNetworkRefresh()
}

func (m NetworkModel) Update(msg tea.Msg) (NetworkModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case networkRefreshMsg:
		m.ifaces = sysinfo.GetNetInterfaces()
		for _, iface := range m.ifaces {
			if _, ok := m.rxBars[iface.Name]; !ok {
				m.rxBars[iface.Name] = newNetworkGaugeBar()
				m.txBars[iface.Name] = newNetworkGaugeBar()
			}
			rxPct := float64(iface.RxBytes) / float64(iface.RxBytes+1)
			if iface.RxBytes == 0 {
				rxPct = 0
			}
			txPct := float64(iface.TxBytes) / float64(iface.TxBytes+1)
			if iface.TxBytes == 0 {
				txPct = 0
			}
			rxBar := m.rxBars[iface.Name]
			txBar := m.txBars[iface.Name]
			rxBar.SetPercent(rxPct)
			txBar.SetPercent(txPct)
			m.rxBars[iface.Name] = rxBar
			m.txBars[iface.Name] = txBar
		}
		return m, doNetworkRefresh()

	case progress.FrameMsg:
		for _, name := range m.getInterfaceNames() {
			var cmd tea.Cmd
			m.rxBars[name], cmd = m.rxBars[name].Update(msg)
			cmds = append(cmds, cmd)
			m.txBars[name], cmd = m.txBars[name].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		key := msg.String()

		if m.filtering {
			switch key {
			case "esc":
				m.filtering = false
				m.filterInput.Blur()
				m.filterInput.Reset()
				return m, nil
			default:
				var cmd tea.Cmd
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		switch key {
		case "/":
			m.filtering = true
			cmd := m.filterInput.Focus()
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case "r":
			m.ifaces = sysinfo.GetNetInterfaces()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m NetworkModel) getInterfaceNames() []string {
	var names []string
	for _, iface := range m.ifaces {
		names = append(names, iface.Name)
	}
	return names
}

func (m NetworkModel) SetSize(w, h int) NetworkModel {
	m.width = w
	m.height = h
	m.filterInput.SetWidth(w - 8)
	return m
}

func (m NetworkModel) View() tea.View {
	if !m.ready {
		return tea.NewView("  Loading network info...")
	}

	var b strings.Builder

	filterText := m.styles.DimText.Render("  / filter interfaces  r refresh")
	b.WriteString(filterText + "\n\n")

	b.WriteString(m.viewInterfaces())

	return tea.NewView(b.String())
}

func (m NetworkModel) viewInterfaces() string {
	var b strings.Builder

	filtered := m.ifaces
	if m.filtering && m.filterInput.Value() != "" {
		filter := strings.ToLower(m.filterInput.Value())
		var filteredIfaces []sysinfo.NetInterface
		for _, iface := range m.ifaces {
			if strings.Contains(strings.ToLower(iface.Name), filter) {
				filteredIfaces = append(filteredIfaces, iface)
			}
		}
		filtered = filteredIfaces
	}

	if len(filtered) == 0 {
		b.WriteString(m.styles.DimText.Render("  No interfaces found."))
		return b.String()
	}

	cols := 4
	if m.width < 100 {
		cols = 2
	}

	for i, iface := range filtered {
		b.WriteString(m.interfaceBox(iface))
		if (i+1)%cols == 0 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (m NetworkModel) interfaceBox(iface sysinfo.NetInterface) string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00BCD4")).
		Render(iface.Name)

	rxBar, txBar := m.rxBars[iface.Name], m.txBars[iface.Name]

	rxLabel := m.styles.DimText.Render("RX")
	txLabel := m.styles.DimText.Render("TX")
	rxVal := sysinfo.HumanBytes(iface.RxBytes)
	txVal := sysinfo.HumanBytes(iface.TxBytes)

	boxWidth := m.width / 4
	if boxWidth < 25 {
		boxWidth = 25
	}

	content := fmt.Sprintf("%s\n%s %s  %s %s\n%s %s", title, rxLabel, rxBar.View(), rxVal, txLabel, txBar.View(), txVal)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(lipgloss.Color("#9B59B6"), lipgloss.Color("#00BCD4")).
		Padding(0, 2).
		Width(boxWidth - 2).
		Render(content)
}
