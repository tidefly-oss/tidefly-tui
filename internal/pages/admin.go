package pages

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/tidefly-oss/tidefly-tui/internal/env"
	"github.com/tidefly-oss/tidefly-tui/internal/styles"
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, 64*1024, 3, 2,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

type adminField int

const (
	fieldFirstName adminField = iota
	fieldLastName
	fieldEmail
	fieldPassword
	fieldCount
)

var adminLabels = [fieldCount]string{"First name", "Last name", "Email", "Password"}

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

func (m *AdminModel) Init() tea.Cmd { return textinput.Blink }

func (m *AdminModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AdminCreated:
		m.loading = false
		return m, navigate(PageDone, SetupConfig{})
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
		dbURL := env.GetOrLoad("DATABASE_URL")
		if dbURL == "" {
			return AdminError{"DATABASE_URL not set — is the backend running?"}
		}
		db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err != nil {
			return AdminError{fmt.Sprintf("database connection failed: %v", err)}
		}
		hash, err := hashPassword(password)
		if err != nil {
			return AdminError{fmt.Sprintf("password hashing failed: %v", err)}
		}
		type User struct {
			ID       string
			Email    string
			Password string
			Name     string
			Role     string
			Active   bool
		}
		var count int64
		db.WithContext(context.Background()).
			Table("users").Where("email = ?", email).Count(&count)
		if count > 0 {
			return AdminError{"a user with this email already exists"}
		}
		if err := db.WithContext(context.Background()).Table("users").Create(
			&User{
				ID:       uuid.New().String(),
				Email:    email,
				Password: hash,
				Name:     firstName + " " + lastName,
				Role:     "admin",
				Active:   true,
			},
		).Error; err != nil {
			return AdminError{fmt.Sprintf("failed to create user: %v", err)}
		}
		return AdminCreated{}
	}
}

func (m *AdminModel) View() string {
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Title.Render("Admin Account"),
		styles.Subtitle.Render("Create your first administrator account"),
		"",
	)
	form := ""
	for i, input := range m.inputs {
		lbl := styles.InputLabel
		if adminField(i) == m.focused {
			lbl = lbl.Foreground(styles.Primary)
		}
		form += lbl.Render(adminLabels[i]) + "\n"
		form += input.View() + "\n\n"
	}
	errMsg := ""
	if m.err != "" {
		errMsg = styles.StatusErr.Render("✗ "+m.err) + "\n\n"
	}
	loading := ""
	if m.loading {
		loading = styles.StatusWarn.Render("⟳ creating account...") + "\n\n"
	}
	return styles.Frame(
		termWidth, termHeight, lipgloss.JoinVertical(
			lipgloss.Left,
			header, form, errMsg, loading,
			styles.Help.Render("tab/↑↓ between fields  •  enter confirm  •  q quit"),
		),
	)
}
