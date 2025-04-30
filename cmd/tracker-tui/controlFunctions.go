package main

import (
	"os"
	"strings"
	"tracker-tui/download"
	"tracker-tui/filemgmt"

	tea "github.com/charmbracelet/bubbletea"
)

func startControls(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "left":
		m.menuChoice = 0
		return m, nil
	case "right":
		m.menuChoice = 1
		return m, nil

	case "enter":
		m.sheetInput.Blur()
		if m.menuChoice == 0 {
			m.menuFocus = "sheetInput"
			m.sheetInput.Focus()
		} else {
			items, _ := filemgmt.ReturnListOfFiles()
			m.csvList.SetItems(items)
			m.menuFocus = "list"
		}
		return m, nil
	}
	return m, nil
}

func listControls(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	var homeDir string
	var readErr error
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.artistChosen = false
		m.sheetInput.SetValue("")
		m.menuFocus = "start"
		m.menuChoice = 0
		return m, nil
	case "enter":
		clear(m.columns)
		clear(m.rows)
		index := m.csvList.Index()
		if _, ok := m.selected[index]; ok {
			delete(m.selected, index)
		} else {
			m.selected[index] = struct{}{}
		}
		m.csvChosen = m.csvList.SelectedItem().FilterValue()
		m.artistChosen = true
		m.sheetInput.SetValue("")
		m.menuFocus = "start"
		m.menuChoice = 0

		homeDir, _ = os.UserHomeDir()

		m.columns, m.rows, readErr = filemgmt.ReadCSVFile(homeDir + "/Documents/tracker-tui/" + m.csvChosen)
		if readErr != nil {
			return m, tea.Quit
		}

		m.csvTable.SetColumns(m.columns)
		m.csvTable.SetRows(m.rows)
		m.csvTable.Focus()
		return m, nil
	}
	var cmd tea.Cmd
	m.csvList, cmd = m.csvList.Update(msg)
	return m, cmd
}

func sheetInputControls(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var homeDir string
	var cmd tea.Cmd
	var downloadErr error
	var readErr error
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.artistChosen = false
		m.sheetInput.SetValue("")
		m.menuFocus = "start"
		m.menuChoice = 0
		return m, nil

	case "enter":
		if len(m.sheetInput.Value()) > 1 {
			url, convertErr := download.ConvertSheetURL(strings.TrimSuffix(m.sheetInput.Value(), "\n"))
			if convertErr != nil && convertErr.Error() == "invalid Google Sheets URL" {
				m.sheetInput.SetValue("")
				m.sheetInput.Focus()
				return m, nil
			}
			m.sheetInput.SetValue(url)
			m.csvChosen, downloadErr = download.DownloadFile(m.sheetInput.Value(), "SomeSheet.csv")
			if downloadErr != nil {
				panic(downloadErr)
			}

			m.artistChosen = true
			m.menuFocus = "start"
			homeDir, _ = os.UserHomeDir()

			m.columns, m.rows, readErr = filemgmt.ReadCSVFile(homeDir + "/Documents/tracker-tui/" + m.csvChosen)
			if readErr != nil {
				return m, tea.Quit
			}
			m.csvTable.SetColumns(m.columns)
			m.csvTable.SetRows(m.rows)
			m.csvTable.Focus()
			return m, nil
		}
	}

	m.sheetInput, cmd = m.sheetInput.Update(msg)
	return m, cmd
}

func playerControls(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.artistChosen = false
		m.sheetInput.SetValue("")
		m.menuFocus = "list"
		m.menuChoice = 0
		m.csvTable.Blur()
		return m, nil
	}
	m.csvTable, cmd = m.csvTable.Update(msg)
	return m, cmd
}
