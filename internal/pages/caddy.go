package pages

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type caddyStep int

const (
	caddyStepToggle caddyStep = iota
	caddyStepDomain
	caddyStepEmail
	caddyStepStaging
)

type CaddyModel struct {
	step    caddyStep
	enabled bool
	cursor  int

	domainInput textinput.Model
	emailInput  textinput.Model

	cfg SetupConfig
}

func NewCaddy(cfg SetupConfig) *CaddyModel {
	domain := textinput.New()
	domain.Placeholder = "apps.example.com"
	domain.CharLimit = 253

	email := textinput.New()
	email.Placeholder = "admin@example.com"
	email.CharLimit = 255

	return &CaddyModel{
		step:        caddyStepToggle,
		domainInput: domain,
		emailInput:  email,
		cfg:         cfg,
	}
}

func (m *CaddyModel) Init() tea.Cmd { return textinput.Blink }

func (m *CaddyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch m.step {

		case caddyStepToggle:
			switch {
			case key.Matches(keyMsg, keys.Up), key.Matches(keyMsg, keys.Down):
				m.enabled = !m.enabled
			case key.Matches(keyMsg, keys.Enter):
				if !m.enabled {
					m.cfg.CaddyEnabled = false
					cfg := m.cfg
					return m, func() tea.Msg {
						return NavigateTo{Page: PageSMTP, Config: cfg}
					}
				}
				m.step = caddyStepDomain
				m.domainInput.Focus()
				return m, textinput.Blink
			case key.Matches(keyMsg, keys.Quit):
				return m, tea.Quit
			}

		case caddyStepDomain:
			switch {
			case key.Matches(keyMsg, keys.Enter):
				if m.domainInput.Value() == "" {
					return m, nil
				}
				m.step = caddyStepEmail
				m.domainInput.Blur()
				m.emailInput.Focus()
				return m, textinput.Blink
			case key.Matches(keyMsg, keys.Quit):
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.domainInput, cmd = m.domainInput.Update(msg)
				return m, cmd
			}

		case caddyStepEmail:
			switch {
			case key.Matches(keyMsg, keys.Enter):
				if m.emailInput.Value() == "" {
					return m, nil
				}
				m.step = caddyStepStaging
				m.emailInput.Blur()
				m.cursor = 1
				return m, nil
			case key.Matches(keyMsg, keys.Quit):
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.emailInput, cmd = m.emailInput.Update(keyMsg)
				return m, cmd
			}

		case caddyStepStaging:
			switch {
			case key.Matches(keyMsg, keys.Up), key.Matches(keyMsg, keys.Down):
				if m.cursor == 0 {
					m.cursor = 1
				} else {
					m.cursor = 0
				}
			case key.Matches(keyMsg, keys.Enter):
				m.cfg.CaddyEnabled = true
				m.cfg.CaddyDomain = m.domainInput.Value()
				m.cfg.CaddyEmail = m.emailInput.Value()
				m.cfg.CaddyStaging = m.cursor == 0
				cfg := m.cfg
				return m, func() tea.Msg {
					return NavigateTo{Page: PageSMTP, Config: cfg}
				}
			case key.Matches(keyMsg, keys.Quit):
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *CaddyModel) View() string {
	switch m.step {

	case caddyStepToggle:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("Caddy / SSL"),
			styles.Subtitle.Render("Automatically expose services with HTTPS via Let's Encrypt?"),
			"",
			lipgloss.NewStyle().Foreground(styles.Muted).Render(
				"  Requires: DNS wildcard *.yourdomain.com → server IP\n"+
					"  Ports 80 + 443 must be reachable from the internet",
			),
			"",
		)
		yes := "   Enable Caddy + SSL"
		no := "   Skip — no automatic HTTPS"
		if m.enabled {
			yes = styles.MenuItemSelected.Render("") +
				lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("Enable Caddy + SSL")
		} else {
			no = styles.MenuItemSelected.Render("") +
				lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("Skip — no automatic HTTPS")
		}
		help := styles.Help.Render("↑/↓ toggle  •  enter confirm  •  q quit")
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header, yes+"\n\n", no+"\n\n", help,
			),
		)

	case caddyStepDomain:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("Base Domain"),
			styles.Subtitle.Render("Subdomains will be: {service}.yourdomain.com"),
			"",
			lipgloss.NewStyle().Foreground(styles.Muted).Render(
				"  DNS setup required:\n"+
					"  *.apps.example.com  →  A  →  <server-ip>\n"+
					"   apps.example.com  →  A  →  <server-ip>",
			),
			"",
		)
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				styles.InputLabel.Render("Base domain"),
				m.domainInput.View(),
				"",
				styles.Help.Render("enter confirm  •  q quit"),
			),
		)

	case caddyStepEmail:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("ACME Email"),
			styles.Subtitle.Render("Let's Encrypt will send certificate expiry notices here"),
			"",
		)
		return styles.Frame(
			termWidth, termHeight, lipgloss.JoinVertical(
				lipgloss.Left,
				header,
				styles.InputLabel.Render("Email address"),
				m.emailInput.View(),
				"",
				styles.Help.Render("enter confirm  •  q quit"),
			),
		)

	case caddyStepStaging:
		header := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Title.Render("Let's Encrypt CA"),
			styles.Subtitle.Render("Which Certificate Authority should Caddy use?"),
			"",
		)
		opts := []struct{ label, desc string }{
			{"Staging CA", "For testing — no rate limits, but browser shows untrusted cert"},
			{"Production CA", "Real certificates — recommended for live deployments"},
		}
		list := ""
		for i, o := range opts {
			isSelected := i == m.cursor
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
