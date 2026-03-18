package pages

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codifystudios/tidefly/tui/internal/installer"
	"github.com/codifystudios/tidefly/tui/internal/styles"
)

type installLine struct{ text string }
type installDone struct{ err error }

type runtimeOption struct {
	runtime installer.Runtime
	label   string
	found   bool
}

type RuntimeModel struct {
	cursor  int
	options []runtimeOption
	// install state
	installing bool
	lines      []string
	linesCh    chan string
	spinner    spinner.Model
	err        string
}

func NewRuntime(dockerFound, podmanFound bool) *RuntimeModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	opts := []runtimeOption{
		{runtime: installer.Docker, label: "Docker", found: dockerFound},
		{runtime: installer.Podman, label: "Podman (rootless)", found: podmanFound},
	}

	// Auto-select if only one is found
	cursor := 0
	if !dockerFound && podmanFound {
		cursor = 1
	}

	return &RuntimeModel{
		cursor:  cursor,
		options: opts,
		spinner: s,
	}
}

func (m *RuntimeModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *RuntimeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case installLine:
		m.lines = append(m.lines, msg.text)
		// Keep last 8 lines
		if len(m.lines) > 8 {
			m.lines = m.lines[len(m.lines)-8:]
		}
		return m, listenInstall(m.linesCh)

	case installDone:
		if msg.err != nil {
			m.installing = false
			m.err = msg.err.Error()
			return m, nil
		}
		m.installing = false
		chosen := m.options[m.cursor]
		socketPath := "/var/run/docker.sock"
		if chosen.runtime == installer.Podman {
			socketPath = "/run/user/1000/podman/podman.sock"
		}
		return m, func() tea.Msg {
			return NavigateTo{
				Page: PageEnvironment,
				Config: SetupConfig{
					Runtime:    string(chosen.runtime),
					SocketPath: socketPath,
				},
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.installing {
			return m, nil
		}
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Enter):
			chosen := m.options[m.cursor]
			if chosen.found {
				socketPath := "/var/run/docker.sock"
				if chosen.runtime == installer.Podman {
					socketPath = "/run/user/1000/podman/podman.sock"
				}
				return m, func() tea.Msg {
					return NavigateTo{
						Page: PageEnvironment,
						Config: SetupConfig{
							Runtime:    string(chosen.runtime),
							SocketPath: socketPath,
						},
					}
				}
			}
			// Not found — install
			m.installing = true
			m.lines = nil
			m.err = ""
			m.linesCh = make(chan string, 64)
			return m, tea.Batch(
				m.spinner.Tick,
				listenInstall(m.linesCh),
				runInstall(chosen.runtime, m.linesCh),
			)
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func listenInstall(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return installLine{text: line}
	}
}

func runInstall(rt installer.Runtime, ch chan string) tea.Cmd {
	return func() tea.Msg {
		err := installer.Install(rt, ch)
		if err == nil {
			err = installer.PostInstallSetup(rt, ch)
		}
		close(ch)
		return installDone{err: err}
	}
}

func (m *RuntimeModel) View() string {
	if m.installing {
		log := ""
		for _, l := range m.lines {
			log += lipgloss.NewStyle().Foreground(styles.Muted).Render("  "+l) + "\n"
		}
		chosen := m.options[m.cursor]
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("Installing "+chosen.label),
			styles.Subtitle.Render("Using official install script — this may take a minute"),
			"",
			m.spinner.View()+"  Installing...",
			"",
			log,
		)
		return styles.Frame(termWidth, termHeight, content)
	}

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Container Runtime"),
		styles.Subtitle.Render("Which runtime should Tidefly use?"),
		"",
	)

	list := ""
	for i, opt := range m.options {
		status := ""
		if opt.found {
			status = " " + styles.StatusOK.Render("(installed)")
		} else {
			status = " " + lipgloss.NewStyle().Foreground(styles.Warning).Render("(not found — will be installed)")
		}

		if i == m.cursor {
			list += styles.MenuItemSelected.Render("") +
				lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(opt.label) +
				status + "\n\n"
		} else {
			list += fmt.Sprintf("   %s%s\n\n", opt.label, status)
		}
	}

	errMsg := ""
	if m.err != "" {
		errMsg = "\n" + styles.StatusErr.Render("✗ "+m.err) + "\n"
	}

	help := styles.Help.Render("↑/↓ select  •  enter confirm  •  q quit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, list, errMsg, help,
		),
	)
}
