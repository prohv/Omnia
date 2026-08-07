package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"omnia/internal/jobs"
	"omnia/pkg/engines"
)

func TestEngineRegistry(t *testing.T) {
	reg := engines.NewRegistry()

	pdfEngine := engines.NewPDFCPUEngine()
	imgEngine := engines.NewImageEngine()
	openxmlEngine := engines.NewOpenXMLEngine()
	officeEngine := engines.NewLibreOfficeEngine()

	reg.Register(pdfEngine)
	reg.Register(imgEngine)
	reg.Register(openxmlEngine)
	reg.Register(officeEngine)

	if len(reg.ListEngines()) != 4 {
		t.Fatalf("expected 4 registered engines, got %d", len(reg.ListEngines()))
	}

	// 1. Resolve PDF Job
	pdfJob := jobs.Job{
		ID:        "j-pdf",
		InputPath: "document.pdf",
		Operation: jobs.OperationCompress,
	}
	e, err := reg.GetEngineForJob(pdfJob)
	if err != nil {
		t.Errorf("failed to resolve PDF engine: %v", err)
	}
	if e.Name() != pdfEngine.Name() {
		t.Errorf("expected PDFCPUEngine, got %s", e.Name())
	}

	// 2. Resolve Image Job
	imgJob := jobs.Job{
		ID:           "j-img",
		InputPath:    "photo.png",
		Operation:    jobs.OperationConvert,
		TargetFormat: "pdf",
	}
	e, err = reg.GetEngineForJob(imgJob)
	if err != nil {
		t.Errorf("failed to resolve Image engine: %v", err)
	}
	if e.Name() != imgEngine.Name() {
		t.Errorf("expected ImageEngine, got %s", e.Name())
	}

	// 3. Resolve OpenXML Job
	xmlJob := jobs.Job{
		ID:           "j-xml",
		InputPath:    "report.docx",
		Operation:    jobs.OperationExtractText,
		TargetFormat: "txt",
	}
	e, err = reg.GetEngineForJob(xmlJob)
	if err != nil {
		t.Errorf("failed to resolve OpenXML engine: %v", err)
	}
	if e.Name() != openxmlEngine.Name() {
		t.Errorf("expected OpenXMLEngine, got %s", e.Name())
	}

	// 4. Resolve Office Document Job
	officeJob := jobs.Job{
		ID:           "j-doc",
		InputPath:    "report.docx",
		Operation:    jobs.OperationConvert,
		TargetFormat: "pdf",
	}
	e, err = reg.GetEngineForJob(officeJob)
	if err != nil {
		t.Errorf("failed to resolve Office engine: %v", err)
	}
	if e.Name() != officeEngine.Name() {
		t.Errorf("expected LibreOfficeEngine, got %s", e.Name())
	}
}

func TestImageEngineExecution(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy 1x1 image file
	srcPath := filepath.Join(tempDir, "source.png")
	dstPath := filepath.Join(tempDir, "output.jpg")

	// 1x1 red PNG pixel bytes
	pngBytes := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xdd, 0x8d, 0xb0, 0x00, 0x00, 0x00,
		0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(srcPath, pngBytes, 0644); err != nil {
		t.Fatalf("failed to write test PNG file: %v", err)
	}

	engine := engines.NewImageEngine()
	job := jobs.Job{
		ID:           "test-img-conv",
		InputPath:    srcPath,
		OutputPath:   dstPath,
		Operation:    jobs.OperationConvert,
		TargetFormat: "jpg",
	}

	if !engine.CanHandle(job) {
		t.Fatalf("ImageEngine should be able to handle PNG to JPG conversion")
	}

	ctx := context.Background()
	if err := engine.Execute(ctx, job); err != nil {
		t.Fatalf("ImageEngine execution failed: %v", err)
	}

	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Errorf("expected converted output file %s to exist", dstPath)
	}
}
