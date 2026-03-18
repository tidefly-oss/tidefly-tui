package pages

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codifystudios/tidefly/tui/internal/styles"
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < 1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Enter):
			env := "production"
			if m.cursor == 0 {
				env = "development"
			}
			m.cfg.Environment = env
			cfg := m.cfg
			return m, func() tea.Msg {
				return NavigateTo{Page: PageDashboard, Config: cfg}
			}
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
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
			desc:  "Optimized, secure defaults, your own SMTP",
		},
	}

	list := ""
	for i, opt := range opts {
		isSelected := i == m.cursor

		label := opt.label
		desc := lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(3).Render(opt.desc)

		if isSelected {
			label = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(opt.label)
		}

		cursor := "  "
		if isSelected {
			cursor = styles.MenuItemSelected.Render("")
		}

		list += cursor + label + "\n" + desc + "\n"

		// Dev warning — only show when selected
		if i == 0 && isSelected && opt.warn != "" {
			list += "\n" + lipgloss.NewStyle().
				Foreground(styles.Warning).
				Bold(true).
				PaddingLeft(3).
				Render(opt.warn) + "\n"
		}
		list += "\n"
	}

	help := styles.Help.Render("↑/↓ navigate  •  enter select  •  q quit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, list, help,
		),
	)
}
