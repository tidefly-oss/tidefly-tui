package pages

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

const labelPassword = "Password"

// ── Admin page ────────────────────────────────────────────────────────────────

type adminField int

const (
	fieldFirstName adminField = iota
	fieldLastName
	fieldEmail
	fieldPassword
	fieldCount
)

var adminLabels = [fieldCount]string{"First name", "Last name", "Email", labelPassword}

type AdminCreated struct{}
type AdminError struct{ Msg string }

type AdminModel struct {
	inputs  [fieldCount]textinput.Model
	focused adminField
	err     string
	loading bool
}

func NewAdmin() *AdminModel {
	m := &AdminModel{}
	for i := range m.inputs {
		t := textinput.New()
		t.Placeholder = adminLabels[i]
		t.CharLimit = 100
		t.Prompt = ""
		if adminField(i) == fieldPassword {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.inputs[i] = t
	}
	m.inputs[fieldFirstName].Focus()
	return m
}

func (m *AdminModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *AdminModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case AdminCreated:
		m.loading = false
		return m, func() tea.Msg { return NavigateTo{Page: PageDone} }

	case AdminError:
		m.loading = false
		m.err = msg.Msg
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Tab), key.Matches(msg, keys.Down):
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % fieldCount
			m.inputs[m.focused].Focus()
			return m, textinput.Blink

		case key.Matches(msg, keys.Up):
			m.inputs[m.focused].Blur()
			if m.focused == 0 {
				m.focused = fieldCount - 1
			} else {
				m.focused--
			}
			m.inputs[m.focused].Focus()
			return m, textinput.Blink

		case key.Matches(msg, keys.Enter):
			if m.focused < fieldCount-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
				return m, textinput.Blink
			}
			m.loading = true
			m.err = ""
			return m, submitAdmin(m.inputs)
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func submitAdmin(inputs [fieldCount]textinput.Model) tea.Cmd {
	firstName := strings.TrimSpace(inputs[fieldFirstName].Value())
	lastName := strings.TrimSpace(inputs[fieldLastName].Value())
	email := strings.TrimSpace(inputs[fieldEmail].Value())
	password := inputs[fieldPassword].Value()

	return func() tea.Msg {
		if firstName == "" || lastName == "" || email == "" || password == "" {
			return AdminError{"all fields are required"}
		}
		if len(password) < 8 {
			return AdminError{"password must be at least 8 characters"}
		}
		if !strings.Contains(email, "@") {
			return AdminError{"invalid email address"}
		}

		body := fmt.Sprintf(
			`{"first_name":%q,"last_name":%q,"email":%q,"password":%q}`,
			firstName, lastName, email, password,
		)

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:8181/api/v1/setup/admin", strings.NewReader(body))
		if err != nil {
			return AdminError{fmt.Sprintf("failed to create request: %v", err)}
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return AdminError{fmt.Sprintf("failed to connect to backend: %v", err)}
		}
		defer func() { _ = resp.Body.Close() }()

		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			return AdminCreated{}
		case http.StatusConflict:
			return AdminError{"setup already completed — a user already exists"}
		default:
			return AdminError{fmt.Sprintf("backend error: %s", resp.Status)}
		}
	}
}

func (m *AdminModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Admin Account"),
		styles.Subtitle.Render("Create your first administrator account"),
		"",
	)

	var form strings.Builder
	for i, input := range m.inputs {
		lbl := styles.InputLabel
		if adminField(i) == m.focused {
			lbl = lbl.Foreground(styles.Primary)
		}
		form.WriteString(lbl.Render(adminLabels[i]) + "\n")
		form.WriteString(input.View() + "\n\n")
	}

	errMsg := ""
	if m.err != "" {
		errMsg = styles.StatusErr.Render("✗ "+m.err) + "\n\n"
	}

	loading := ""
	if m.loading {
		loading = styles.StatusWarn.Render("⟳ creating account...") + "\n\n"
	}

	help := styles.Help.Render("tab/↑↓ between fields  •  enter confirm  •  q quit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, form.String(), errMsg, loading, help,
		),
	)
}
