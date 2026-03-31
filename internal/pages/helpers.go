package pages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

func navigate(page Page, cfg SetupConfig) tea.Cmd {
	return func() tea.Msg { return NavigateTo{Page: page, Config: cfg} }
}

func renderMenu(opts []string, cursor int) string {
	out := ""
	for i, opt := range opts {
		if i == cursor {
			out += styles.MenuItemSelected.Render("") +
				lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(opt) + "\n\n"
		} else {
			out += "   " + opt + "\n\n"
		}
	}
	return out
}
