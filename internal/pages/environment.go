package pages

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type EnvironmentModel struct {
	cursor int
	cfg    SetupConfig
}

func NewEnvironment(cfg SetupConfig) *EnvironmentModel {
	return &EnvironmentModel{cfg: cfg}
}

func (m *EnvironmentModel) Init() tea.Cmd { return nil }

func (m *EnvironmentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		env := EnvProduction
		if m.cursor == 0 {
			env = EnvDevelopment
		}
		m.cfg.Environment = env
		return m, navigate(PageDashboard, m.cfg)
	case key.Matches(keyMsg, keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m *EnvironmentModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Environment"),
		styles.Subtitle.Render("Where are you deploying Tidefly?"),
		"",
	)

	type envOpt struct {
		label string
		desc  string
		warn  string
	}
	opts := []envOpt{
		{
			label: "Development",
			desc:  "Hot reload, verbose logs, Mailpit — for contributing to Tidefly",
			warn:  "⚠  NOT for production use — no hardened security defaults",
		},
		{
			label: "Production",
			desc:  "Secure defaults, HTTPS, your own SMTP — for live deployments",
		},
	}

	list := ""
	for i, opt := range opts {
		isSelected := i == m.cursor
		label := opt.label
		if isSelected {
			label = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(opt.label)
		}
		cursor := "  "
		if isSelected {
			cursor = styles.MenuItemSelected.Render("")
		}
		desc := lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(3).Render(opt.desc)
		list += cursor + label + "\n" + desc + "\n"
		if i == 0 && isSelected && opt.warn != "" {
			list += "\n" + lipgloss.NewStyle().
				Foreground(styles.Warning).Bold(true).PaddingLeft(3).
				Render(opt.warn) + "\n"
		}
		list += "\n"
	}

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, list,
			styles.Help.Render("↑/↓ navigate  •  enter select  •  q quit"),
		),
	)
}
