package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"k8s.io/client-go/kubernetes"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int {
	return 1
}

func (d itemDelegate) Spacing() int {
	return 0
}

func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
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
					m.state = textinputView
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
