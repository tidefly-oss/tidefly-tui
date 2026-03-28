package pages

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
	Tab   key.Binding
}

var keys = keyMap{
	Up:    key.NewBinding(key.WithKeys("up")),
	Down:  key.NewBinding(key.WithKeys("down")),
	Enter: key.NewBinding(key.WithKeys("enter")),
	Back:  key.NewBinding(key.WithKeys("esc", "backspace")),
	Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Tab:   key.NewBinding(key.WithKeys("tab")),
}
