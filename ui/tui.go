package ui

import (
	"fmt"
	"strconv"

	"github.com/MatchaTi/vimnm/network"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle         = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	detailTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	detailValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	detailPanelStyle = lipgloss.NewStyle().PaddingLeft(2).BorderLeft(true).BorderForeground(lipgloss.Color("62"))
	keyConnect       = key.NewBinding(
		key.WithKeys("enter", "l"),
		key.WithHelp("enter/l", "connect"),
	)
	keyDisconnect = key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "disconnect"),
	)
)

type sessionState int

const (
	stateListView sessionState = iota
	statePasswordView
	stateConnecting
)

type connectResultMsg struct {
	err error
}

type disconnectResultMsg struct {
	err error
}

type model struct {
	list            list.Model
	state           sessionState
	password        textinput.Model
	selectedNetwork network.Wifi
	err             error
	spinner         spinner.Model
	width           int
	height          int
}

func (m model) renderFrame(content string) string {
	frameStyle := lipgloss.NewStyle()

	if m.width > 0 {
		frameStyle = frameStyle.Width(m.width)
	}

	if m.height > 0 {
		frameStyle = frameStyle.Height(m.height)
	}

	if m.width > 0 || m.height > 0 {
		content = frameStyle.Render(content)
	}

	return appStyle.Render(content)
}

func listPaneWidth(total int) int {
	if total <= 0 {
		return 60
	}

	left := total * 65 / 100
	if left < 45 {
		left = 45
	}

	if total-left < 28 {
		left = total - 28
	}

	if left < 20 {
		left = total
	}

	return left
}

func parseSignal(signal string) int {
	value, err := strconv.Atoi(signal)
	if err != nil {
		return 0
	}

	if value < 0 {
		return 0
	}

	if value > 100 {
		return 100
	}

	return value
}

func signalQuality(signal string) string {
	strength := parseSignal(signal)

	switch {
	case strength >= 80:
		return "Excellent"
	case strength >= 60:
		return "Good"
	case strength >= 40:
		return "Fair"
	case strength > 0:
		return "Weak"
	default:
		return "Unknown"
	}
}

func formatSecurity(security string) string {
	if security == "" || security == "--" {
		return "Open"
	}

	return security
}

func (m model) selectedItem() (network.Wifi, bool) {
	i, ok := m.list.SelectedItem().(network.Wifi)
	if !ok {
		return network.Wifi{}, false
	}

	return i, true
}

func (m model) detailView() string {
	selected, ok := m.selectedItem()
	if !ok {
		return detailPanelStyle.Render("No network selected")
	}

	status := "Available"
	if selected.Active {
		status = "Connected"
	}

	detail := "Network Details\n\n"
	detail += "SSID:\n" + detailValueStyle.Render(selected.SSID) + "\n\n"
	detail += "Status:\n" + detailValueStyle.Render(status) + "\n\n"
	detail += "Signal:\n" + detailValueStyle.Render(selected.Signal+"% ("+signalQuality(selected.Signal)+")") + "\n\n"
	detail += "Security:\n" + detailValueStyle.Render(formatSecurity(selected.Security)) + "\n\n"
	detail += detailTitleStyle.Render("Actions") + "\n"
	detail += "enter/l: connect\n"
	detail += "d: disconnect\n"
	detail += "/: filter\n"
	detail += "q: quit"

	if m.err != nil {
		detail += "\n\n" + detailTitleStyle.Render("Last Error") + "\n"
		detail += m.err.Error()
	}

	return detailPanelStyle.Render(detail)
}

func fetchNetworkItems() []list.Item {
	networks := network.GetNetworks()
	var items []list.Item

	for _, net := range networks {
		items = append(items, net)
	}

	return items
}

func InitialModel() model {

	mList := list.New(fetchNetworkItems(), list.NewDefaultDelegate(), 60, 20)
	mList.Title = "Available Networks"

	mList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{keyConnect, keyDisconnect}
	}
	mList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{keyConnect, keyDisconnect}
	}

	ti := textinput.New()
	ti.Placeholder = "Enter password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.CharLimit = 64
	ti.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		list:     mList,
		state:    stateListView,
		password: ti,
		spinner:  s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func connectCmd(ssid, password string) tea.Cmd {
	return func() tea.Msg {
		err := network.Connect(ssid, password)
		return connectResultMsg{err: err}
	}
}

func disconnectCmd(ssid string) tea.Cmd {
	return func() tea.Msg {
		err := network.Disconnect(ssid)
		return disconnectResultMsg{err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v
		m.list.SetSize(listPaneWidth(m.width), m.height)

	case connectResultMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.list.SetItems(fetchNetworkItems())
		}

		m.state = stateListView
		m.password.SetValue("")
		return m, nil

	case disconnectResultMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
			m.list.SetItems(fetchNetworkItems())
		}

		return m, nil

	case tea.KeyMsg:
		if m.state == stateListView && m.list.FilterState() == list.Filtering {
			break
		} else if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		switch m.state {
		case stateListView:
			switch msg.String() {
			case "d":
				i, ok := m.list.SelectedItem().(network.Wifi)
				if ok {
					if i.Active {
						m.selectedNetwork = i
						m.err = nil
						return m, disconnectCmd(i.SSID)
					} else {
						m.err = fmt.Errorf("Network '%s' is not currently connected", i.SSID)
						return m, nil
					}
				}
			case "enter", "l":
				i, ok := m.list.SelectedItem().(network.Wifi)
				if ok {
					m.selectedNetwork = i
					m.err = nil

					if i.Security == "" || i.Security == "--" {
						m.state = stateConnecting
						return m, tea.Batch(connectCmd(i.SSID, ""), m.spinner.Tick)
					}

					m.state = statePasswordView
					m.password.Focus()
					return m, nil
				}
			}

		case statePasswordView:
			switch msg.String() {
			case "esc":
				m.state = stateListView
				m.password.Blur()
				return m, nil

			case "enter":
				m.state = stateConnecting
				return m, tea.Batch(connectCmd(m.selectedNetwork.SSID, m.password.Value()), m.spinner.Tick)
			}
		}

	}

	if m.state == stateListView {
		m.list, cmd = m.list.Update(msg)

	} else if m.state == statePasswordView {
		m.password, cmd = m.password.Update(msg)

	} else if m.state == stateConnecting {
		m.spinner, cmd = m.spinner.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.state == stateConnecting {
		view := fmt.Sprintf("\n%s Connecting to %s... ", m.spinner.View(), m.selectedNetwork.SSID)
		return m.renderFrame(view)
	}

	if m.state == statePasswordView {
		view := "\nEnter the password: " + m.selectedNetwork.SSID + "\n\n"
		view += " " + m.password.View() + "\n\n"
		view += " [ Enter: Connect! ] [ Esc/h: Cancel ] \n"
		return m.renderFrame(view)
	}

	if m.width > 0 && m.height > 0 {
		m.list.SetSize(listPaneWidth(m.width), m.height)
	}

	leftPane := m.list.View()
	rightPane := m.detailView()
	joined := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	return m.renderFrame(joined)
}
