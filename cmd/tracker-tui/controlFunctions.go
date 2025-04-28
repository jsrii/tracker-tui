package main

import (
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
			m.list.SetItems(items)
			m.menuFocus = "list"
		}
		return m, nil
	}
	return m, nil
}

func listControls(m model, msg tea.KeyMsg) (model, tea.Cmd) {

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
		index := m.list.Index()
		if _, ok := m.selected[index]; ok {
			delete(m.selected, index)
		} else {
			m.selected[index] = struct{}{}
		}
		m.csvChosen = m.list.SelectedItem().FilterValue()
		m.artistChosen = true
		m.sheetInput.SetValue("")
		m.menuFocus = "start"
		m.menuChoice = 0
		return m, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func sheetInputControls(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var downloadErr error
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
			m.downloadingFile = true

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

			return m, nil
		}
	}

	m.sheetInput, cmd = m.sheetInput.Update(msg)
	return m, cmd
}

func playerControls(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.artistChosen = false
		m.sheetInput.SetValue("")
		m.menuFocus = "start"
		m.menuChoice = 0
		return m, nil
	}
	return m, nil
}
