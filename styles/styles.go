package styles

import (
	"github.com/charmbracelet/lipgloss"
)

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
		Foreground(lipgloss.Color("#c4746e")).
		Background(lipgloss.Color("#232323")).
		PaddingTop(0).
		PaddingRight(1).
		Align(lipgloss.Center)

	DocStyle = lipgloss.NewStyle().
			Padding(2, 0)

	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	ButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#232323")).
			Margin(1, 2, 0, 1).
			Padding(0, 2)

	ActiveButtonStyle = ButtonStyle.
				Foreground(lipgloss.Color("#232323")).
				Background(lipgloss.Color("#8a9a7b")).
				Underline(true).
				Margin(1, 2, 0, 1).
				Padding(0, 2)

	TextStyling = lipgloss.NewStyle().
			Align(lipgloss.Left).
			Foreground(lipgloss.Color("#FAFAFA")).
			Margin(0, 0, 0, 0).
			Padding(1, 2)

	ListTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#232323")).
			Background(lipgloss.Color("#8ba4b0")).
			Padding(0, 1)

	ListStatusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render

	ListSelection = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8a9a7b")).
			Border(tabBorder, false).
			BorderForeground(lipgloss.Color("#8e8a7b")).
			BorderLeft(true).
			PaddingLeft(1)
)
