package pages

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codifystudios/tidefly/tui/internal/styles"
)

// DetectionResult holds runtime detection results.
// Exported so AppModel can pass it to NewRuntime.
type DetectionResult struct {
	OS          string
	Arch        string
	DockerFound bool
	PodmanFound bool
}

// HomeModel is exported so AppModel can call Result() on it.
type HomeModel struct {
	spinner  spinner.Model
	result   *DetectionResult
	detected bool
}

func NewHome() *HomeModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)
	return &HomeModel{spinner: s}
}

// Result returns the detection result — nil until detection is complete.
func (m *HomeModel) Result() *DetectionResult {
	return m.result
}

func (m *HomeModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, detect())
}

func detect() tea.Cmd {
	return func() tea.Msg {
		result := DetectionResult{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		}
		result.DockerFound = commandExists("docker") && commandRunning("docker", "info")
		result.PodmanFound = commandExists("podman") && commandRunning("podman", "info")
		return result
	}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func commandRunning(cmd string, args ...string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, cmd, args...).Run() == nil
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DetectionResult:
		m.result = &msg
		m.detected = true
		return m, nil

	case spinner.TickMsg:
		if !m.detected {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Enter) && m.detected:
			docker := m.result != nil && m.result.DockerFound
			podman := m.result != nil && m.result.PodmanFound
			return m, func() tea.Msg {
				return NavigateTo{
					Page:   PageRuntime,
					Config: SetupConfig{},
					// Pass detection via a separate message so NewRuntime gets it
					Data: fmt.Sprintf("%v:%v", docker, podman),
				}
			}
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *HomeModel) View() string {
	if !m.detected {
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				styles.Subtitle.Render("Container Management Platform"),
				"",
				m.spinner.View()+"  Detecting your environment...",
			),
		)
	}

	r := m.result
	osLabel := map[string]string{
		"darwin":  "🍎 macOS",
		"linux":   "🐧 Linux",
		"windows": "🪟 Windows",
	}[r.OS]
	if osLabel == "" {
		osLabel = "💻 " + r.OS
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.InputLabel.Render("System"),
		fmt.Sprintf("  %s  (%s)", osLabel, r.Arch),
		"",
		styles.InputLabel.Render("Container Runtimes"),
		runtimeStatus("Docker", r.DockerFound),
		runtimeStatus("Podman", r.PodmanFound),
		"",
		styles.Help.Render("press enter to continue  •  q to quit"),
	)

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Subtitle.Render("Container Management Platform"),
			"",
			content,
		),
	)
}

func runtimeStatus(name string, found bool) string {
	if found {
		return fmt.Sprintf("  %s  %s", styles.StatusOK.Render("✓"), name)
	}
	return fmt.Sprintf(
		"  %s  %s  %s",
		styles.StatusErr.Render("✗"), name,
		lipgloss.NewStyle().Foreground(styles.Muted).Render("(not found — will be installed)"),
	)
}
