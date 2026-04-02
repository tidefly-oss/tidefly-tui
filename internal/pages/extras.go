package pages

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type ExtrasModel struct {
	cursor  int
	options []extraOption
	cfg     SetupConfig
}

type extraOption struct {
	label   string
	desc    string
	enabled bool
}

func (m *ExtrasModel) Init() tea.Cmd { return nil }

func (m *ExtrasModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case keyMsg.String() == " ":
			m.options[m.cursor].enabled = !m.options[m.cursor].enabled
		case key.Matches(keyMsg, keys.Enter):
			cfg := m.cfg
			return m, func() tea.Msg {
				return NavigateTo{Page: PageStart, Config: cfg}
			}
		case key.Matches(keyMsg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ExtrasModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Optional Services"),
		styles.Subtitle.Render("Always included: Traefik · Postgres · Redis"),
		"",
	)

	list := ""
	for i, opt := range m.options {
		checkbox := lipgloss.NewStyle().Foreground(styles.Muted).Render("[ ]")
		if opt.enabled {
			checkbox = styles.StatusOK.Render("[✓]")
		}

		label := opt.label
		desc := lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(7).Render(opt.desc)

		if i == m.cursor {
			label = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(opt.label)
		}

		cur := "  "
		if i == m.cursor {
			cur = styles.MenuItemSelected.Render("")
		}

		list += cur + checkbox + "  " + label + "\n" + desc + "\n\n"
	}

	help := styles.Help.Render("↑/↓ navigate  •  space toggle  •  enter confirm  •  q quit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, list, help,
		),
	)
}
