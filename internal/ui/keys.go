package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit          key.Binding
	Tab           key.Binding
	BackTab       key.Binding
	Refresh       key.Binding
	Open          key.Binding
	OpenInspector key.Binding
	Logs          key.Binding
	Details       key.Binding
	Up            key.Binding
	Down          key.Binding
	PageUp        key.Binding
	PageDown      key.Binding
}

var keys = keyMap{
	Quit:          key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Tab:           key.NewBinding(key.WithKeys("tab")),
	BackTab:       key.NewBinding(key.WithKeys("shift+tab")),
	Refresh:       key.NewBinding(key.WithKeys("r")),
	Open:          key.NewBinding(key.WithKeys("o")),
	OpenInspector: key.NewBinding(key.WithKeys("i")),
	Logs:          key.NewBinding(key.WithKeys("l")),
	Details:       key.NewBinding(key.WithKeys("d")),
	Up:            key.NewBinding(key.WithKeys("up", "k")),
	Down:          key.NewBinding(key.WithKeys("down", "j")),
	PageUp:        key.NewBinding(key.WithKeys("pgup", "ctrl+u")),
	PageDown:      key.NewBinding(key.WithKeys("pgdown", "ctrl+d")),
}
