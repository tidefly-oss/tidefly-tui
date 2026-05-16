package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/env"
	"github.com/tidefly-oss/tidefly-tui/internal/pages"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
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
		case pages.PageDevPaths:
			m.current = pages.NewDevPaths(m.cfg)
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
	if src.DevPlanePath != "" {
		dst.DevPlanePath = src.DevPlanePath
	}
	if src.DevUIPath != "" {
		dst.DevUIPath = src.DevUIPath
	}
}

func runUninstall() {
	baseDir := env.PlaneDir()
	ctx := context.Background()

	rt := "docker"
	if _, err := exec.LookPath("docker"); err != nil {
		if _, err := exec.LookPath("podman"); err == nil {
			rt = "podman"
		}
	}

	logInfo := func(msg string) { fmt.Printf("\033[34m[tidefly]\033[0m %s\n", msg) }
	logOK := func(msg string) { fmt.Printf("\033[32m[tidefly]\033[0m %s\n", msg) }

	logInfo("Stopping and removing containers...")

	for _, name := range []string{"tidefly_ui", "tidefly_ui_dev"} {
		_ = exec.CommandContext(ctx, rt, "rm", "-f", name).Run()
	}

	for _, cf := range []string{
		filepath.Join(baseDir, "docker-compose.yaml"),
		filepath.Join(baseDir, "docker-compose.dev.yaml"),
	} {
		if _, err := os.Stat(cf); err == nil {
			cmd := exec.CommandContext(ctx, rt, "compose", "-f", cf, "down", "--remove-orphans", "--volumes")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		}
	}

	logInfo("Removing networks...")
	for _, network := range []string{
		"tidefly_internal", "tidefly_proxy",
		"tidefly_internal_dev", "tidefly_proxy_dev",
	} {
		_ = exec.CommandContext(ctx, rt, "network", "rm", network).Run()
	}

	logInfo("Removing config directory...")
	if err := os.RemoveAll(baseDir); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\033[31m[tidefly]\033[0m failed to remove %s: %v\n", baseDir, err)
		os.Exit(1)
	}

	logOK("Uninstall complete — run tidefly-tui to install fresh.")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "uninstall" {
		runUninstall()
		return
	}

	p := tea.NewProgram(
		NewApp(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
