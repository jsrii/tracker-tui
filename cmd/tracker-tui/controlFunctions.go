package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tracker-tui/audio"
	"tracker-tui/download"
	"tracker-tui/filemgmt"
	"tracker-tui/styles"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gopxl/beep/speaker"
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

		m.columns, m.rows, readErr = filemgmt.ReadCSVFile(homeDir + "/Documents/tracker-tui/csv/" + m.csvChosen)
		if readErr != nil {
			return m, tea.Quit
		}

		mainColumns, mainRows, _ := filemgmt.GenerateMainTable(m.columns, m.rows)

		m.mainCSVTable.SetColumns(mainColumns)
		m.mainCSVTable.SetRows(mainRows)

		m.mainCSVTable.Focus()
		m.csvTableState = false
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
			clear(m.columns)
			clear(m.rows)
			url, convertErr := download.ConvertSheetURL(strings.TrimSuffix(m.sheetInput.Value(), "\n"))
			if convertErr != nil && convertErr.Error() == "invalid Google Sheets URL" {
				m.sheetInput.SetValue("")
				m.sheetInput.Focus()
				return m, nil
			}
			m.sheetInput.SetValue(url)
			m.csvChosen, downloadErr = download.DownloadFile(m.sheetInput.Value(), "SomeSheet.csv", true)
			if downloadErr != nil {
				fmt.Println(downloadErr)
				panic(downloadErr)
			}

			m.artistChosen = true
			m.menuFocus = "start"
			homeDir, _ = os.UserHomeDir()

			m.columns, m.rows, readErr = filemgmt.ReadCSVFile(homeDir + "/Documents/tracker-tui/csv/" + m.csvChosen)
			if readErr != nil {
				return m, tea.Quit
			}

			mainColumns, mainRows, _ := filemgmt.GenerateMainTable(m.columns, m.rows)

			m.mainCSVTable.SetColumns(mainColumns)
			m.mainCSVTable.SetRows(mainRows)

			m.mainCSVTable.Focus()
			m.tableWidth = 44
			return m, nil
		}
	}

	m.sheetInput, cmd = m.sheetInput.Update(msg)
	return m, cmd
}

