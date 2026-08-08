# Omnia

> **Fast, intelligent, cross-platform file processing CLI in Go.**

Omnia is a modern command-line tool that orchestrates best-in-class open-source engines to convert, compress, process, and inspect documents, presentations, PDFs, and images. It follows a **"Native Go First"** philosophy—running 100% dependency-free for PDFs, Images, and Text, while delegating complex Office document rendering to LibreOffice (`soffice`).

---

## Key Features

- **Native Go First**: Images (`.png`, `.jpg`, `.webp`), PDFs (`.pdf`), and Text (`.docx`, `.pptx`, `.xlsx`, `.txt`) convert natively in Go with zero extra software required.
- **Universal Office to PDF**: Converts `.pptx`, `.docx`, `.xlsx`, `.odt`, and `.rtf` to visual PDFs using headless LibreOffice.
- **Same-Directory Output**: Created files save directly in the input file's directory by default (customizable via `--out`).
- **Verified PDF Cleanup**: PDF conversions verify new `.pdf` existence and non-zero file size before safely removing original non-PDF input files.
- **Existing PDF Auto-Skipping**: Input files already ending in `.pdf` are automatically detected and skipped, preserving them completely untouched.
- **High-Performance Worker Pool**: Concurrently processes folders of files using bounded Go worker goroutines (`min(CPU, 6)` default).
- **Magic-Number MIME Detection**: Uses file signature signatures instead of unreliable file extensions.
- **Rich Terminal UI**: Thread-safe live progress bars, ETA tracking, and execution summary tables powered by `pterm`.
- **Automated Setup & Doctor**: Built-in `omnia doctor` diagnostic tool and cross-platform `omnia setup` command to install dependencies via `winget`, `brew`, or `apt`.
- **Lightweight Binary**: Single static executable compressed with UPX (~4.6 MB).

---

## Technology Stack

| Component | Choice | Scope |
| :--- | :--- | :--- |
| **Language** | Go | Fast static binary, native concurrency |
| **CLI Framework** | Cobra (`spf13/cobra`) | Industry-standard CLI command parsing |
| **Configuration** | Viper (`spf13/viper`) | `~/.omnia/config.yaml` + environment variables (`OMNIA_*`) |
| **PDF Engine** | `pdfcpu` | Native Go PDF compression, optimization, split/merge, metadata |
| **Image Engine** | `image/*`, `golang.org/x/image` | Native Go Image → PDF, PNG/JPG/WEBP/BMP conversions |
| **OpenXML Engine** | `archive/zip` + `encoding/xml` | Native Go text extraction from `.pptx`, `.docx`, `.xlsx` |
| **Office Engine** | LibreOffice Headless (`soffice`) | Render `.pptx`, `.docx`, `.xlsx` → PDF / Slide images |
| **Logging** | `log/slog` | Structured JSON/Text logging |

---

## Quick Start

### 1. Installation

The easiest way to install Omnia globally is using Go:

```bash
go install github.com/prohv/omnia@latest
```

*(Alternatively, you can build from source by cloning the repository and running `go build .`)*

### 2. Automated Setup

Omnia requires LibreOffice to process Office documents natively. You do not need to hunt for installers! Just run:

```bash
omnia setup
```

This command automatically detects your OS and installs the necessary background engines via your native package manager (`winget` on Windows, `brew` on macOS, or `apt` on Linux).

### 3. Verify Readiness

Check system engine readiness to ensure everything is perfect:
```bash
omnia doctor
```

---

## CLI Commands & Examples

### 1. Convert Files (`omnia convert`)
Convert single or multiple files to target formats (defaults to PDF):

```bash
# Convert presentation slides to PDF (saves in same directory & cleans original)
omnia convert presentation.pptx

# Convert Word document to PDF
omnia convert report.docx --to pdf

# Convert image to PDF (Native Go — 0 external tools needed)
omnia convert photo.png --to pdf

# Convert multiple specific files together
omnia convert doc1.docx pres2.pptx photo3.png --to pdf

# Extract plain text from PPTX/DOCX natively in 2ms
omnia convert presentation.pptx --to txt
```

### 2. Batch Process Directories Concurrently (`omnia process`)
Scan a folder and process all files in parallel using worker goroutines:

```bash
# Convert all files in a folder to PDF using 6 parallel workers
omnia process my_folder/ --to pdf

# Process directory recursively with 8 workers
omnia process "$HOME/Downloads" --workers 8 --to pdf --recursive
```

### 3. Compress Files (`omnia compress`)
Optimize PDF or Image file sizes:

```bash
# Compress PDF file size
omnia compress document.pdf --level high

# Compress image quality
omnia compress photo.jpg --level balanced
```

### 4. File Information (`omnia info`)
Inspect magic-number MIME type signatures, file sizes, and routing details:

```bash
omnia info document.pdf
```

---

## Internal Architecture

```text
                                    OMNIA CLI
      (omnia process / omnia convert / omnia compress / omnia info / omnia doctor)
                                        │
                                  File Scanner
                                        │
                             MIME Magic Detector
                                        │
                                   Smart Router
                                        │
                                 Pipeline Planner
                                        │
                               Bounded Worker Pool
                                        │
                    ┌───────────────────┴───────────────────┐
                    │                                       │
            Native Go Engines                       External Engine
         - PDFCPUEngine (pdfcpu)                  - LibreOfficeEngine
         - ImageEngine (image/*)                    (soffice --headless)
         - OpenXMLEngine (zip+xml)
                    │                                       │
                    └───────────────────┬───────────────────┘
                                        │
                             pterm Progress UI & Summary
```

---

## Testing

Run the full unit test suite:

```bash
go test -v ./tests/...
```

---

## License

Distributed under the MIT License. See `LICENSE` for more information.
