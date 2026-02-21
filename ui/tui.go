package ui

import (
	"github.com/MatchaTi/vimnm/network"
	"github.com/charmbracelet/bubbles/list"
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
)

type model struct {
	list            list.Model
	state           sessionState
	password        textinput.Model
	selectedNetwork network.Wifi
}

func InitialModel() model {
	networks := network.GetNetworks()
	var items []list.Item

	for _, net := range networks {
		items = append(items, net)
	}

	mList := list.New(items, list.NewDefaultDelegate(), 60, 20)
	mList.Title = "Available Networks"

	ti := textinput.New()
	ti.Placeholder = "Enter password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.CharLimit = 64
	ti.Width = 40

	return model{
		list:     mList,
		state:    stateListView,
		password: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
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

					if i.Security != "" && i.Security != "--" {
						m.state = statePasswordView
						m.password.Focus()
						m.password.SetValue("")
						return m, nil
					}
					// Handle open network connection here
				}
			}
		case statePasswordView:
			switch msg.String() {
			case "esc", "h":
				m.state = stateListView
				m.password.Blur()
				return m, nil
			case "enter":
				// Handle connection logic here using m.selectedNetwork and m.password.Value()
				m.state = stateListView
				m.password.Blur()
				return m, nil
			}
		}

	}

	if m.state == stateListView {
		m.list, cmd = m.list.Update(msg)

	} else if m.state == statePasswordView {
		m.password, cmd = m.password.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.state == statePasswordView {
		view := "Enter password for " + m.selectedNetwork.SSID + ":\n\n"
		view += " " + m.password.View() + "\n\n"
		view += " [ Enter: Connect ] [ Esc/h: Batal ] \n"
		return appStyle.Render(view)
	}

	return appStyle.Render(m.list.View())
}
