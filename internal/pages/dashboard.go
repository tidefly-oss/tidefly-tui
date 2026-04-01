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
	return &DashboardModel{cfg: cfg, cursor: 0}
}

func (m *DashboardModel) Init() tea.Cmd {
	if m.cfg.Environment == EnvDevelopment {
		m.cfg.WithDashboard = false
		cfg := m.cfg
		return func() tea.Msg {
			return NavigateTo{Page: PageCaddy, Config: cfg}
		}
	}
	return nil
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
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
			cfg := m.cfg
			return m, func() tea.Msg {
				return NavigateTo{Page: PageCaddy, Config: cfg}
			}
		case key.Matches(keyMsg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *DashboardModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Web Dashboard"),
		styles.Subtitle.Render("Deploy the frontend alongside the API?"),
		"",
	)
	opts := []struct{ label, desc string }{
		{
			label: "Yes — with Dashboard",
			desc:  "Full Tidefly UI (SvelteKit) + API server",
		},
		{
			label: "No — API only",
			desc:  "Backend only — bring your own frontend or use the API directly",
		},
	}
	list := ""
	for i, o := range opts {
		isSelected := i == m.cursor
		label := o.label
		if isSelected {
			label = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(o.label)
		}
		cursor := "  "
		if isSelected {
			cursor = styles.MenuItemSelected.Render("")
		}
		desc := lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(3).Render(o.desc)
		list += cursor + label + "\n" + desc + "\n\n"
	}
	help := styles.Help.Render("↑/↓ navigate  •  enter select  •  q quit")
	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, list, help,
		),
	)
}
