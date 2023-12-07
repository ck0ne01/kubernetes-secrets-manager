package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
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
	help         help.Model
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
	textinputView
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case initialList, namespacesList, secretsList:
		return updateList(msg, m)
	case texteditView:
		return updateTextarea(msg, m)
	case textinputView:
		return updateInputView(msg, m)
	}
	return updateList(msg, m)
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

	h := newHelpModel()

	return model{list: l, help: h, state: initialList}
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
			m.help.View(textareaKeys),
		) + "\n\n"
	case textinputView:
		return fmt.Sprintf(
			"Enter the name of your new secret.\n\n%s\n\n%s",
			m.textinput.View(),
			m.help.View(textinputKeys),
		) + "\n"
	}
	return "no view found"
}
