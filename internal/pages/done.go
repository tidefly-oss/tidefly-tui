package pages

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type DoneModel struct{ cfg SetupConfig }

func NewDone(cfg SetupConfig) *DoneModel { return &DoneModel{cfg: cfg} }

func (m *DoneModel) Init() tea.Cmd { return nil }

func (m *DoneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *DoneModel) View() string {
	cfg := m.cfg
	backendPort := os.Getenv("APP_PORT")
	if backendPort == "" {
		backendPort = "8181"
	}

	header := styles.StatusOK.Render("✓ Tidefly is running!")

	var primaryURL string
	if cfg.WithDashboard {
		if cfg.CaddyEnabled && cfg.CaddyDomain != "" {
			primaryURL = "https://tidefly." + cfg.CaddyDomain
		} else {
			primaryURL = "http://localhost:3000"
		}
	} else {
		primaryURL = fmt.Sprintf("http://localhost:%s", backendPort)
	}

	access := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		lipgloss.NewStyle().Foreground(styles.White).Bold(true).Render("Access Tidefly:"),
		"",
		styles.StatusOK.Render("  → "+primaryURL),
		"",
	)

	links := ""
	if cfg.CaddyEnabled && cfg.CaddyDomain != "" {
		links = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.InputLabel.Render("Services:"),
			styles.Help.Render(fmt.Sprintf("  API  → https://api.%s", cfg.CaddyDomain)),
		)
	} else if cfg.CaddyLater {
		links = styles.StatusWarn.Render("  ⚠  Configure your domain in Settings → Proxy Domain")
	}

	help := "\n" + styles.Help.Render("press q to exit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, access, links, help,
		),
	)
}
