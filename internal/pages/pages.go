package pages

import tea "github.com/charmbracelet/bubbletea"

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	OSLinux        = "linux"
	PodmanSocket   = "/run/user/1000/podman/podman.sock"
)

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
	PageCaddy
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
	Runtime       string
	SocketPath    string
	Environment   string
	WithDashboard bool

	CaddyEnabled bool
	CaddyDomain  string
	CaddyEmail   string
	CaddyStaging bool

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
