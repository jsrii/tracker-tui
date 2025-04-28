package download

// CODE FROM https://gist.github.com/cnu/026744b1e86c6d9e22313d06cba4c2e9
import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type WriteCounter struct {
	Total uint64
}

// Write implements io.Writer.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	return n, nil
}

func ConvertSheetURL(sheetURL string) (string, error) {
	re := regexp.MustCompile(`https://docs\.google\.com/spreadsheets/d/([a-zA-Z0-9-_]+)/edit\?gid=([0-9]+)`)

	// Match the URL
	matches := re.FindStringSubmatch(sheetURL)
	if len(matches) < 3 {
		return "", fmt.Errorf("invalid Google Sheets URL")
	}

	// Extract the spreadsheet ID and GID
	spreadsheetID := matches[1]
	sheetGID := matches[2]

	// Construct the CSV export URL
	downloadURL := fmt.Sprintf(
		"https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%s",
		spreadsheetID,
		sheetGID,
	)

	return downloadURL, nil
}

func DownloadFile(url string, fallbackFilename string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	downloadDir := filepath.Join(homeDir, "Documents", "tracker-tui")
	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	// Request the file
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Try to get filename from the "Content-Disposition" header
	contentDisposition := resp.Header.Get("Content-Disposition")
	filename := ""

	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if name, ok := params["filename"]; ok {
				filename = name
			}
		}
	}

	// If no filename found, fallback
	if filename == "" {
		filename = fallbackFilename
	}

	filename = sanitizeFilename(filename)

	filepath := filepath.Join(downloadDir, filename)

	// Create temp file
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return filename, err
	}
	defer out.Close()

	// Write data
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return filename, err
	}

	// Rename file
	err = os.Rename(filepath+".tmp", filepath)
	if err != nil {
		return filename, err
	}

	return filename, nil
}

// sanitizeFilename replaces all slashes and backslashes with underscores
func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[\\/]+`)
	return re.ReplaceAllString(name, "_")
}
