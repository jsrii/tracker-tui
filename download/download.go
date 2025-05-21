package download

// CODE FROM https://gist.github.com/cnu/026744b1e86c6d9e22313d06cba4c2e9
import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

func ConvertLink(input string) (string, error) {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	host := strings.ToLower(parsedURL.Host)
	if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be") {
		// Just return the original input for YouTube links
		return input, nil
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "f" {
		return "", fmt.Errorf("unexpected URL format")
	}

	downloadID := parts[1]
	newURL := "https://api.pillowcase.su/api/download/" + downloadID
	return newURL, nil
}

func DownloadFile(url string, fallbackFilename string, csvOrAudio bool) (string, error) {
	homeDir, err := os.UserHomeDir()
	var downloadDir string

	if strings.HasPrefix(url, "https://youtu") {
		filepath, error := downloadFromYT(url, fallbackFilename)
		return filepath, error
	}

	if err != nil {
		return "", err
	}
	if csvOrAudio {
		downloadDir = filepath.Join(homeDir, "Documents", "tracker-tui", "csv")
	} else {
		downloadDir = filepath.Join(homeDir, "Documents", "tracker-tui", "songs")

	}
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

func downloadFromYT(url string, fallbackFilename string) (string, error) {
	homeDir, _ := os.UserHomeDir()
	downloadDir := filepath.Join(homeDir, "Documents", "tracker-tui", "songs")

	err := os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	// yt-dlp -o "~/Documents/tracker-tui/songs/%(title)s.%(ext)s" -t mp3 https://youtu.be/sA3TpJzsFHc
	outputTemplate := fmt.Sprintf("%s/Documents/tracker-tui/songs/%s.%%(ext)s", homeDir, fallbackFilename)
	cmd := exec.Command(
		"yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"-o", outputTemplate,
		url,
	)
	// Run the command
	if err := cmd.Run(); err != nil {
		return "", nil
	}
	return fallbackFilename + ".mp3", nil
}

// sanitizeFilename replaces all slashes and backslashes with underscores
func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[\\/]+`)
	return re.ReplaceAllString(name, "_")
}
