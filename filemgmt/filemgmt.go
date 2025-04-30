package filemgmt

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	downloadDir := filepath.Join(homeDir, "Documents", "tracker-tui")
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

func returnProperLength(record string, column int) int {
	if column == 0 || column == 1 {
		return 25
	}
	for i := 1; i < len(record); i++ {
		if record[i] == '(' {
			return i
		}
		if isLower(record[i-1]) && isUpper(record[i]) {
			return i
		}
		if isUpper(record[i-1]) && isLower(record[i]) && i > 5 {
			return i - 1
		}
	}
	return len(record)
}

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func isUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}
