package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/wav"
)

func ReturnPlayer(filePath string) (beep.StreamSeekCloser, beep.Format, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, beep.Format{}, err
	}

	fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))

	switch fileExt {
	case "wav":
		return wav.Decode(f)
	case "mp3":
		return mp3.Decode(f)
	case "flac":
		return flac.Decode(f)
	default:
		return nil, beep.Format{}, fmt.Errorf("unsupported file format: %s", fileExt)
	}
}
