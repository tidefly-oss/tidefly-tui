package pages

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type DoneModel struct {
	cfg SetupConfig
}

func NewDone(cfg SetupConfig) *DoneModel {
	return &DoneModel{cfg: cfg}
}

func (m *DoneModel) Init() tea.Cmd { return nil }

func (m *DoneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
	}
	return m, nil
}

func dashboardURL(cfg SetupConfig) string {
	switch {
	case cfg.CaddyEnabled && cfg.CaddyDomain != "":
		return "https://dashboard." + cfg.CaddyDomain
	case cfg.Environment == EnvDevelopmentLocal:
		return "http://localhost:5173"
	default:
		return "http://localhost:3000"
	}
}

func (m *DoneModel) View() string {
	cfg := m.cfg

	backendPort := os.Getenv("APP_PORT")
	if backendPort == "" {
		backendPort = "8181"
	}

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.StatusOK.Render("✓ Tidefly is running!"),
		"",
	)

	primaryURL := dashboardURL(cfg)

	access := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(styles.White).Bold(true).Render("Access Tidefly:"),
		"",
		styles.StatusOK.Render("  → "+primaryURL),
		"",
	)

	var links string
	switch {
	case cfg.Environment == EnvDevelopmentLocal:
		links = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.InputLabel.Render("Start your dev environment:"),
			"",
			styles.Help.Render("  Terminal 1 →  cd "+cfg.DevPlanePath+" && air"),
			styles.Help.Render("  Terminal 2 →  cd "+cfg.DevUIPath+" && pnpm dev"),
			"",
			styles.Help.Render(fmt.Sprintf("  API     → http://localhost:%s", backendPort)),
			styles.Help.Render(fmt.Sprintf("  Docs    → http://localhost:%s/docs", backendPort)),
			styles.Help.Render("  UI      → http://localhost:5173"),
		)
	case cfg.CaddyEnabled && cfg.CaddyDomain != "":
		links = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.InputLabel.Render("Services:"),
			styles.Help.Render(fmt.Sprintf("  Dashboard → https://dashboard.%s", cfg.CaddyDomain)),
			styles.Help.Render(fmt.Sprintf("  API Docs  → https://dashboard.%s/docs", cfg.CaddyDomain)),
		)
	default:
		links = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.InputLabel.Render("Services:"),
			styles.Help.Render("  Dashboard → http://localhost:3000"),
			styles.Help.Render(fmt.Sprintf("  API       → http://localhost:%s", backendPort)),
			styles.Help.Render(fmt.Sprintf("  API Docs  → http://localhost:%s/docs", backendPort)),
		)
	}

	help := "\n" + styles.Help.Render("press q to exit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, access, links, help,
		),
	)
}
