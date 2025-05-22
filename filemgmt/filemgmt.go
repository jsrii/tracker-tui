package filemgmt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"tracker-tui/styles"

	list "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ColorPrimary               string
	ColorBackground            string
	ColorText                  string
	ColorAccent                string
	ColorHighlight             string
	ColorDialogBorder          string
	ColorTableBorder           string
	ColorSelectedText          string
	ColorStatusMessage         map[string]string
	ColorAltText               string
	ColorAltBackground         string
	ColorListSelection         string
	ColorListTitleFg           string
	ColorActiveSelectedBtnFG   string
	ColorActiveSelectedBtnBG   string
	ColorActiveUnselectedBtnFG string
	ColorActiveUnselectedBtnBG string
	ColorAltSelectedBtnFG      string
	ColorAltSelectedBtnBG      string
}
type item struct {
	fileName, dateOfCreation string
}

func (i item) Title() string       { return i.fileName }
func (i item) Description() string { return i.dateOfCreation }
func (i item) FilterValue() string { return i.fileName }

func ReturnListOfFiles() ([]list.Item, error) {
	var items []list.Item

	homeDir, _ := os.UserHomeDir()
	downloadDir := filepath.Join(homeDir, "Documents", "tracker-tui", "csv")
	err := os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	directory, err := os.ReadDir(downloadDir)
	if err != nil {
		return nil, err
	}

	for _, file := range directory {
		info, err := file.Info()
		if err != nil {
			continue
		}
		modTime := info.ModTime().Format("2006/01/02")

		items = append(items, item{
			fileName:       file.Name(),
			dateOfCreation: modTime,
		})
	}

	return items, nil
}

func ReadCSVFile(filename string) ([]table.Column, []table.Row, error) {
	var columns []table.Column
	var rows []table.Row
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	reader := csv.NewReader(f)
	var row int = 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error reading CSV data:", err)
			break
		}

		if row == 0 {
			for i := range record {
				columns = append(columns, table.Column{Title: record[i], Width: returnProperLength(record[i], i)})
			}
		} else {
			var subRow table.Row
			for i := range record {
				subRow = append(subRow, record[i])
			}
			rows = append(rows, subRow)
		}

		fmt.Println(record)
		row++
	}

	return columns, rows, nil
}

func GenerateMainTable(largeColumns []table.Column, largeRows []table.Row) ([]table.Column, []table.Row, error) {
	var columns []table.Column
	var rows []table.Row

	columns = append(columns, table.Column{Title: "Files in Era", Width: 25})
	columns = append(columns, table.Column{Title: "Name of Era", Width: 35})

	for i := range largeRows {
		if largeRows[i][len(largeRows[i])-2] == "" && len(largeRows[i][1]) > 1 && len(largeRows[i][0]) > 1 {
			if len(largeRows[i]) >= 2 {
				rows = append(rows, table.Row{
					largeRows[i][0],
					FormatTitle(largeRows[i][1]),
				})
			}
		}
	}

	return columns, rows, nil
}

func GenerateEraTable(largeColumns []table.Column, largeRows []table.Row, matchString string) ([]table.Column, []table.Row, error) {
	var columns []table.Column
	var rows []table.Row
	caseUpper := strings.ToUpper(FormatTitle(matchString))

	if len(largeColumns) > 1 {
		columns = append(columns, largeColumns[1:]...)
	}

	for i := range largeRows {
		if strings.ToUpper(largeRows[i][0]) == caseUpper {
			if len(largeRows[i]) > 1 {
				rows = append(rows, largeRows[i][1:])
			}
		}
	}

	return columns, rows, nil
}

func returnProperLength(record string, column int) int {
	if column == 0 || column == 1 {
		return 15 + 12*column // this is my favourite line out of all the files :3
	}

	if idx := strings.Index(record, "("); idx != -1 {
		return idx
	}

	return len(record)
}

func FormatTitle(input string) string {
	if idx := strings.Index(input, "("); idx != -1 {
		return strings.TrimSpace(input[:idx])
	}
	return strings.TrimSpace(input)
}

func InitTheme() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, "Documents", "tracker-tui")
	configPath := filepath.Join(configDir, "config.json")

	// Ensure config dir exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultTheme := Theme{
			ColorPrimary:               "#c4746e",
			ColorBackground:            "#232323",
			ColorText:                  "#c5c9c5",
			ColorAccent:                "#8a9a7b",
			ColorHighlight:             "#8ba4b0",
			ColorDialogBorder:          "#874BFD",
			ColorTableBorder:           "240",
			ColorSelectedText:          "#131313",
			ColorAltText:               "#c5c9c5",
			ColorAltBackground:         "#232323",
			ColorListSelection:         "#8a9a7b",
			ColorListTitleFg:           "#232323",
			ColorActiveSelectedBtnFG:   "#232323",
			ColorActiveSelectedBtnBG:   "#87a987",
			ColorActiveUnselectedBtnFG: "#c5c9c5",
			ColorActiveUnselectedBtnBG: "#232323",
			ColorAltSelectedBtnFG:      "#c5c9c5",
			ColorAltSelectedBtnBG:      "#434343",
		}

		data, _ := json.MarshalIndent(defaultTheme, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("could not write default config: %w", err)
		}
	}

	// Read and apply the config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return fmt.Errorf("invalid theme config: %w", err)
	}

	// Set the theme
	styles.ColorPrimary = lipgloss.Color(theme.ColorPrimary)
	styles.ColorBackground = lipgloss.Color(theme.ColorBackground)
	styles.ColorText = lipgloss.Color(theme.ColorText)
	styles.ColorAccent = lipgloss.Color(theme.ColorAccent)
	styles.ColorHighlight = lipgloss.Color(theme.ColorHighlight)
	styles.ColorDialogBorder = lipgloss.Color(theme.ColorDialogBorder)
	styles.ColorTableBorder = lipgloss.Color(theme.ColorTableBorder)
	styles.ColorSelectedText = lipgloss.Color(theme.ColorSelectedText)
	styles.ColorAltText = lipgloss.Color(theme.ColorAltText)
	styles.ColorAltBackground = lipgloss.Color(theme.ColorAltBackground)
	styles.ColorListSelection = lipgloss.Color(theme.ColorListSelection)
	styles.ColorListTitleFg = lipgloss.Color(theme.ColorListTitleFg)
	styles.ColorActiveSelectedBtnFG = lipgloss.Color(theme.ColorActiveSelectedBtnFG)
	styles.ColorActiveSelectedBtnBG = lipgloss.Color(theme.ColorActiveSelectedBtnBG)
	styles.ColorActiveUnselectedBtnFG = lipgloss.Color(theme.ColorActiveUnselectedBtnFG)
	styles.ColorActiveUnselectedBtnBG = lipgloss.Color(theme.ColorActiveUnselectedBtnBG)
	styles.ColorAltSelectedBtnFG = lipgloss.Color(theme.ColorAltSelectedBtnFG)
	styles.ColorAltSelectedBtnBG = lipgloss.Color(theme.ColorAltSelectedBtnBG)

	return nil
}
