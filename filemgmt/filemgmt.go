package filemgmt

import (
	"os"
	"path/filepath"

	list "github.com/charmbracelet/bubbles/list"
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
