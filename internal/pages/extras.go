package pages

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

type extrasOption struct {
	label   string
	desc    string
	enabled bool
}

type ExtrasModel struct {
	cfg     SetupConfig
	cursor  int
	options []extrasOption
}

func NewExtras(cfg SetupConfig) *ExtrasModel {
	return &ExtrasModel{
		cfg: cfg,
		options: []extrasOption{
			{
				label:   "CrowdSec + Caddy Bouncer",
				desc:    "Automatic IP blocking via community threat intelligence",
				enabled: true,
			},
			{
				label:   "Fail2ban",
				desc:    "Blocks brute-force attacks on SSH and manifest",
				enabled: true,
			},
			{
				label:   "OWASP / Coraza WAF",
				desc:    "Web application firewall via Coraza (built into Caddy)",
				enabled: false,
			},
		},
	}
}

func (m *ExtrasModel) Init() tea.Cmd { return nil }

func (m *ExtrasModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(keyMsg, keys.Down):
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case key.Matches(keyMsg, keys.Space):
			m.options[m.cursor].enabled = !m.options[m.cursor].enabled
		case key.Matches(keyMsg, keys.Enter):
			cfg := m.cfg
			cfg.HardenCrowdSec = m.options[0].enabled
			cfg.HardenFail2ban = m.options[1].enabled
			cfg.HardenCoraza = m.options[2].enabled
			return m, func() tea.Msg {
				return NavigateTo{Page: PageStart, Config: cfg}
			}
		case key.Matches(keyMsg, keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ExtrasModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Security Hardening"),
		styles.Subtitle.Render("Optional — can be configured later"),
		"",
	)

	warning := lipgloss.NewStyle().
		Foreground(styles.StatusWarn.GetForeground()).
		Render("⚠  You are responsible for your own security configuration.\n"+
			"   These are sensible defaults — review and adjust for your environment.") + "\n\n"

	list := ""
	for i, opt := range m.options {
		checkbox := "[ ]"
		if opt.enabled {
			checkbox = styles.StatusOK.Render("[✓]")
		}
		label := opt.label
		if i == m.cursor {
			label = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render(opt.label)
			checkbox = lipgloss.NewStyle().Foreground(styles.Primary).Render(checkbox)
			if !opt.enabled {
				checkbox = lipgloss.NewStyle().Foreground(styles.Primary).Render("[ ]")
			}
		}
		desc := lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(6).Render(opt.desc)
		list += "  " + checkbox + "  " + label + "\n" + desc + "\n\n"
	}

	help := styles.Help.Render("↑/↓ navigate  •  space toggle  •  enter continue  •  q quit")

	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, warning, list, help,
		),
	)
}
