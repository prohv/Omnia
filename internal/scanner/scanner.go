package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scanner handles discovery and path collection for files and directories.
type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

// ScanPath collects target files from a file path or directory.
func (s *Scanner) ScanPath(rootPath string, recursive bool) ([]string, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("scanner: failed to access path %s: %w", rootPath, err)
	}

	// Single file
	if !info.IsDir() {
		if s.shouldSkip(rootPath, info) {
			return []string{}, nil
		}
		abs, err := filepath.Abs(rootPath)
		if err != nil {
			return nil, err
		}
		return []string{abs}, nil
	}

	var files []string

	if recursive {
		err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Skip output directory or hidden directories
				name := info.Name()
				if name == "output" || (strings.HasPrefix(name, ".") && name != ".") {
					return filepath.SkipDir
				}
				return nil
			}

			if !s.shouldSkip(path, info) {
				abs, err := filepath.Abs(path)
				if err == nil {
					files = append(files, abs)
				}
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, fmt.Errorf("scanner: failed to read directory %s: %w", rootPath, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			fullPath := filepath.Join(rootPath, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if !s.shouldSkip(fullPath, info) {
				abs, err := filepath.Abs(fullPath)
				if err == nil {
					files = append(files, abs)
				}
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("scanner: error walking path %s: %w", rootPath, err)
	}

	return files, nil
}

func (s *Scanner) shouldSkip(path string, info os.FileInfo) bool {
	name := info.Name()

	// Skip hidden files
	if strings.HasPrefix(name, ".") {
		return true
	}
	// Skip office temp files (~$file.docx)
	if strings.HasPrefix(name, "~$") {
		return true
	}
	// Skip executables and plan/prd files
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".exe" || ext == ".dll" || ext == ".out" || ext == ".test" {
		return true
	}

	return false
}
