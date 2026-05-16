package pages

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type devPathField int

const (
	devPathPlane devPathField = iota
	devPathUI
	devPathFieldCount
)

type DevPathsModel struct {
	inputs  [devPathFieldCount]textinput.Model
	focused devPathField
	err     string
	cfg     SetupConfig
}

func NewDevPaths(cfg SetupConfig) *DevPathsModel {
	home, _ := os.UserHomeDir()

	defaults := [devPathFieldCount]string{
		filepath.Join(home, "Desktop/development/tidefly/tidefly-plane"),
		filepath.Join(home, "Desktop/development/tidefly/tidefly-ui"),
	}

	var inputs [devPathFieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 512
		t.Prompt = ""
		t.Width = 60
		t.SetValue(defaults[i])
		inputs[i] = t
	}
	inputs[devPathPlane].Focus()

	return &DevPathsModel{
		inputs:  inputs,
		focused: devPathPlane,
		cfg:     cfg,
	}
}

func (m *DevPathsModel) Init() tea.Cmd { return textinput.Blink }

func (m *DevPathsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Tab), key.Matches(msg, keys.Down):
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % devPathFieldCount
			m.inputs[m.focused].Focus()
			return m, textinput.Blink

		case key.Matches(msg, keys.Up):
			m.inputs[m.focused].Blur()
			if m.focused == 0 {
				m.focused = devPathFieldCount - 1
			} else {
				m.focused--
			}
			m.inputs[m.focused].Focus()
			return m, textinput.Blink

		case key.Matches(msg, keys.Enter):
			if m.focused < devPathFieldCount-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
				return m, textinput.Blink
			}

			planePath := strings.TrimSpace(m.inputs[devPathPlane].Value())
			uiPath := strings.TrimSpace(m.inputs[devPathUI].Value())

			if planePath == "" || uiPath == "" {
				m.err = "both paths are required"
				return m, nil
			}
			if _, err := os.Stat(planePath); err != nil {
				m.err = "tidefly-plane path not found: " + planePath
				return m, nil
			}
			if _, err := os.Stat(uiPath); err != nil {
				m.err = "tidefly-ui path not found: " + uiPath
				return m, nil
			}

			m.cfg.DevPlanePath = planePath
			m.cfg.DevUIPath = uiPath
			cfg := m.cfg
			return m, func() tea.Msg {
				return NavigateTo{Page: PageStart, Config: cfg}
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m *DevPathsModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Local Dev Paths"),
		styles.Subtitle.Render("Where are the source repos on your machine?"),
		"",
		lipgloss.NewStyle().Foreground(styles.Muted).Render(
			"  Postgres + Redis run in Docker.\n"+
				"  Backend and UI are started locally with 'go run' and 'pnpm dev'.",
		),
		"",
	)

	labels := [devPathFieldCount]string{
		"tidefly-plane path",
		"tidefly-ui path",
	}

	form := ""
	for i, input := range m.inputs {
		lbl := styles.InputLabel
		if devPathField(i) == m.focused {
			lbl = lbl.Foreground(styles.Primary)
		}
		form += lbl.Render(labels[i]) + "\n"
		form += input.View() + "\n\n"
	}

	errMsg := ""
	if m.err != "" {
		errMsg = styles.StatusErr.Render("✗ "+m.err) + "\n\n"
	}

	help := styles.Help.Render("tab/↑↓ navigate  •  enter confirm  •  q quit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, form, errMsg, help,
		),
	)
}
