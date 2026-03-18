package pages

import (
	"bufio"
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

	"github.com/codifystudios/tidefly/tui/internal/styles"
)

type startStepResult struct {
	step int
	err  error
}

type startStep struct {
	label  string
	done   bool
	failed bool
	errMsg string
}

type StartModel struct {
	cfg      SetupConfig
	spinner  spinner.Model
	steps    []startStep
	current  int
	finished bool
	hasError bool
}

// root gibt den Projekt-Root zurück.
// task installer ruft "cd tui && go run ..." auf — cwd ist tui/.
func root() string {
	cwd, _ := os.Getwd()
	if filepath.Base(cwd) == "tui" {
		return filepath.Dir(cwd)
	}
	// direkt vom Root gestartet
	if _, err := os.Stat(filepath.Join(cwd, "tui")); err == nil {
		return cwd
	}
	return filepath.Dir(cwd)
}

func composePaths(prod bool) (composeF, envF string) {
	r := root()
	if prod {
		return filepath.Join(r, "deploy", "prod", "docker-compose.yaml"),
			filepath.Join(r, "deploy", "prod", ".env")
	}
	return filepath.Join(r, "deploy", "dev", "docker-compose.yaml"),
		filepath.Join(r, "deploy", "dev", ".env")
}

func scriptsDir() string { return filepath.Join(root(), "scripts") }
func backendDir() string { return filepath.Join(root(), "backend") }

func NewStart(cfg SetupConfig) *StartModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)
	return &StartModel{cfg: cfg, spinner: s, steps: buildSteps(cfg)}
}

func buildSteps(cfg SetupConfig) []startStep {
	// Dev: nur Infra (kein Frontend-Container)
	// Prod: Backend + optional Frontend
	coreLabel := "Starting infrastructure (Traefik, Postgres, Redis, Mailpit)"
	if cfg.Environment == "production" {
		coreLabel = "Starting services (Traefik, Postgres, Redis, Backend"
		if cfg.WithDashboard {
			coreLabel += ", Frontend"
		}
		if cfg.MinIO {
			coreLabel += ", MinIO"
		}
		coreLabel += ")"
	} else if cfg.MinIO {
		coreLabel = "Starting infrastructure (Traefik, Postgres, Redis, Mailpit, MinIO)"
	}

	return []startStep{
		{label: "Generating secrets"},
		{label: "Writing environment config"},
		{label: coreLabel},
		{label: "Waiting for Postgres"},
		{label: "Waiting for Redis"},
		{label: "Running database migrations"},
	}
}

func (m *StartModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick, func() tea.Msg {
			time.Sleep(50 * time.Millisecond)
			return startStepResult{step: -1}
		},
	)
}

