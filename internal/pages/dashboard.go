package pages

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type DashboardModel struct {
	cursor int
	cfg    SetupConfig
}

func NewDashboard(cfg SetupConfig) *DashboardModel {
	return &DashboardModel{cfg: cfg, cursor: 1}
}

func (m *DashboardModel) Init() tea.Cmd {
	if m.cfg.Environment == EnvDevelopment {
		m.cfg.WithDashboard = false
		return navigate(PageCaddy, m.cfg)
	}
	return nil
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, keys.Down):
		if m.cursor < 1 {
			m.cursor++
		}
	case key.Matches(keyMsg, keys.Enter):
		m.cfg.WithDashboard = m.cursor == 0
		return m, navigate(PageCaddy, m.cfg)
	case key.Matches(keyMsg, keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m *DashboardModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Web Dashboard"),
		styles.Subtitle.Render("Deploy the Tidefly UI alongside the API?"),
		"",
	)
	opts := []string{
		"Yes — with Dashboard  (Full Tidefly UI + API)",
		"No — API only  (manage via API or connect your own UI)",
	}
	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, renderMenu(opts, m.cursor),
			styles.Help.Render("↑/↓ navigate  •  enter select  •  q quit"),
		),
	)
}
