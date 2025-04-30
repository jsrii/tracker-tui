package main

import (
	"fmt"
	"os"
	"tracker-tui/filemgmt"
	"tracker-tui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

type model struct {
	termWidth    int
	termHeight   int
	headerStyles lipgloss.Style

	sheetInput textinput.Model

	artistChosen bool
	menuFocus    string // "start", "sheetInput", "list"
	menuChoice   int

	csvList   list.Model
	listItems []list.Item
	selected  map[int]struct{}
	csvChosen string

	csvTable table.Model
	columns  []table.Column
	rows     []table.Row
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

	loadingSpinner := spinner.New()
	loadingSpinner.Spinner = spinner.Dot
	loadingSpinner.Tick()

	sheetInput := textinput.New()
	sheetInput.Placeholder = "https://docs.google.com/spreadsheets/d/Sheet_ID/htmlview?gid=some_gid#gid=some_gid"
	sheetInput.CharLimit = 200
	sheetInput.Width = 81

	filesListAdditionalStyles := list.NewDefaultDelegate()
	filesListAdditionalStyles.Styles.SelectedTitle = styles.ListSelection
	filesListAdditionalStyles.Styles.SelectedDesc = styles.ListSelection

	filesList := list.New(items, filesListAdditionalStyles, 0, 0)
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

	csvTableAdditionalStyles := table.DefaultStyles()
	csvTableAdditionalStyles.Selected = styles.CsvTableSelectedStyle
	csvTable := table.New()
	csvTable.SetStyles(csvTableAdditionalStyles)

	return model{
		selected:     make(map[int]struct{}),
		artistChosen: false,
		sheetInput:   sheetInput,
		headerStyles: styles.Header,
		csvList:      filesList,
		listItems:    items,
		menuFocus:    "start", // <<<<<< start screen
		menuChoice:   0,
		csvTable:     csvTable,
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
		m.csvList.SetSize(termWidth, termHeight-4)
		m.csvTable.SetHeight(termHeight - 3)
	}
	return m, cmd
}

func (m model) View() string {
	s := styles.Header.Width(m.termWidth).Render("tracker-tui")

	switch m.artistChosen {
	case true:
		s += lipgloss.JoinHorizontal(lipgloss.Top, "\n"+styles.CsvTableBaseStyle.Render(m.csvTable.View()), "\n"+lipgloss.NewStyle().Width(m.termWidth-m.csvTable.Width()).Height(m.termHeight-1).Foreground(lipgloss.Color("#cdcdcd")).Background(lipgloss.Color("#8a9a7b")).Render("Music Player"))
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
			s += styles.DocStyle.Render(m.csvList.View())
		}
	}
	return s
}
