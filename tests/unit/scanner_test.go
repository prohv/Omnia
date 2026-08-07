package unit

import (
	"os"
	"path/filepath"
	"testing"

	"omnia/internal/scanner"
)

func TestScanner(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// tempDir/
	//   file1.docx
	//   file2.png
	//   ~$temp.docx (should be skipped)
	//   .hidden.txt (should be skipped)
	//   subdir/
	//     file3.pdf

	_ = os.WriteFile(filepath.Join(tempDir, "file1.docx"), []byte("docx"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "file2.png"), []byte("png"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "~$temp.docx"), []byte("temp"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, ".hidden.txt"), []byte("hidden"), 0644)

	subDir := filepath.Join(tempDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file3.pdf"), []byte("pdf"), 0644)

	s := scanner.NewScanner()

	// Test non-recursive
	nonRecFiles, err := s.ScanPath(tempDir, false)
	if err != nil {
		t.Fatalf("unexpected error scanning non-recursive: %v", err)
	}
	if len(nonRecFiles) != 2 {
		t.Errorf("expected 2 non-recursive files, got %d", len(nonRecFiles))
	}

	// Test recursive
	recFiles, err := s.ScanPath(tempDir, true)
	if err != nil {
		t.Fatalf("unexpected error scanning recursive: %v", err)
	}
	if len(recFiles) != 3 {
		t.Errorf("expected 3 recursive files, got %d", len(recFiles))
	}
}
