package filemgmt

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	list "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
)

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

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func isUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func FormatTitle(input string) string {
	if idx := strings.Index(input, "("); idx != -1 {
		return strings.TrimSpace(input[:idx])
	}
	return strings.TrimSpace(input)
}
