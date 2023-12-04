package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh/terminal"
	v1 "k8s.io/api/core/v1"
)

func main() {
	// TODO: reactivate fullscreen
	// if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type (
	state    int
	errorMsg error
)

type model struct {
	secrets    *v1.SecretList
	secretName string
	secretData string
	quitting   bool
	err        error
	list       list.Model
	textarea   textarea.Model
	state      state
}

const (
	initialList state = iota
	namespacesList
	secretsList
	texteditView
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	switch m.state {
	case initialList, namespacesList, secretsList:
		return updateList(msg, m)
	case texteditView:
		return updateEditView(msg, m)
	}
	return updateList(msg, m)
}

func updateList(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case namespacesItemsMsg:
		m.list.SetItems(msg)

	case secretsMsg:
		m.secrets = msg
		secretsListItems := secretNamesToListItems(msg)
		m.list.SetItems(secretsListItems)

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			switch state := m.state; state {
			case initialList:
				m.state = namespacesList
				m.list.Title = "Choose a namespace"
				return m, handleNamespaceQuery()
			case namespacesList:
				var namespace string
				i, ok := m.list.SelectedItem().(item)
				if ok {
					namespace = string(i)
					m.list.Title = "Choose a secret"
					m.state = secretsList
				}
				return m, handleSecretsQuery(namespace)
			case secretsList:
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.secretName = string(i)
					m.state = texteditView
					secretData := getSecretData(m.secrets, m.secretName)
					fmt.Println(secretData)
					ti := textarea.New()
					ti.SetValue(secretData)
					width, _, _ := terminal.GetSize(0)
					ti.SetWidth(width - 1)
					ti.Focus()
					m.textarea = ti
				}
				return m, textarea.Blink
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func handleNamespaceQuery() tea.Cmd {
	return func() tea.Msg {
		namespaces, err := fetchNamespaces()
		if err != nil {
			return errorMsg(err)
		}
		return namespacesToListItems(namespaces)
	}
}

func handleSecretsQuery(namespace string) tea.Cmd {
	return func() tea.Msg {
		secrets, err := fetchSecrets(namespace)
		if err != nil {
			return errorMsg(err)
		}
		return secretsMsg(secrets)
	}
}

func updateEditView(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
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

func initialModel() model {
	initialItems := []list.Item{
		item("Create a new secret"),
		item("Update an existing secret"),
	}

	const defaultWidth = 20

	l := list.New(initialItems, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Choose an action"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model{list: l, state: initialList}
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("See you later!")
	}
	switch m.state {
	case initialList, namespacesList, secretsList:
		return m.list.View()
	case texteditView:
		return fmt.Sprintf(
			"Enter Secret Data.\n\n%s\n\n%s",
			m.textarea.View(),
			"(ctrl+c to quit, ctrl+s to save)",
		) + "\n\n"
	}
	return "no view found"
}
