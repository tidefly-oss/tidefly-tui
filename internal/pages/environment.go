package pages

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type envOption struct {
	value string
	label string
	desc  string
}

type EnvironmentModel struct {
	cfg     SetupConfig
	cursor  int
	options []envOption
}

func NewEnvironment(cfg SetupConfig) *EnvironmentModel {
	return &EnvironmentModel{
		cfg: cfg,
		options: []envOption{
			{
				value: EnvProduction,
				label: "Production",
				desc:  "Full stack via Docker — Postgres, Redis, Caddy, Backend, UI",
			},
			{
				value: EnvDevelopmentLocal,
				label: "Development (local)",
				desc:  "Infra via Docker, backend + UI run locally",
			},
		},
	}
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
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case key.Matches(keyMsg, keys.Enter):
			chosen := m.options[m.cursor]
			cfg := m.cfg
			cfg.Environment = chosen.value
			next := PageCaddy
			if chosen.value == EnvDevelopmentLocal {
				next = PageDevPaths
			}
			return m, func() tea.Msg {
				return NavigateTo{Page: next, Config: cfg}
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
	list := ""
	for i, opt := range m.options {
		if i == m.cursor {
			list += styles.MenuItemSelected.Render("") +
				lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(opt.label) + "\n" +
				lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(4).Render(opt.desc) + "\n\n"
		} else {
			list += fmt.Sprintf("   %s\n", opt.label) +
				lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(4).Render(opt.desc) + "\n\n"
		}
	}
	help := styles.Help.Render("↑/↓ select  •  enter confirm  •  q quit")
	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, list, help,
		),
	)
}
