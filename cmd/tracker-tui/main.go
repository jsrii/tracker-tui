package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"tracker-tui/filemgmt"
	"tracker-tui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
)

type errMsg struct{ err error }
type tickMsg time.Time

type audioReadyMsg struct {
	stream beep.StreamSeekCloser
	format beep.Format
}

type model struct {
	termWidth    int
	termHeight   int
	headerStyles lipgloss.Style

	sheetInput textinput.Model

	artistChosen bool
	menuFocus    string // "start", "sheetInput", "list"
	menuChoice   int
	csvList      list.Model
	listItems    []list.Item
	selected     map[int]struct{}
	csvChosen    string

	mainCSVTable   table.Model
	erasTable      table.Model
	columns        []table.Column
	rows           []table.Row
	mainColumns    []table.Column
	mainRows       []table.Row
	erasColumns    []table.Column
	erasRows       []table.Row
	selectedLink   string
	selectedSong   table.Row
	csvTableState  bool
	isPlaying      bool
	decodedFile    beep.StreamSeekCloser
	fileFormat     beep.Format
	tableWidth     int
	controlState   bool
	pControlSelect int
	songProgress   progress.Model
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

func initialModel() model {
	emptyRow := table.Row{"No Song Currently Selected", "", "", "", ""}
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
	filesList.Title = "Browsing " + osHomeDir + "/Documents/tracker-tui/csv/"
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

	mainCSVTableAdditionalStyles := table.DefaultStyles()
	mainCSVTableAdditionalStyles.Selected = styles.CsvTableSelectedStyle

	mainCSVTable := table.New()
	mainCSVTable.SetStyles(mainCSVTableAdditionalStyles)

	erasTable := table.New()
	erasTable.SetStyles(mainCSVTableAdditionalStyles)

	return model{
		selected:       make(map[int]struct{}),
		artistChosen:   false,
		sheetInput:     sheetInput,
		headerStyles:   styles.Header,
		csvList:        filesList,
		listItems:      items,
		menuFocus:      "start", // <<<<<< start screen
		menuChoice:     0,
		mainCSVTable:   mainCSVTable,
		erasTable:      erasTable,
		selectedLink:   "Not Selected yet",
		csvTableState:  false,
		isPlaying:      false,
		selectedSong:   emptyRow,
		tableWidth:     44,
		pControlSelect: 1,
		controlState:   true,
		songProgress:   progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		if m.isPlaying && m.decodedFile != nil {
			cmd = m.songProgress.SetPercent(float64(m.decodedFile.Position()) / float64(m.decodedFile.Len()))
			return m, tea.Batch(cmd, tick())
		}
		return m, tick() // keep ticking even if paused
	case progress.FrameMsg:
		progressModel, cmd := m.songProgress.Update(msg)
		m.songProgress = progressModel.(progress.Model)
		return m, cmd

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
		termWidth, termHeight, _ := term.GetSize(os.Stdout.Fd())
		m.termWidth = termWidth
		m.termHeight = termHeight
		m.csvList.SetSize(termWidth, termHeight-4)
		m.mainCSVTable.SetHeight(termHeight - 3)
		m.erasTable.SetHeight(termHeight - 3)
		if m.decodedFile != nil {
			fmt.Print(float64(m.decodedFile.Position()) / float64(m.decodedFile.Len()))
		}
		return m, nil

	case errMsg:
		fmt.Println("Error:", msg.err)
		return m, nil

	case audioReadyMsg:
		// speaker setup
		if m.isPlaying {
			speaker.Clear()
		}

		m.decodedFile = msg.stream
		m.fileFormat = msg.format
		m.isPlaying = true

		go func(stream beep.StreamSeekCloser, format beep.Format) {
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			done := make(chan bool)

			speaker.Play(beep.Seq(stream, beep.Callback(func() {
				done <- true
			})))

			<-done
		}(msg.stream, msg.format)

		cmd := m.songProgress.SetPercent(float64(m.decodedFile.Position()) / float64(m.decodedFile.Len()))
		return m, tea.Batch(cmd, tick())
	}

	return m, cmd
}

func (m model) View() string {
	var s string

	if m.termWidth < 140 {
		requiredWidth := 140

		// Styles
		highlight := lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
		centered := lipgloss.NewStyle().Align(lipgloss.Center).Width(50)

		// Compose message
		msg := lipgloss.JoinVertical(lipgloss.Top,
			"",
			centered.Render("Terminal size too small:"),
			centered.Render(fmt.Sprintf("  Width = %s",
				highlight.Render(fmt.Sprint(m.termWidth)))),
			"",
			centered.Render("Needed for current config:"),
			centered.Render(fmt.Sprintf("  Width = %d", requiredWidth)),
		)

		// Get terminal size

		// Frame message to center
		s += lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, msg)
		return s
	}
	switch m.artistChosen {
	case true:
		s = styles.Header.Width(m.termWidth).Render("tracker-tui")
		songName := lipgloss.NewStyle().Foreground(lipgloss.Color("#c4746e")).Height(3).Foreground(lipgloss.Color("#c4746e")).MarginBottom(2).AlignVertical(lipgloss.Center).PaddingLeft(1).PaddingRight(1).Render(filemgmt.FormatTitle(m.selectedSong[0]))
		artist := lipgloss.NewStyle().MarginBottom(1).Render(strings.Split(m.csvChosen, ".csv")[0])
		prev := m.renderButton("<< prev", 0, m.controlState)
		playPause := m.renderButton("play/pause", 1, m.controlState)
		skip := m.renderButton("skip >>", 2, m.controlState)
		playButtons := lipgloss.JoinHorizontal(lipgloss.Center, prev, playPause, skip)
		var link string
		if m.selectedLink == "Not Selected yet" {
			link = m.selectedLink
		} else {
			link = "file from: https://" + strings.Split(strings.Split(m.selectedLink, "https://")[1], "/")[0]
		}
		songProgression := lipgloss.NewStyle().MarginBottom(1).AlignHorizontal(lipgloss.Center).Render(m.songProgress.View())
		link = lipgloss.NewStyle().MarginTop(1).Render(link)
		player := lipgloss.JoinVertical(lipgloss.Center, songName, artist, songProgression, playButtons, link)

		if m.csvTableState {
			s += lipgloss.JoinHorizontal(lipgloss.Center, "\n"+styles.CsvTableBaseStyle.Height(m.termHeight-3).Render(m.erasTable.View()), lipgloss.NewStyle().Width(m.termWidth-m.tableWidth-9).Height(m.termHeight-1).AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center).Render("\n"+player))
		} else {
			s += lipgloss.JoinHorizontal(lipgloss.Center, "\n"+styles.CsvTableBaseStyle.Height(m.termHeight-3).Render(m.mainCSVTable.View()), lipgloss.NewStyle().Width(m.termWidth-m.tableWidth-20).Height(m.termHeight-1).AlignVertical(lipgloss.Center).AlignHorizontal(lipgloss.Center).Render("\n"+player))

		}
	case false:
		switch m.menuFocus {
		case "start":
			headerLogo := styles.Header.Render("   __                  __                   __        _ \n  / /__________ ______/ /_____  _____      / /___  __(_)\n / __/ ___/ __ `/ ___/ //_/ _ \\/ ___/_____/ __/ / / / / \n/ /_/ /  / /_/ / /__/ ,< /  __/ /  /_____/ /_/ /_/ / /  \n\\__/_/   \\__,_/\\___/_/|_|\\___/_/         \\__/\\__,_/_/   \n                                                        ")
			var okButton string
			var cancelButton string
			if m.menuChoice == 0 {
				okButton = styles.ActiveButtonStyle.MarginRight(3).Render("Yes (Add new link)")
				cancelButton = styles.ButtonStyle.Render("No (Browse)")
			} else {
				okButton = styles.ButtonStyle.MarginRight(3).Render("Yes (Add new link)")
				cancelButton = styles.ActiveButtonStyle.Render("No (Browse)")
			}
			msg := lipgloss.JoinVertical(lipgloss.Top,
				lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(m.termWidth).Render(headerLogo),
				styles.TextStyling.Align(lipgloss.Center).Width(m.termWidth).Render("Enter in new Sheet Tracker link or browse downloaded trackers"),
				lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Width(m.termWidth).Render(lipgloss.JoinHorizontal(lipgloss.Center, okButton, cancelButton)),
			)

			s += lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, msg)

		case "sheetInput":
			s = styles.Header.Width(m.termWidth).Render("tracker-tui")
			s += styles.TextStyling.Width(m.termWidth).Render("\nEnter the link to the Google Sheet Tracker:\n\n", m.sheetInput.View()+"\n\n")
		case "list":
			s = styles.Header.Width(m.termWidth).Render("tracker-tui")
			s += styles.DocStyle.Render(m.csvList.View())
		}
	}
	return s
}

func (m model) renderButton(label string, index int, parentState bool) string {
	bg := lipgloss.Color("#232323")
	fg := lipgloss.Color("#c5c9c5")
	if m.pControlSelect == index && parentState {
		bg = lipgloss.Color("#434343")
		fg = lipgloss.Color("#c5c9c5")
	} else if m.pControlSelect == index && !parentState {
		bg = lipgloss.Color("#87a987")
		fg = lipgloss.Color("#232323")
	}
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Padding(1).
		MarginRight(1).
		Render(label)
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
