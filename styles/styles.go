package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Define all color constants in one place
var (
	ColorPrimary       = lipgloss.Color("#c4746e")
	ColorBackground    = lipgloss.Color("#232323")
	ColorText          = lipgloss.Color("#c5c9c5")
	ColorAccent        = lipgloss.Color("#8a9a7b")
	ColorHighlight     = lipgloss.Color("#8ba4b0")
	ColorDialogBorder  = lipgloss.Color("#874BFD")
	ColorTableBorder   = lipgloss.Color("240")
	ColorSelectedText  = lipgloss.Color("#131313")
	ColorAltText       = lipgloss.Color("#c5c9c5")
	ColorAltBackground = lipgloss.Color("#232323")
	ColorListSelection = lipgloss.Color("#8a9a7b")
	ColorListTitleFg   = lipgloss.Color("#232323")

	ColorActiveSelectedBtnFG = lipgloss.Color("#232323")
	ColorActiveSelectedBtnBG = lipgloss.Color("#87a987")

	ColorActiveUnselectedBtnFG = lipgloss.Color("#c5c9c5")
	ColorActiveUnselectedBtnBG = lipgloss.Color("#232323")

	ColorAltSelectedBtnFG = lipgloss.Color("#c5c9c5")
	ColorAltSelectedBtnBG = lipgloss.Color("#434343")
)

// Define styles
var (
	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Background(ColorBackground).
		PaddingTop(0).
		PaddingRight(0).
		PaddingLeft(1).
		Align(lipgloss.Center)

	DocStyle = lipgloss.NewStyle().
			Padding(2, 0)

	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDialogBorder).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	ButtonStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBackground).
			Margin(1, 2, 0, 1).
			AlignHorizontal(lipgloss.Center).
			Padding(0, 2)

	ActiveButtonStyle = ButtonStyle.
				Foreground(ColorBackground).
				Background(ColorAccent).
				Underline(true)

	TextStyling = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(ColorText).
			Padding(1, 2)

	ListTitle = lipgloss.NewStyle().
			Foreground(ColorListTitleFg).
			Background(ColorHighlight).
			Padding(0, 1)

	ListSelection = lipgloss.NewStyle().
			Foreground(ColorListSelection).
			Border(tabBorder, false).
			BorderForeground(ColorListSelection).
			BorderLeft(true).
			PaddingLeft(1)

	CsvTableBaseStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorTableBorder)

	CsvTableSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorSelectedText).
				Background(ColorAccent).
				Bold(false)

	CsvTableSelectedStyleAlt = lipgloss.NewStyle().
					Foreground(ColorAltText).
					Background(ColorAltBackground).
					Bold(false)
)
