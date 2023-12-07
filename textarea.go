package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func initialTextArea(secretData string) textarea.Model {
	width, _, _ := term.GetSize(0)

	ti := textarea.New()
	ti.SetWidth(width - 1)
	ti.Focus()

	if len(secretData) > 0 {
		ti.SetValue(secretData)
		return ti
	}

	ti.Placeholder = "password: secret"
	return ti
}

func updateTextarea(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlS:
			textareaValue := m.textarea.Value()
			secretsData := stringToSecretData(&textareaValue)
			return m, tea.Sequence(handleSaveToFile(m.secretName, secretsData), encryptFileWithSops(m.secretName), tea.Quit)
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errorMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func handleSaveToFile(secretName string, secretData secretData) tea.Cmd {
	return func() tea.Msg {
		secretYamlContent, err := createSecretYamlContent(secretName, secretData)
		if err != nil {
			fmt.Println(err)
		}

		return os.WriteFile(fmt.Sprintf("%s.yaml", secretName), secretYamlContent, 0666)
	}
}
