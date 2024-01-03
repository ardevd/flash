package tui

import (
	"context"
	"fmt"

	"github.com/ardevd/flash/internal/lnd"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
)

var channelsList []list.Item

type errMsg error
type LoadingModel struct {
	lndService *lndclient.GrpcLndServices
	ctx        context.Context
	spinner    spinner.Model
	quitting   bool
	err        error
}

func InitLoading(service *lndclient.GrpcLndServices) LoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return LoadingModel{spinner: s, lndService: service, ctx: context.Background()}
}

func (m LoadingModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick,
		tea.EnterAltScreen)
}

func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		windowSizeMsg = msg

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	case DataLoaded:
		dashboard := InitDashboard(m.lndService, lnd.NodeData(msg))
		return dashboard.Update(windowSizeMsg)

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m LoadingModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	str := fmt.Sprintf("\n\n   %s Loading node data...press q to quit\n\n", m.spinner.View())
	if m.quitting {
		return str + "\n"
	}
	return str
}
