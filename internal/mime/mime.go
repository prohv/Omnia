package mime

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// DetectMime uses magic-number detection to identify the MIME type of a file.
func DetectMime(filePath string) (string, error) {
	mtype, err := mimetype.DetectFile(filePath)
	if err != nil {
		// Fallback to extension matching if file cannot be read
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".pdf":
			return "application/pdf", nil
		case ".docx":
			return "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil
		case ".pptx":
			return "application/vnd.openxmlformats-officedocument.presentationml.presentation", nil
		case ".xlsx":
			return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
		case ".png":
			return "image/png", nil
		case ".jpg", ".jpeg":
			return "image/jpeg", nil
		case ".txt":
			return "text/plain", nil
		default:
			return "application/octet-stream", fmt.Errorf("failed to detect MIME for %s: %w", filePath, err)
		}
	}
	return mtype.String(), nil
}