func playerControls(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "left":
		if !m.controlState && m.pControlSelect > 0 {

			m.pControlSelect--
		}
		return m, nil
	case "right":
		if !m.controlState && m.pControlSelect <= 1 {
			m.pControlSelect++
		}
		return m, nil
	case "tab":
		m.controlState = !m.controlState
		if m.controlState {
			m.erasTable.Focus()
			m.mainCSVTable.Focus()

			altStyles := table.DefaultStyles()
			altStyles.Selected = styles.CsvTableSelectedStyle

			m.erasTable.SetStyles(altStyles)
			m.mainCSVTable.SetStyles(altStyles)
		} else {
			altStyles := table.DefaultStyles()
			altStyles.Selected = styles.CsvTableSelectedStyleAlt

			m.erasTable.SetStyles(altStyles)
			m.mainCSVTable.SetStyles(altStyles)

			m.erasTable.Blur()
			m.mainCSVTable.Blur()
		}
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.csvTableState {
			mainColumns, mainRows, _ := filemgmt.GenerateMainTable(m.columns, m.rows)

			m.mainCSVTable.SetColumns(mainColumns)
			m.mainCSVTable.SetRows(mainRows)
			m.csvTableState = false
			if m.controlState {
				m.mainCSVTable.Focus()
				m.erasTable.Blur()
			}
			m.tableWidth = 44
			return m, tea.ClearScreen
		} else {
			items, _ := filemgmt.ReturnListOfFiles()
			m.csvList.SetItems(items)

			m.artistChosen = false
			m.sheetInput.SetValue("")
			m.menuFocus = "list"
			m.menuChoice = 0
			if m.controlState {
				m.erasTable.Focus()
				m.mainCSVTable.Blur()
			}

			return m, tea.ClearScreen
		}

	case "enter", " ":
		if !m.controlState {
			switch m.pControlSelect {
			case 0:
				m.isDownloading = true
				m.erasTable.MoveUp(1)
				m.selectedLink = m.erasTable.SelectedRow()[len(m.erasTable.SelectedRow())-1]
				m.selectedSong = m.erasTable.SelectedRow()
				parsedLink, convertErr := download.ConvertLink(m.selectedLink)
				if convertErr != nil {
					return m, nil
				}

				return m, tea.Cmd(func() tea.Msg {
					fileName, downloadErr := download.DownloadFile(parsedLink, "somesong.mp3", false)
					if downloadErr != nil {
						return errMsg{err: downloadErr}
					}

					homeDir, _ := os.UserHomeDir()
					fullPath := filepath.Join(homeDir, "Documents", "tracker-tui", "songs", fileName)

					// Wait for file to exist, but still inside a goroutine
					for {
						if _, err := os.Stat(fullPath); err == nil {
							break
						}
						time.Sleep(1 * time.Second)
					}

					decodedFile, fileFormat, songErr := audio.ReturnPlayer(fullPath)
					if songErr != nil {
						return errMsg{err: songErr}
					}

					return audioReadyMsg{stream: decodedFile, format: fileFormat}
				})
			case 1:
				if m.isPlaying {
					speaker.Suspend()
					m.isPlaying = false
				} else {
					speaker.Resume()
					m.isPlaying = true
				}
			case 2:
				m.isDownloading = true
				m.erasTable.MoveDown(1)
				m.selectedLink = m.erasTable.SelectedRow()[len(m.erasTable.SelectedRow())-1]
				m.selectedSong = m.erasTable.SelectedRow()
				parsedLink, convertErr := download.ConvertLink(m.selectedLink)
				if convertErr != nil {
					return m, nil
				}

				return m, tea.Cmd(func() tea.Msg {
					fileName, downloadErr := download.DownloadFile(parsedLink, "somesong.mp3", false)
					if downloadErr != nil {
						return errMsg{err: downloadErr}
					}

					homeDir, _ := os.UserHomeDir()
					fullPath := filepath.Join(homeDir, "Documents", "tracker-tui", "songs", fileName)

					// Wait for file to exist, but still inside a goroutine
					for {
						if _, err := os.Stat(fullPath); err == nil {
							break
						}
						time.Sleep(1 * time.Second)
					}

					decodedFile, fileFormat, songErr := audio.ReturnPlayer(fullPath)
					if songErr != nil {
						return errMsg{err: songErr}
					}

					return audioReadyMsg{stream: decodedFile, format: fileFormat}
				})
			}
			return m, nil
		}
		if m.csvTableState {
			m.isDownloading = true
			m.selectedLink = m.erasTable.SelectedRow()[len(m.erasTable.SelectedRow())-1]
			m.selectedSong = m.erasTable.SelectedRow()
			parsedLink, convertErr := download.ConvertLink(m.selectedLink)
			if convertErr != nil {
				return m, nil
			}
			return m, tea.Cmd(func() tea.Msg {
				fileName, downloadErr := download.DownloadFile(parsedLink, "somesong.mp3", false)
				if downloadErr != nil {
					return errMsg{err: downloadErr}
				}

				homeDir, _ := os.UserHomeDir()
				fullPath := filepath.Join(homeDir, "Documents", "tracker-tui", "songs", fileName)

				// Wait for file to exist, but still inside a goroutine
				for {
					if _, err := os.Stat(fullPath); err == nil {
						break
					}
					time.Sleep(1 * time.Second)
				}

				decodedFile, fileFormat, songErr := audio.ReturnPlayer(fullPath)
				if songErr != nil {
					return errMsg{err: songErr}
				}

				return audioReadyMsg{stream: decodedFile, format: fileFormat}
			})

		} else {
			clear(m.erasColumns)
			clear(m.erasRows)
			m.erasColumns, m.erasRows, _ = filemgmt.GenerateEraTable(m.columns, m.rows, m.mainCSVTable.SelectedRow()[1])
			m.tableWidth = 0
			for i := range m.erasColumns {
				m.tableWidth += m.erasColumns[i].Width + 1
			}
			m.tableWidth = m.tableWidth - 1

			m.erasTable.SetColumns(m.erasColumns)
			m.erasTable.SetRows(m.erasRows)
			m.erasTable.Focus()
			m.mainCSVTable.Blur()
			m.csvTableState = true
			m.erasTable.HelpView()
			return m, tea.ClearScreen
		}

	}

	if m.csvTableState {
		m.erasTable, cmd = m.erasTable.Update(msg)
	} else {
		m.mainCSVTable, cmd = m.mainCSVTable.Update(msg)
	}
	return m, cmd
}
