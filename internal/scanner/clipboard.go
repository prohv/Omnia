package scanner

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ReadClipboardPaths reads newline-delimited file paths from system clipboard.
func ReadClipboardPaths() ([]string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-Command", "Get-Clipboard")
	case "darwin":
		cmd = exec.Command("pbpaste")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var validPaths []string

	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		trimmed = strings.Trim(trimmed, "\"")
		trimmed = strings.Trim(trimmed, "'")

		if trimmed != "" {
			if _, err := os.Stat(trimmed); err == nil {
				validPaths = append(validPaths, trimmed)
			}
		}
	}

	return validPaths, nil
}