func (m *StartModel) runStep(step int) tea.Cmd {
	cfg := m.cfg
	isProd := cfg.Environment == "production"
	runtime := cfg.Runtime
	if runtime == "" {
		runtime = "docker"
	}

	cf, ef := composePaths(isProd)

	composeArgs := func(extra ...string) []string {
		args := []string{"compose", "-f", cf, "--env-file", ef}
		return append(args, extra...)
	}

	podmanSock := cfg.SocketPath
	if podmanSock == "" {
		podmanSock = "/run/user/1000/podman/podman.sock"
	}

	withEnv := func(cmd *exec.Cmd) *exec.Cmd {
		if runtime == "podman" {
			cmd.Env = append(
				os.Environ(),
				"DOCKER_HOST=unix://"+podmanSock,
				"DOCKER_SOCK="+podmanSock,
			)
		}
		return cmd
	}

	steps := buildSteps(cfg)
	if step >= len(steps) {
		return nil
	}
	label := steps[step].label

	return func() tea.Msg {
		var err error

		switch {
		case strings.HasPrefix(label, "Generating secrets"):
			scriptPath := filepath.Join(scriptsDir(), "rotate-secrets.sh")
			if _, statErr := os.Stat(scriptPath); statErr != nil {
				err = fmt.Errorf("script not found: %s (root: %s)", scriptPath, root())
				break
			}
			err = runScript(scriptPath)
			if err == nil && cfg.MinIO {
				_ = runScript(scriptPath, "--minio")
			}

		case strings.HasPrefix(label, "Writing environment"):
			vars := map[string]string{
				"RUNTIME_TYPE":   runtime,
				"RUNTIME_SOCKET": cfg.SocketPath,
			}
			if cfg.TraefikEnabled {
				vars["TRAEFIK_ENABLED"] = "true"
				vars["TRAEFIK_BASE_DOMAIN"] = cfg.TraefikDomain
				vars["TRAEFIK_ACME_EMAIL"] = cfg.TraefikEmail
				if cfg.TraefikStaging {
					vars["TRAEFIK_ACME_STAGING"] = "true"
					vars["TRAEFIK_ACME_CA_SERVER"] = "https://acme-staging-v02.api.letsencrypt.org/directory"
				}
			}
			if cfg.SMTPEnabled {
				vars["SMTP_HOST"] = cfg.SMTPHost
				vars["SMTP_PORT"] = cfg.SMTPPort
				vars["SMTP_USER"] = cfg.SMTPUser
				vars["SMTP_PASSWORD"] = cfg.SMTPPassword
				vars["SMTP_FROM"] = cfg.SMTPFrom
				vars["SMTP_TLS"] = cfg.SMTPTLS
			}
			// MinIO profile — nur in Dev via profiles, in Prod immer im Compose
			if !isProd {
				if cfg.MinIO {
					vars["COMPOSE_PROFILES"] = "minio"
				} else {
					vars["COMPOSE_PROFILES"] = ""
				}
			}
			err = writeEnvVars(ef, vars)

		case strings.HasPrefix(label, "Starting"):
			if _, statErr := os.Stat(cf); statErr != nil {
				err = fmt.Errorf("compose file not found: %s", cf)
				break
			}
			if _, statErr := os.Stat(ef); statErr != nil {
				err = fmt.Errorf("env file not found: %s", ef)
				break
			}

			var args []string
			var services []string

			if isProd {
				// Prod: Backend + optional Frontend/MinIO via profiles
				services = []string{"traefik", "postgres", "redis", "backend"}
				args = []string{"up", "-d"}
				if cfg.WithDashboard {
					args = append([]string{"--profile", "dashboard"}, args...)
					services = append(services, "frontend")
				}
				if cfg.MinIO {
					args = append([]string{"--profile", "minio"}, args...)
					services = append(services, "minio")
				}
			} else {
				// Dev: nur Infra — kein Frontend-Container
				services = []string{"traefik", "postgres", "redis", "mailpit"}
				args = []string{"up", "-d"}
				if cfg.MinIO {
					args = append([]string{"--profile", "minio"}, args...)
					services = append(services, "minio")
				}
			}

			cmd := withEnv(exec.Command(runtime, append(composeArgs(args...), services...)...))
			out, e := cmd.CombinedOutput()
			if e != nil {
				err = fmt.Errorf("compose up failed: %s", strings.TrimSpace(string(out)))
			}

		case strings.HasPrefix(label, "Waiting for Postgres"):
			deadline := time.Now().Add(60 * time.Second)
			ready := false
			for time.Now().Before(deadline) {
				out, e := exec.Command(
					runtime, "inspect",
					"--format", "{{.State.Health.Status}}",
					"tidefly_postgres",
				).Output()
				if e == nil && strings.TrimSpace(string(out)) == "healthy" {
					ready = true
					break
				}
				time.Sleep(2 * time.Second)
			}
			if !ready {
				err = fmt.Errorf("postgres not healthy after 60s")
			}

		case strings.HasPrefix(label, "Waiting for Redis"):
			deadline := time.Now().Add(30 * time.Second)
			ready := false
			for time.Now().Before(deadline) {
				out, e := exec.Command(
					runtime, "inspect",
					"--format", "{{.State.Health.Status}}",
					"tidefly_redis",
				).Output()
				if e == nil && strings.TrimSpace(string(out)) == "healthy" {
					ready = true
					break
				}
				time.Sleep(2 * time.Second)
			}
			if !ready {
				err = fmt.Errorf("redis not healthy after 30s")
			}

		case strings.HasPrefix(label, "Running database"):
			out, e := exec.Command(
				"bash", "-c",
				fmt.Sprintf(
					"cd %s && set -a && source %s && set +a && go run ./cmd/tidefly --migrate-only",
					backendDir(), ef,
				),
			).CombinedOutput()
			if e != nil {
				err = fmt.Errorf("migrations failed: %s", strings.TrimSpace(string(out)))
			}
		}

		return startStepResult{step: step, err: err}
	}
}

func (m *StartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startStepResult:
		if msg.step == -1 {
			return m, m.runStep(0)
		}
		if msg.err != nil {
			m.steps[msg.step].failed = true
			m.steps[msg.step].errMsg = msg.err.Error()
			m.finished = true
			m.hasError = true
			return m, nil
		}
		m.steps[msg.step].done = true
		next := msg.step + 1
		if next < len(m.steps) {
			m.current = next
			return m, m.runStep(next)
		}
		m.finished = true
		return m, func() tea.Msg { return NavigateTo{Page: PageAdmin, Config: m.cfg} }

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch {
		case m.hasError && msg.String() == "r":
			for i, s := range m.steps {
				if s.failed {
					m.steps[i].failed = false
					m.steps[i].errMsg = ""
					m.finished = false
					m.hasError = false
					m.current = i
					return m, tea.Batch(m.spinner.Tick, m.runStep(i))
				}
			}
		case key.Matches(msg, keys.Quit):
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

	help := ""
	if m.hasError {
		help = "\n" + styles.StatusWarn.Render("Setup failed") + "\n" +
			styles.Help.Render("r to retry  •  q to quit")
	} else {
		help = styles.Help.Render("q to quit")
	}

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, steps, help,
		),
	)
}

func runScript(path string, args ...string) error {
	cmd := exec.Command("bash", append([]string{path}, args...)...)
	return cmd.Run()
}

func writeEnvVars(path string, vars map[string]string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	updated := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		for k, v := range vars {
			re := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(k) + `\s*=.*$`)
			if re.MatchString(line) {
				line = k + "=" + v
				updated[k] = true
				break
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for k, v := range vars {
		if !updated[k] {
			lines = append(lines, k+"="+v)
		}
	}

	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	return w.Flush()
}
