package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary = lipgloss.Color("#7C3AED")
	Success = lipgloss.Color("#10B981")
	Warning = lipgloss.Color("#F59E0B")
	Danger  = lipgloss.Color("#EF4444")
	Muted   = lipgloss.Color("#6B7280")
	White   = lipgloss.Color("#F9FAFB")

	Title = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true).
		MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	MenuItem = lipgloss.NewStyle().
			Foreground(White).
			PaddingLeft(2)

	MenuItemSelected = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true).
				PaddingLeft(1).
				SetString("› ")

	StatusOK   = lipgloss.NewStyle().Foreground(Success).Bold(true)
	StatusErr  = lipgloss.NewStyle().Foreground(Danger).Bold(true)
	StatusWarn = lipgloss.NewStyle().Foreground(Warning).Bold(true)

	InputLabel = lipgloss.NewStyle().
			Foreground(White).
			Bold(true).
			MarginTop(1)

	Help = lipgloss.NewStyle().Foreground(Muted)

	// Banner – die drei Farb-Ebenen spiegeln die originalen ANSI-Codes wider:
	//   99 = lila  (Wordmark)
	//   45 = cyan  (Punkte-Reihen)
	//   51 = helles Cyan (Tagline)
	bannerWordmarkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	bannerDotsStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	bannerTaglineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))

	Banner = lipgloss.JoinVertical(
		lipgloss.Center,
		bannerWordmarkStyle.Render(
			"████████╗██╗██████╗ ███████╗███████╗██╗     ██╗   ██╗\n"+
				"╚══██╔══╝██║██╔══██╗██╔════╝██╔════╝██║     ╚██╗ ██╔╝\n"+
				"   ██║   ██║██║  ██║█████╗  █████╗  ██║      ╚████╔╝ \n"+
				"   ██║   ██║██║  ██║██╔══╝  ██╔══╝  ██║       ╚██╔╝  \n"+
				"   ██║   ██║██████╔╝███████╗██║     ███████╗   ██║   \n"+
				"   ╚═╝   ╚═╝╚═════╝ ╚══════╝╚═╝     ╚══════╝   ╚═╝   ",
		),
	)
)

// Frame rendert den äußeren Container mit Banner oben + Page-Inhalt darunter.
func Frame(width, height int, content string) string {
	innerWidth := width - 2 - 6 // border:2 + padding-left:3 + padding-right:3
	innerHeight := height - 4   // border:2 + padding-top:1 + padding-bottom:1
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		Banner,
		lipgloss.NewStyle().Foreground(Muted).Render("─────────────────────────────────────────────────────"),
		"",
		content,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(1, 3).
		Width(innerWidth).
		Height(innerHeight).
		Render(body)
}
