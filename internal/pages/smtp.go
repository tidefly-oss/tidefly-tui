package pages

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type smtpField int

const (
	smtpHost smtpField = iota
	smtpPort
	smtpUser
	smtpPassword
	smtpFrom
	smtpFieldCount
)

type SMTPModel struct {
	enabled   bool
	step      int
	focused   smtpField
	inputs    [smtpFieldCount]textinput.Model
	tlsCursor int
	cfg       SetupConfig
}

func NewSMTP(cfg SetupConfig) *SMTPModel {
	placeholders := [smtpFieldCount]string{
		"smtp.example.com", "587", "apikey", "", "noreply@example.com",
	}
	var inputs [smtpFieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 255
		t.Prompt = ""
		if smtpField(i) == smtpPassword {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		inputs[i] = t
	}
	inputs[smtpPort].SetValue("587")

	return &SMTPModel{
		enabled:   false,
		step:      0,
		inputs:    inputs,
		tlsCursor: 1,
		cfg:       cfg,
	}
}

func (m *SMTPModel) Init() tea.Cmd { return textinput.Blink }

func (m *SMTPModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch m.step {
	case 0:
		switch {
		case key.Matches(keyMsg, keys.Up), key.Matches(keyMsg, keys.Down):
			m.enabled = !m.enabled
		case key.Matches(keyMsg, keys.Enter):
			if !m.enabled {
				m.cfg.SMTPEnabled = false
				return m, navigate(PageStart, m.cfg)
			}
			m.step = 1
			m.inputs[0].Focus()
			return m, textinput.Blink
		case key.Matches(keyMsg, keys.Quit):
			return m, tea.Quit
		}

	case 1:
		switch {
		case key.Matches(keyMsg, keys.Tab), key.Matches(keyMsg, keys.Down):
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % smtpFieldCount
			m.inputs[m.focused].Focus()
			return m, textinput.Blink
		case key.Matches(keyMsg, keys.Up):
			m.inputs[m.focused].Blur()
			if m.focused == 0 {
				m.focused = smtpFieldCount - 1
			} else {
				m.focused--
			}
			m.inputs[m.focused].Focus()
			return m, textinput.Blink
		case key.Matches(keyMsg, keys.Enter):
			if m.focused < smtpFieldCount-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
				return m, textinput.Blink
			}
			m.inputs[m.focused].Blur()
			m.step = 2
			return m, nil
		case key.Matches(keyMsg, keys.Quit):
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.inputs[m.focused], cmd = m.inputs[m.focused].Update(keyMsg)
			return m, cmd
		}

	case 2:
		switch {
		case key.Matches(keyMsg, keys.Up):
			if m.tlsCursor > 0 {
				m.tlsCursor--
			}
		case key.Matches(keyMsg, keys.Down):
			if m.tlsCursor < 2 {
				m.tlsCursor++
			}
		case key.Matches(keyMsg, keys.Enter):
			tlsValues := []string{"none", "starttls", "tls"}
			m.cfg.SMTPEnabled = true
			m.cfg.SMTPHost = m.inputs[smtpHost].Value()
			m.cfg.SMTPPort = m.inputs[smtpPort].Value()
			m.cfg.SMTPUser = m.inputs[smtpUser].Value()
			m.cfg.SMTPPassword = m.inputs[smtpPassword].Value()
			m.cfg.SMTPFrom = m.inputs[smtpFrom].Value()
			m.cfg.SMTPTLS = tlsValues[m.tlsCursor]
			return m, navigate(PageStart, m.cfg)
		case key.Matches(keyMsg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *SMTPModel) View() string {
	labels := [smtpFieldCount]string{
		"SMTP Host", "Port", "Username", "Password", "From address",
	}

	switch m.step {
	case 0:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("Email / SMTP"),
			styles.Subtitle.Render("Configure outgoing email for notifications and alerts?"),
			"",
			lipgloss.NewStyle().Foreground(styles.Muted).Render(
				"  Works with any SMTP provider:\n"+
					"  Resend · Postmark · Mailgun · your own server",
			),
			"",
		)
		opts := []string{
			"Skip — no email notifications",
			"Configure SMTP",
		}
		cursor := 0
		if m.enabled {
			cursor = 1
		}
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header, renderMenu(opts, cursor),
				styles.Help.Render("↑/↓ toggle  •  enter confirm  •  q quit"),
			),
		)

	case 1:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("SMTP Configuration"), "",
		)
		form := ""
		for i, input := range m.inputs {
			lbl := styles.InputLabel
			if smtpField(i) == m.focused {
				lbl = lbl.Foreground(styles.Primary)
			}
			form += lbl.Render(labels[i]) + "\n"
			form += input.View() + "\n\n"
		}
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header, form,
				styles.Help.Render("tab/↑↓ navigate  •  enter next  •  q quit"),
			),
		)

	case 2:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("SMTP Encryption"),
			styles.Subtitle.Render("How should Tidefly connect to your mail server?"),
			"",
		)
		opts := []string{
			"None — no encryption (local/dev only)",
			"STARTTLS — upgrade after connecting (port 587, most providers)",
			"TLS — direct TLS connection (port 465)",
		}
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header, renderMenu(opts, m.tlsCursor),
				styles.Help.Render("↑/↓ navigate  •  enter confirm  •  q quit"),
			),
		)
	}
	return ""
}
