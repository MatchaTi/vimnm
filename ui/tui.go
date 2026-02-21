package ui

import (
	"fmt"

	"github.com/MatchaTi/vimnm/network"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
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

type model struct {
	list            list.Model
	state           sessionState
	password        textinput.Model
	selectedNetwork network.Wifi
	err             error
	spinner         spinner.Model
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

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

	case tea.KeyMsg:
		if m.state == stateListView && m.list.FilterState() == list.Filtering {
			break
		} else if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		switch m.state {
		case stateListView:
			switch msg.String() {
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
		return appStyle.Render(view)
	}

	if m.state == statePasswordView {
		view := "\nEnter the password: " + m.selectedNetwork.SSID + "\n\n"
		view += " " + m.password.View() + "\n\n"
		view += " [ Enter: Connect! ] [ Esc/h: Cancel ] \n"
		return appStyle.Render(view)
	}

	view := m.list.View()

	if m.err != nil {
		view += "\n\n ❌ Failed: " + m.err.Error()
	}

	return appStyle.Render(view)
}
