package pages

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codifystudios/tidefly/tui/internal/styles"
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

func (m *DoneModel) View() string {
	cfg := m.cfg

	backendPort := os.Getenv("APP_PORT")
	if backendPort == "" {
		backendPort = "8181"
	}
	frontendPort := os.Getenv("FRONTEND_PORT")
	if frontendPort == "" {
		frontendPort = "5173"
	}

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.StatusOK.Render("✓ Tidefly is running!"),
		"",
	)

	// Primary access URL
	primaryURL := ""
	if cfg.WithDashboard {
		if cfg.TraefikEnabled && cfg.TraefikDomain != "" {
			primaryURL = "https://tidefly." + cfg.TraefikDomain
		} else {
			primaryURL = "http://localhost:" + frontendPort
		}
	} else {
		primaryURL = "http://localhost:" + backendPort
	}

	access := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(styles.White).Bold(true).Render("Access Tidefly:"),
		"",
		styles.StatusOK.Render("  → "+primaryURL),
		"",
	)

	// Dev links
	links := ""
	if cfg.Environment == "development" {
		links = lipgloss.JoinVertical(
			lipgloss.Left,
			styles.InputLabel.Render("Dev tools:"),
			styles.Help.Render(fmt.Sprintf("  API         → http://localhost:%s", backendPort)),
			styles.Help.Render(fmt.Sprintf("  Swagger     → http://localhost:%s/swagger/index.html", backendPort)),
			styles.Help.Render("  Mailpit     → http://mailpit.localhost"),
			styles.Help.Render("  Traefik     → http://traefik.localhost"),
		)
		if cfg.MinIO {
			links += "\n" + styles.Help.Render("  MinIO       → http://minio.localhost")
			links += "\n" + styles.Help.Render("  S3 API      → http://s3.localhost")
		}
	} else {
		// Prod links
		if cfg.TraefikEnabled {
			links = lipgloss.JoinVertical(
				lipgloss.Left,
				styles.InputLabel.Render("Services:"),
				styles.Help.Render(fmt.Sprintf("  API         → https://api.%s", cfg.TraefikDomain)),
			)
			if cfg.WithDashboard {
				links += "\n" + styles.Help.Render(fmt.Sprintf("  Dashboard   → https://tidefly.%s", cfg.TraefikDomain))
			}
		}
	}

	help := "\n" + styles.Help.Render("press q to exit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, access, links, help,
		),
	)
}
