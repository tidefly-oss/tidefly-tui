package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codifystudios/tidefly/tui/internal/env"
	"github.com/codifystudios/tidefly/tui/internal/pages"
	"github.com/codifystudios/tidefly/tui/internal/styles"
)

type AppModel struct {
	current pages.Model
	cfg     pages.SetupConfig
	width   int
	height  int
	ready   bool
	spinner spinner.Model
}

func NewApp() AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)
	return AppModel{
		current: pages.NewHome(),
		spinner: s,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.current.Init())
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		m.ready = true
		pages.SetSize(ws.Width, ws.Height)
	}

	if !m.ready {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if nav, ok := msg.(pages.NavigateTo); ok {
		mergeConfig(&m.cfg, nav.Config)

		switch nav.Page {
		case pages.PageHome:
			m.current = pages.NewHome()
		case pages.PageRuntime:
			var d *pages.DetectionResult
			if r, ok := m.current.(*pages.HomeModel); ok {
				d = r.Result()
			}
			docker, podman := false, false
			if d != nil {
				docker = d.DockerFound
				podman = d.PodmanFound
			}
			m.current = pages.NewRuntime(docker, podman)
		case pages.PageEnvironment:
			m.current = pages.NewEnvironment(m.cfg)
		case pages.PageDashboard:
			m.current = pages.NewDashboard(m.cfg)
		case pages.PageCaddy:
			m.current = pages.NewCaddy(m.cfg)
		case pages.PageSMTP:
			m.current = pages.NewSMTP(m.cfg)
		case pages.PageStart:
			m.current = pages.NewStart(m.cfg)
		case pages.PageAdmin:
			m.current = pages.NewAdmin()
		case pages.PageDone:
			m.current = pages.NewDone(m.cfg)
		case pages.PageExtras:
		}
		return m, m.current.Init()
	}

	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	if !m.ready {
		return "\n\n  " + m.spinner.View() + "  Loading..."
	}
	return m.current.View()
}

func mergeConfig(dst *pages.SetupConfig, src pages.SetupConfig) {
	if src.Runtime != "" {
		dst.Runtime = src.Runtime
	}
	if src.SocketPath != "" {
		dst.SocketPath = src.SocketPath
	}
	if src.Environment != "" {
		dst.Environment = src.Environment
	}
	dst.WithDashboard = src.WithDashboard
	dst.CaddyEnabled = src.CaddyEnabled
	if src.CaddyDomain != "" {
		dst.CaddyDomain = src.CaddyDomain
	}
	if src.CaddyEmail != "" {
		dst.CaddyEmail = src.CaddyEmail
	}
	dst.CaddyStaging = src.CaddyStaging
	dst.SMTPEnabled = src.SMTPEnabled
	if src.SMTPHost != "" {
		dst.SMTPHost = src.SMTPHost
	}
	if src.SMTPPort != "" {
		dst.SMTPPort = src.SMTPPort
	}
	if src.SMTPUser != "" {
		dst.SMTPUser = src.SMTPUser
	}
	if src.SMTPPassword != "" {
		dst.SMTPPassword = src.SMTPPassword
	}
	if src.SMTPFrom != "" {
		dst.SMTPFrom = src.SMTPFrom
	}
	if src.SMTPTLS != "" {
		dst.SMTPTLS = src.SMTPTLS
	}
}

func main() {
	_ = env.Load()

	p := tea.NewProgram(
		NewApp(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
