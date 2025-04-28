package main

import (
	"fmt"
	"os"
	"tracker-tui/filemgmt"
	"tracker-tui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

type model struct {
	sheetInput      textinput.Model
	artistChosen    bool
	headerStyles    lipgloss.Style
	list            list.Model
	selected        map[int]struct{}
	menuFocus       string // "start", "sheetInput", "list"
	menuChoice      int
	termWidth       int
	termHeight      int
	downloadingFile bool
	listItems       []list.Item
	csvChosen       string
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func initialModel() model {
	items, _ := filemgmt.ReturnListOfFiles()

	sheetInput := textinput.New()
	sheetInput.Placeholder = "https://docs.google.com/spreadsheets/d/Sheet_ID/htmlview?gid=some_gid#gid=some_gid"
	sheetInput.CharLimit = 200
	sheetInput.Width = 81

	additionalStyles := list.NewDefaultDelegate()
	additionalStyles.Styles.SelectedTitle = styles.ListSelection
	additionalStyles.Styles.SelectedDesc = styles.ListSelection

	filesList := list.New(items, additionalStyles, 0, 0)
	osHomeDir, _ := os.UserHomeDir()
	filesList.Title = "Browsing " + osHomeDir + "/Documents/tracker-tui"
	filesList.Styles.Title = styles.ListTitle
	filesList.KeyMap.Quit.SetEnabled(false)
	filesList.KeyMap.Quit.Unbind()
	filesList.KeyMap.CancelWhileFiltering = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cancel filtering"),
	)
	filesList.KeyMap.ClearFilter = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "clear filtering"),
	)
	filesList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "choose csv"),
			),
			key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "exit to main menu"),
			),
		}
	}

	return model{
		selected:        make(map[int]struct{}),
		artistChosen:    false,
		sheetInput:      sheetInput,
		headerStyles:    styles.Header,
		list:            filesList,
		listItems:       items,
		menuFocus:       "start", // <<<<<< start screen
		menuChoice:      0,
		downloadingFile: false,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.artistChosen {
		case true:
			return playerControls(m, msg)
		case false:
			switch m.menuFocus {
			case "start":
				return startControls(m, msg)

			case "list":
				return listControls(m, msg)

			case "sheetInput":
				return sheetInputControls(m, msg)
			}
		}

	case tea.WindowSizeMsg:
		var termWidth, termHeight, _ = term.GetSize(os.Stdout.Fd())
		m.termWidth = termWidth
		m.termHeight = termHeight
		m.list.SetSize(termWidth, termHeight-4)
	}
	return m, cmd
}

func (m model) View() string {
	s := styles.Header.Width(m.termWidth).Render("tracker-tui")

	switch m.artistChosen {
	case true:
		s += styles.TextStyling.Width(m.termWidth).Render(m.csvChosen)
	case false:
		switch m.menuFocus {
		case "start":
			var okButton string
			var cancelButton string
			if m.menuChoice == 0 {
				okButton = styles.ActiveButtonStyle.MarginRight(3).Render("Yes (Add new link)")
				cancelButton = styles.ButtonStyle.Render("No (Browse)")
			} else {
				okButton = styles.ButtonStyle.MarginRight(3).Render("Yes (Add new link)")
				cancelButton = styles.ActiveButtonStyle.Render("No (Browse)")
			}
			s += styles.TextStyling.Width(m.termWidth).Render("\n\nWould you like to add a new Unreleased Music tracker or browse ones you've already downloaded?")
			s += lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
		case "sheetInput":
			s += styles.TextStyling.Width(m.termWidth).Render("\nEnter the link to the Google Sheet Tracker:\n\n", m.sheetInput.View()+"\n\n")
		case "list":
			s += styles.DocStyle.Render(m.list.View())
		}
	}
	return s
}
