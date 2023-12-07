package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Quit key.Binding
	Save key.Binding
}

var textareaKeys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
}

var textinputKeys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Save: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Save, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Save, k.Quit},
	}
}

func newHelpModel() help.Model {
	return help.New()
}
