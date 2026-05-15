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
			var env string
			var nextPage Page
			if m.cursor == 0 {
				env = EnvDevelopmentLocal
				nextPage = PageDevPaths
			} else {
				env = EnvProduction
				nextPage = PageCaddy
			}
			m.cfg.Environment = env
			cfg := m.cfg
			return m, func() tea.Msg {
				return NavigateTo{Page: nextPage, Config: cfg}
			}
		case key.Matches(keyMsg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *EnvironmentModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Environment"),
		styles.Subtitle.Render("How do you want to run Tidefly?"),
		"",
	)

	type envOpt struct {
		label string
		desc  string
		warn  string
	}

	opts := []envOpt{
		{
			label: "Development (local)",
			desc:  "Infra in Docker, backend + UI run locally — fast HMR, no image builds",
			warn:  "⚠  Requires Go + Node.js / pnpm installed locally",
		},
		{
			label: "Production",
			desc:  "Everything in Docker — optimized, secure defaults, your own SMTP",
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

		if isSelected && opt.warn != "" {
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
