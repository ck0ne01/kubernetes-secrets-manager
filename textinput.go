package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func initialTextInput() textinput.Model {
	width, _, _ := term.GetSize(0)

	ti := textinput.New()
	ti.Placeholder = "my-secret"
	ti.CharLimit = 156
	ti.Width = width - 1
	ti.Focus()
	return ti
}

func updateInputView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textinput.Focused() {
				m.textinput.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			m.secretName = m.textinput.Value()
			m.state = texteditView
			m.textarea = initialTextArea("")
			return m, nil
		default:
			if !m.textinput.Focused() {
				cmd = m.textinput.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errorMsg:
		m.err = msg
		return m, nil
	}

	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
