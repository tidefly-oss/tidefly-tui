package pages

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/codifystudios/tidefly/tui/internal/styles"
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
	step      int // 0=toggle, 1=fields, 2=tls
	focused   smtpField
	inputs    [smtpFieldCount]textinput.Model
	tlsCursor int // 0=none, 1=starttls, 2=tls
	cfg       SetupConfig
}

func NewSMTP(cfg SetupConfig) *SMTPModel {
	labels := [smtpFieldCount]string{
		"SMTP Host", "Port", "Username", "Password", "From address",
	}
	placeholders := [smtpFieldCount]string{
		"smtp.example.com", "587", "apikey", "", "noreply@example.com",
	}

	var inputs [smtpFieldCount]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 255
		if smtpField(i) == smtpPassword {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		t.Prompt = ""
		_ = labels[i]
		inputs[i] = t
	}
	// sensible defaults
	inputs[smtpPort].SetValue("587")

	return &SMTPModel{
		enabled:   false,
		step:      0,
		inputs:    inputs,
		tlsCursor: 1, // starttls default
		cfg:       cfg,
	}
}

func (m *SMTPModel) Init() tea.Cmd { return textinput.Blink }

func (m *SMTPModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch m.step {

		// ── Step 0: toggle ────────────────────────────────────────────────
		case 0:
			switch {
			case key.Matches(keyMsg, keys.Enter):
				if !m.enabled {
					m.cfg.SMTPEnabled = false
					cfg := m.cfg
					return m, func() tea.Msg {
						return NavigateTo{Page: PageStart, Config: cfg}
					}
				}
				m.step = 1
				m.inputs[0].Focus()
				return m, textinput.Blink
			case key.Matches(keyMsg, keys.Quit):
				return m, tea.Quit
			}

		// ── Step 1: fields ────────────────────────────────────────────────
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
				// Last field — go to TLS step
				m.inputs[m.focused].Blur()
				m.step = 2
				return m, nil
			case key.Matches(keyMsg, keys.Quit):
				return m, tea.Quit
			}

		// ── Step 2: TLS ───────────────────────────────────────────────────
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
				cfg := m.cfg
				return m, func() tea.Msg {
					return NavigateTo{Page: PageStart, Config: cfg}
				}
			case key.Matches(keyMsg, keys.Quit):
				return m, tea.Quit
			}
		}

		// Handle inputs
		if m.step == 1 {
			var cmd tea.Cmd
			m.inputs[m.focused], cmd = m.inputs[m.focused].Update(keyMsg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *SMTPModel) View() string {
	switch m.step {

	case 0:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("Email / SMTP"),
			styles.Subtitle.Render("Configure outgoing mail for notifications and alerts?"),
			"",
			lipgloss.NewStyle().Foreground(styles.Muted).Render(
				"  Works with any SMTP provider:\n"+
					"  Resend · Postmark · Mailgun · your own server",
			),
			"",
		)
		yes := "   Configure SMTP"
		no := "   Skip — no email notifications"
		if m.enabled {
			yes = styles.MenuItemSelected.Render("") +
				lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("Configure SMTP")
		} else {
			no = styles.MenuItemSelected.Render("") +
				lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("Skip — no email notifications")
		}
		help := styles.Help.Render("↑/↓ toggle  •  enter confirm  •  q quit")
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header, yes+"\n\n", no+"\n\n", help,
			),
		)

	case 1:
		labels := [smtpFieldCount]string{
			"SMTP Host", "Port", "Username", "Password", "From address",
		}
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("SMTP Configuration"),
			"",
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
		help := styles.Help.Render("tab/↑↓ navigate  •  enter next  •  q quit")
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header, form, help,
			),
		)

	case 2:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("SMTP Encryption"),
			styles.Subtitle.Render("How should Tidefly connect to your mail server?"),
			"",
		)
		opts := []struct{ label, desc string }{
			{"None", "No encryption — only for local/dev servers"},
			{"STARTTLS", "Upgrade to TLS after connecting — most providers (port 587)"},
			{"TLS", "Direct TLS connection — port 465"},
		}
		list := ""
		for i, o := range opts {
			isSelected := i == m.tlsCursor
			label := o.label
			if isSelected {
				label = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(o.label)
			}
			cur := "  "
			if isSelected {
				cur = styles.MenuItemSelected.Render("")
			}
			desc := lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(3).Render(o.desc)
			list += cur + label + "\n" + desc + "\n\n"
		}
		help := styles.Help.Render("↑/↓ navigate  •  enter confirm  •  q quit")
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header, list, help,
			),
		)
	}
	return ""
}
