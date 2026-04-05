package pages

import (
	tea "github.com/charmbracelet/bubbletea"
)

func navigate(page Page, cfg SetupConfig) tea.Cmd {
	return func() tea.Msg { return NavigateTo{Page: page, Config: cfg} }
}
