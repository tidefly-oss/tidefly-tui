package pages

import tea "github.com/charmbracelet/bubbletea"

// Model is the interface all page models must implement.
type Model interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type Page int

const (
	PageHome Page = iota
	PageRuntime
	PageEnvironment
	PageDashboard
	PageTraefik
	PageSMTP
	PageExtras
	PageStart
	PageAdmin
	PageDone
)

type NavigateTo struct {
	Page   Page
	Data   string
	Config SetupConfig
}

type SetupConfig struct {
	Runtime    string
	SocketPath string

	Environment string // "development" | "production"

	WithDashboard bool

	TraefikEnabled bool
	TraefikDomain  string
	TraefikEmail   string
	TraefikStaging bool

	SMTPEnabled  bool
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPTLS      string
}

var (
	termWidth  = 80
	termHeight = 24
)

func SetSize(w, h int) {
	termWidth = w
	termHeight = h
}
