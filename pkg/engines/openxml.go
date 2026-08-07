package engines

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"omnia/internal/jobs"
)

// OpenXMLEngine extracts raw text from OpenXML files (.docx, .pptx, .xlsx) natively in Go.
type OpenXMLEngine struct{}

func NewOpenXMLEngine() *OpenXMLEngine {
	return &OpenXMLEngine{}
}

func (e *OpenXMLEngine) Name() string {
	return "OpenXMLEngine (Native Text Extraction)"
}

func (e *OpenXMLEngine) CanHandle(job jobs.Job) bool {
	ext := strings.ToLower(filepath.Ext(job.InputPath))
	isOfficeInput := ext == ".docx" || ext == ".pptx" || ext == ".xlsx"

	if !isOfficeInput {
		return false
	}

	targetExt := strings.ToLower(job.TargetFormat)
	if targetExt == "" && job.OutputPath != "" {
		targetExt = strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")
	}

	if job.Operation == jobs.OperationExtractText || targetExt == "txt" {
		return true
	}

	return false
}

func (e *OpenXMLEngine) Execute(ctx context.Context, job jobs.Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if job.OutputPath == "" {
		return fmt.Errorf("openxml: output path is required")
	}

	r, err := zip.OpenReader(job.InputPath)
	if err != nil {
		return fmt.Errorf("openxml: failed to open zip archive: %w", err)
	}
	defer r.Close()

	var extractedText strings.Builder

	ext := strings.ToLower(filepath.Ext(job.InputPath))
	switch ext {
	case ".docx":
		for _, f := range r.File {
			if f.Name == "word/document.xml" {
				text, err := extractTextFromXML(f)
				if err == nil {
					extractedText.WriteString(text)
				}
			}
		}

	case ".pptx":
		for _, f := range r.File {
			if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
				text, err := extractTextFromXML(f)
				if err == nil {
					extractedText.WriteString(fmt.Sprintf("--- %s ---\n", filepath.Base(f.Name)))
					extractedText.WriteString(text)
					extractedText.WriteString("\n\n")
				}
			}
		}

	case ".xlsx":
		for _, f := range r.File {
			if f.Name == "xl/sharedStrings.xml" {
				text, err := extractTextFromXML(f)
				if err == nil {
					extractedText.WriteString(text)
				}
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
		return fmt.Errorf("openxml: failed to create output directory: %w", err)
	}

	return os.WriteFile(job.OutputPath, []byte(extractedText.String()), 0644)
}

func extractTextFromXML(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return "", err
	}

	decoder := xml.NewDecoder(&buf)
	var textBuilder strings.Builder

	for {
		tok, err := decoder.Token()
		if err == io.EOF || tok == nil {
			break
		}
		switch t := tok.(type) {
		case xml.CharData:
			trimmed := strings.TrimSpace(string(t))
			if trimmed != "" {
				textBuilder.WriteString(trimmed)
				textBuilder.WriteString(" ")
			}
		}
	}

	return textBuilder.String(), nil
}
