package engines

import (
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"omnia/internal/jobs"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
)

// ImageEngine handles native Go image encoding, decoding, format conversion, and Image->PDF generation.
type ImageEngine struct{}

func NewImageEngine() *ImageEngine {
	return &ImageEngine{}
}

func (e *ImageEngine) Name() string {
	return "ImageEngine (Native Go Image)"
}

func (e *ImageEngine) CanHandle(job jobs.Job) bool {
	ext := strings.ToLower(filepath.Ext(job.InputPath))
	mime := strings.ToLower(job.MimeType)

	isImageInput := strings.HasPrefix(mime, "image/") ||
		ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".bmp" || ext == ".gif"

	if !isImageInput {
		return false
	}

	targetExt := strings.ToLower(job.TargetFormat)
	if targetExt == "" && job.OutputPath != "" {
		targetExt = strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")
	}

	if job.Operation == jobs.OperationConvert || job.Operation == jobs.OperationCompress {
		if targetExt == "pdf" || targetExt == "png" || targetExt == "jpg" || targetExt == "jpeg" || targetExt == "webp" || targetExt == "bmp" || targetExt == "gif" || targetExt == "" {
			return true
		}
	}

	return false
}

func (e *ImageEngine) Execute(ctx context.Context, job jobs.Job) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if job.OutputPath == "" {
		return fmt.Errorf("image: output path is required")
	}

	if err := os.MkdirAll(filepath.Dir(job.OutputPath), 0755); err != nil {
		return fmt.Errorf("image: failed to create output directory: %w", err)
	}

	targetExt := strings.ToLower(job.TargetFormat)
	if targetExt == "" {
		targetExt = strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")
	}

	// 1. Image -> PDF Conversion
	if targetExt == "pdf" {
		imp, err := api.Import("form:A4, pos:c", 0)
		if err != nil {
			imp = nil
		}
		err = api.ImportImagesFile([]string{job.InputPath}, job.OutputPath, imp, nil)
		if err != nil {
			return fmt.Errorf("image: failed to convert image to PDF: %w", err)
		}
		return nil
	}

	// 2. Image -> Image Conversion & Compression
	file, err := os.Open(job.InputPath)
	if err != nil {
		return fmt.Errorf("image: failed to open input image: %w", err)
	}
	defer file.Close()

	inputExt := strings.ToLower(filepath.Ext(job.InputPath))
	var img image.Image

	switch inputExt {
	case ".png":
		img, err = png.Decode(file)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".gif":
		img, err = gif.Decode(file)
	case ".bmp":
		img, err = bmp.Decode(file)
	case ".webp":
		img, err = webp.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}

	if err != nil {
		return fmt.Errorf("image: failed to decode input image: %w", err)
	}

	outFile, err := os.Create(job.OutputPath)
	if err != nil {
		return fmt.Errorf("image: failed to create output file: %w", err)
	}
	defer outFile.Close()

	quality := 85
	if job.GetOption("compression", "") == "high" {
		quality = 65
	} else if job.GetOption("compression", "") == "low" {
		quality = 95
	}

	switch targetExt {
	case "png":
		return png.Encode(outFile, img)
	case "jpg", "jpeg":
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	case "gif":
		return gif.Encode(outFile, img, nil)
	default:
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	}
}
