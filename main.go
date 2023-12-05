package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/ssh/terminal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
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
	k8sclientset *kubernetes.Clientset
	secrets      *v1.SecretList
	secretName   string
	secretData   string
	quitting     bool
	err          error
	list         list.Model
	textarea     textarea.Model
	textinput    textinput.Model
	state        state
}

const (
	initialList state = iota
	namespacesList
	secretsList
	texteditView
	textInputView
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
	case textInputView:
		return updateInputView(msg, m)
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
				i, ok := m.list.SelectedItem().(item)
				if !ok {
					fmt.Println("Something went wrong, could not process selection.")
					os.Exit(1)
				}
				if i == "Create a new secret" {
					m.state = textInputView
					m.textinput = initialTextInput()
					return m, textinput.Blink
				}
				if i == "Update an existing secret" {
					m.k8sclientset = createClientSet()
					m.state = namespacesList
					m.list.Title = "Choose a namespace"
					return m, handleNamespaceQuery(m.k8sclientset)
				}
				return m, nil
			case namespacesList:
				var namespace string
				i, ok := m.list.SelectedItem().(item)
				if ok {
					namespace = string(i)
					m.list.Title = "Choose a secret"
					m.state = secretsList
				}
				return m, handleSecretsQuery(m.k8sclientset, namespace)
			case secretsList:
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.secretName = string(i)
					m.state = texteditView
					secretData := getSecretData(m.secrets, m.secretName)
					m.textarea = initialTextArea(secretData)
				}
				return m, textarea.Blink
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func initialTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.CharLimit = 156
	ti.Width = 20
	ti.Focus()
	return ti
}

func initialTextArea(secretData string) textarea.Model {
	ti := textarea.New()
	width, _, _ := terminal.GetSize(0)
	ti.SetWidth(width - 1)
	ti.Focus()

	if len(secretData) > 0 {
		ti.SetValue(secretData)
	}

	return ti
}

func handleNamespaceQuery(clientSet *kubernetes.Clientset) tea.Cmd {
	return func() tea.Msg {
		namespaces, err := fetchNamespaces(clientSet)
		if err != nil {
			return errorMsg(err)
		}
		return namespacesToListItems(namespaces)
	}
}

func handleSecretsQuery(clientSet *kubernetes.Clientset, namespace string) tea.Cmd {
	return func() tea.Msg {
		secrets, err := fetchSecrets(clientSet, namespace)
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
		case tea.KeyCtrlS:
			m.secretName = m.textinput.Value()
			m.state = texteditView
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
	case textInputView:
		return fmt.Sprintf(
			"Enter the name of your new secret.\n\n%s\n\n%s",
			m.textinput.View(),
			"(esc to quit)",
		) + "\n"
	}
	return "no view found"
}
