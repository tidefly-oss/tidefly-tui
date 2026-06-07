package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		var pageCmd tea.Cmd
		m.current, pageCmd = m.current.Update(msg)
		return m, tea.Batch(spinCmd, pageCmd)
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
		case pages.PageExtras:
			m.current = pages.NewExtras(m.cfg)
		case pages.PageStart:
			m.current = pages.NewStart(m.cfg)
		case pages.PageAdmin:
			m.current = pages.NewAdmin(m.cfg)
		case pages.PageDone:
			m.current = pages.NewDone(m.cfg)
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
	if src.DevPlanePath != "" {
		dst.DevPlanePath = src.DevPlanePath
	}
	if src.DevUIPath != "" {
		dst.DevUIPath = src.DevUIPath
	}
	dst.HardenCrowdSec = src.HardenCrowdSec
	dst.HardenFail2ban = src.HardenFail2ban
	dst.HardenCoraza = src.HardenCoraza
}

// ── Uninstall ─────────────────────────────────────────────────────────────────

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
	logWarn := func(msg string) { fmt.Printf("\033[33m[tidefly]\033[0m %s\n", msg) }

	// ── 1. Compose down (core services + volumes) ─────────────────────────────
	logInfo("Stopping core services via compose...")
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

	// ── 2. Remove all user containers with tidefly labels ────────────────────
	logInfo("Removing all Tidefly containers...")
	containerIDs := listContainersByLabel(ctx, rt, "tidefly.service")
	for _, id := range containerIDs {
		_ = exec.CommandContext(ctx, rt, "rm", "-f", id).Run()
	}

	// ── 3. Remove all networks with tidefly_ prefix ───────────────────────────
	logInfo("Removing all Tidefly networks...")
	networks := listNetworksByPrefix(ctx, rt, "tidefly_")
	for _, network := range networks {
		if err := exec.CommandContext(ctx, rt, "network", "rm", network).Run(); err != nil {
			logWarn(fmt.Sprintf("could not remove network %q: %v", network, err))
		}
	}

	// ── 4. Remove Tidefly images ──────────────────────────────────────────────
	logInfo("Removing Tidefly images...")
	for _, img := range []string{
		"tidefly/tidefly-plane:latest",
		"tidefly/tidefly-ui:latest",
		"tidefly/tidefly-caddy:latest",
	} {
		_ = exec.CommandContext(ctx, rt, "rmi", "-f", img).Run()
	}

	// ── 5. Remove config directory ────────────────────────────────────────────
	logInfo("Removing config directory...")
	if err := os.RemoveAll(baseDir); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\033[31m[tidefly]\033[0m failed to remove %s: %v\n", baseDir, err)
		os.Exit(1)
	}

	logOK("Uninstall complete — run tidefly-tui to install fresh.")
}

// listContainersByLabel returns container IDs that have a given label key.
func listContainersByLabel(ctx context.Context, rt, label string) []string {
	out, err := exec.CommandContext(ctx, rt, "ps", "-aq",
		"--filter", "label="+label,
		"--format", "{{.ID}}",
	).Output()
	if err != nil {
		return nil
	}
	var ids []string
	for _, id := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// listNetworksByPrefix returns all network names with the given prefix.
func listNetworksByPrefix(ctx context.Context, rt, prefix string) []string {
	out, err := exec.CommandContext(ctx, rt, "network", "ls",
		"--format", `{"Name":"{{.Name}}"}`,
	).Output()
	if err != nil {
		return nil
	}
	var networks []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var obj struct{ Name string }
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}
		if strings.HasPrefix(obj.Name, prefix) {
			networks = append(networks, obj.Name)
		}
	}
	return networks
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "uninstall":
			runUninstall()
			return
		case "help", "--help", "-h":
			fmt.Println("Usage: tidefly-tui [command]")
			fmt.Println("")
			fmt.Println("Commands:")
			fmt.Println("  (none)      Run the setup wizard")
			fmt.Println("  uninstall   Remove all Tidefly containers, networks, volumes and config")
			fmt.Println("  help        Show this help message")
			return
		}
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
