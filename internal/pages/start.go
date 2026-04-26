package pages

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tidefly-oss/tidefly-tui/internal/env"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

const (
	flagEnvFile   = "--env-file"
	envTypePrefix = "ENV_TYPE="
)

type startStepResult struct {
	step int
	err  error
}

type rollbackDone struct{ err error }

type startStep struct {
	label  string
	done   bool
	failed bool
	errMsg string
}

type StartModel struct {
	cfg         SetupConfig
	spinner     spinner.Model
	steps       []startStep
	current     int
	finished    bool
	hasError    bool
	rollingBack bool
}

func NewStart(cfg SetupConfig) *StartModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)
	return &StartModel{cfg: cfg, spinner: s, steps: buildSteps(cfg)}
}

func buildSteps(cfg SetupConfig) []startStep {
	backendLabel := "Starting backend"
	if cfg.WithDashboard {
		backendLabel = "Starting backend + dashboard"
	}
	return []startStep{
		{label: "Pulling Docker images"},
		{label: "Generating secrets"},
		{label: "Writing environment config"},
		{label: "Creating Docker networks"},
		{label: "Cleaning up existing containers"},
		{label: "Starting core services (Postgres, Redis, Caddy)"},
		{label: "Waiting for Postgres to be healthy"},
		{label: "Waiting for Redis to be healthy"},
		{label: backendLabel},
	}
}

func (m *StartModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			time.Sleep(50 * time.Millisecond)
			return startStepResult{step: -1}
		},
	)
}

func (m *StartModel) composePaths() (cf, envFile string) {
	isProd := m.cfg.Environment == EnvProduction
	baseDir := env.PlaneDir()
	if isProd {
		cf = filepath.Join(baseDir, "docker-compose.yaml")
		envFile = filepath.Join(baseDir, ".env")
	} else {
		cf = filepath.Join(baseDir, "docker-compose.dev.yaml")
		envFile = filepath.Join(baseDir, ".env.dev")
	}
	return
}

func (m *StartModel) rollback() tea.Cmd {
	cf, envFile := m.composePaths()
	rt := m.cfg.Runtime
	if rt == "" {
		rt = "docker"
	}
	return func() tea.Msg {
		if _, err := os.Stat(envFile); err != nil {
			return rollbackDone{}
		}
		args := []string{"compose", "-f", cf, flagEnvFile, envFile, "down", "--remove-orphans", "--volumes"}
		cmd := exec.CommandContext(context.Background(), rt, args...)
		cmd.Env = append(os.Environ(), envTypePrefix+m.cfg.Environment)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return rollbackDone{err: fmt.Errorf("rollback failed: %s", strings.TrimSpace(string(out)))}
		}
		return rollbackDone{}
	}
}

func podmanEnv(cmd *exec.Cmd, rt, socketPath, environment string) *exec.Cmd {
	cmd.Env = append(os.Environ(), envTypePrefix+environment)
	if rt == runtimePodman {
		sock := socketPath
		if sock == "" {
			sock = PodmanSocket
		}
		cmd.Env = append(cmd.Env, "DOCKER_HOST=unix://"+sock, "DOCKER_SOCK="+sock)
	}
	return cmd
}

func stepPullImages(cfg SetupConfig, rt string) error {
	images := []string{
		"tidefly/tidefly-plane:latest",
		"tidefly/tidefly-caddy:latest",
		"postgres:17-alpine",
		"redis:8-alpine",
	}
	if cfg.WithDashboard {
		images = append(images, "tidefly/tidefly-ui:latest")
	}
	for _, img := range images {
		cmd := exec.CommandContext(context.Background(), rt, "pull", img)
		out, e := cmd.CombinedOutput()
		if e != nil {
			return fmt.Errorf("failed to pull %s: %s", img, strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func stepGenerateSecrets(cfg SetupConfig, envFile string) error {
	if _, err := os.Stat(envFile); err == nil {
		return nil
	}

	dir := filepath.Dir(envFile)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	f, err := os.OpenFile(envFile, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create env file: %w", err)
	}
	defer f.Close()

	secrets, err := generateSecrets()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(f, secrets); err != nil {
		return fmt.Errorf("failed to write secrets: %w", err)
	}
	return nil
}

func stepWriteEnv(cfg SetupConfig, rt, envFile string) error {
	vars := map[string]string{
		"RUNTIME_TYPE":   rt,
		"RUNTIME_SOCKET": cfg.SocketPath,
	}
	if cfg.CaddyEnabled {
		vars["CADDY_ENABLED"] = "true"
		if !cfg.CaddyLater {
			vars["CADDY_BASE_DOMAIN"] = cfg.CaddyDomain
			vars["CADDY_ACME_EMAIL"] = cfg.CaddyEmail
			if cfg.CaddyStaging {
				vars["CADDY_ACME_STAGING"] = "true"
			}
		}
	} else {
		vars["CADDY_ENABLED"] = "false"
	}
	if cfg.SMTPEnabled {
		vars["SMTP_HOST"] = cfg.SMTPHost
		vars["SMTP_PORT"] = cfg.SMTPPort
		vars["SMTP_USER"] = cfg.SMTPUser
		vars["SMTP_PASSWORD"] = cfg.SMTPPassword
		vars["SMTP_FROM"] = cfg.SMTPFrom
		vars["SMTP_TLS"] = cfg.SMTPTLS
	}
	return writeEnvVars(envFile, vars)
}

func stepCreateNetworks(cfg SetupConfig, rt string) error {
	for _, network := range []string{"tidefly_proxy", "tidefly_internal"} {
		cmd := exec.CommandContext(
			context.Background(), rt,
			"network", "create", "--driver", "bridge", "--label", "tidefly.managed=true", network,
		)
		cmd.Env = append(os.Environ(), envTypePrefix+cfg.Environment)
		out, e := cmd.CombinedOutput()
		if e != nil && !strings.Contains(string(out), "already exists") {
			return fmt.Errorf("failed to create network %s: %s", network, strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func stepCleanup(cfg SetupConfig, rt, cf, envFile string) error {
	args := []string{"compose", "-f", cf, flagEnvFile, envFile, "down", "--remove-orphans"}
	cmd := exec.CommandContext(context.Background(), rt, args...)
	cmd.Env = append(os.Environ(), envTypePrefix+cfg.Environment)
	out, e := cmd.CombinedOutput()
	if e != nil &&
		!strings.Contains(string(out), "no such file") &&
		!strings.Contains(string(out), "does not exist") {
		return fmt.Errorf("cleanup failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func stepStartCore(cfg SetupConfig, rt, cf, envFile string) error {
	if _, err := os.Stat(cf); err != nil {
		return fmt.Errorf("compose file not found: %s", cf)
	}
	if _, err := os.Stat(envFile); err != nil {
		return fmt.Errorf(".env not found: %s", envFile)
	}
	args := []string{"compose", "-f", cf, flagEnvFile, envFile, "up", "-d", "postgres", "redis", "caddy"}
	cmd := podmanEnv(exec.CommandContext(context.Background(), rt, args...), rt, cfg.SocketPath, cfg.Environment)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return fmt.Errorf("failed to start core services: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func stepWaitHealthy(rt, container string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		out, e := exec.CommandContext(ctx, rt, "inspect", "--format", "{{.State.Health.Status}}", container).Output()
		cancel()
		if e == nil && strings.TrimSpace(string(out)) == "healthy" {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("%s not healthy after %s", container, timeout)
}

func stepStartBackend(cfg SetupConfig, rt, cf, envFile string) error {
	var args []string
	if cfg.WithDashboard {
		args = []string{
			"compose", "-f", cf, flagEnvFile, envFile,
			"--profile", "dashboard", "up", "-d", "backend", "ui",
		}
	} else {
		args = []string{"compose", "-f", cf, flagEnvFile, envFile, "up", "-d", "backend"}
	}
	cmd := podmanEnv(exec.CommandContext(context.Background(), rt, args...), rt, cfg.SocketPath, cfg.Environment)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return fmt.Errorf("failed to start backend: %s", strings.TrimSpace(string(out)))
	}
	time.Sleep(3 * time.Second)
	return nil
}

func (m *StartModel) runStep(step int) tea.Cmd {
	cfg := m.cfg
	rt := cfg.Runtime
	if rt == "" {
		rt = "docker"
	}
	cf, envFile := m.composePaths()
	steps := buildSteps(cfg)
	if step >= len(steps) {
		return nil
	}
	label := steps[step].label

	return func() tea.Msg {
		var err error
		switch {
		case strings.HasPrefix(label, "Pulling"):
			err = stepPullImages(cfg, rt)
		case strings.HasPrefix(label, "Generating"):
			err = stepGenerateSecrets(cfg, envFile)
		case strings.HasPrefix(label, "Writing"):
			err = stepWriteEnv(cfg, rt, envFile)
		case strings.HasPrefix(label, "Creating Docker networks"):
			err = stepCreateNetworks(cfg, rt)
		case strings.HasPrefix(label, "Cleaning up"):
			err = stepCleanup(cfg, rt, cf, envFile)
		case strings.HasPrefix(label, "Starting core"):
			err = stepStartCore(cfg, rt, cf, envFile)
		case strings.HasPrefix(label, "Waiting for Postgres"):
			err = stepWaitHealthy(rt, "tidefly_postgres", 90*time.Second)
		case strings.HasPrefix(label, "Waiting for Redis"):
			err = stepWaitHealthy(rt, "tidefly_redis", 30*time.Second)
		case strings.HasPrefix(label, "Starting backend"):
			err = stepStartBackend(cfg, rt, cf, envFile)
		}
		return startStepResult{step: step, err: err}
	}
}

func (m *StartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case rollbackDone:
		m.rollingBack = false
		if msg.err != nil {
			m.steps[m.current].errMsg += fmt.Sprintf(" (rollback error: %v)", msg.err)
		} else {
			m.steps[m.current].errMsg += " — rollback complete, all containers stopped"
		}
		return m, nil

	case startStepResult:
		if msg.step == -1 {
			return m, m.runStep(0)
		}
		if msg.err != nil {
			m.steps[msg.step].failed = true
			m.steps[msg.step].errMsg = msg.err.Error()
			m.finished = true
			m.hasError = true
			m.rollingBack = true
			return m, m.rollback()
		}
		m.steps[msg.step].done = true
		next := msg.step + 1
		if next < len(m.steps) {
			m.current = next
			return m, m.runStep(next)
		}
		m.finished = true
		return m, navigate(PageAdmin, m.cfg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) && !m.rollingBack {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *StartModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Setting up Tidefly"),
		styles.Subtitle.Render(fmt.Sprintf("Environment: %s", m.cfg.Environment)),
		"",
	)

	steps := ""
	for i, step := range m.steps {
		var icon string
		switch {
		case step.done:
			icon = styles.StatusOK.Render("✓")
		case step.failed:
			icon = styles.StatusErr.Render("✗")
		case i == m.current && !m.finished:
			icon = m.spinner.View()
		default:
			icon = lipgloss.NewStyle().Foreground(styles.Muted).Render("○")
		}
		line := fmt.Sprintf("  %s  %s", icon, step.label)
		if step.failed && step.errMsg != "" {
			line += "\n" + lipgloss.NewStyle().
				Foreground(styles.Danger).PaddingLeft(6).
				Render(step.errMsg)
		}
		steps += line + "\n\n"
	}

	var help string
	switch {
	case m.rollingBack:
		help = "\n" + styles.StatusWarn.Render("⟳ Rolling back — stopping all containers...") + "\n" +
			styles.Help.Render("please wait...")
	case m.hasError:
		help = "\n" + styles.StatusErr.Render("✗ Setup failed — all containers have been stopped") + "\n" +
			styles.Help.Render("fix the issue above and run tidefly-tui again  •  q to quit")
	default:
		help = styles.Help.Render("q to quit")
	}

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, steps, help,
		),
	)
}

func writeEnvVars(path string, vars map[string]string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	lines, updated, err := readAndUpdateLines(f, vars)
	if err != nil {
		return err
	}

	for k, v := range vars {
		if !updated[k] {
			lines = append(lines, k+"="+v)
		}
	}

	return rewriteFile(f, lines)
}

func readAndUpdateLines(f *os.File, vars map[string]string) ([]string, map[string]bool, error) {
	updated := make(map[string]bool)
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := replaceEnvLine(scanner.Text(), vars, updated)
		lines = append(lines, line)
	}
	return lines, updated, scanner.Err()
}

func replaceEnvLine(line string, vars map[string]string, updated map[string]bool) string {
	for k, v := range vars {
		re := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(k) + `\s*=.*$`)
		if re.MatchString(line) {
			updated[k] = true
			return k + "=" + v
		}
	}
	return line
}

func rewriteFile(f *os.File, lines []string) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, l := range lines {
		if _, err := fmt.Fprintln(w, l); err != nil {
			return err
		}
	}
	return w.Flush()
}
