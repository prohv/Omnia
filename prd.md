# Omnia v1.0 — Architecture Blueprint (Native Go + LibreOffice Edition)

> **Vision:** Omnia is a fast, intelligent, cross-platform file processing CLI written in Go. It relies on **Native Go Engines** for Images and PDFs, and integrates **Headless LibreOffice (`soffice`)** as the single external engine for Office document conversions (`.pptx`, `.docx`, `.xlsx`) to PDF.

---

# 1. Core Principles & Engine Strategy

- **Native Go First**: Images (`.png`, `.jpg`, `.webp`) and PDFs (`.pdf`) convert 100% natively in Go with zero external software requirements.
- **Single External Dependency**: LibreOffice (`soffice --headless`) serves as the definitive engine for visual rendering of Office files (`.pptx`, `.docx`, `.xlsx`) to PDF.
- **Universal PDF Target**: Converting any file without parameters defaults to producing a clean PDF output.
- **Cross-Platform**: Seamless operation on Windows, macOS, and Linux.

---

# 2. Technology Stack & Engines

| Component | Choice | Scope | Dependency Needed? |
| :--- | :--- | :--- | :--- |
| **PDF Engine** | `pdfcpu` (`pdfcpu.io`) | PDF compression, metadata, split/merge, optimization | **NO (Native Go)** |
| **Image Engine** | Go standard library (`image/*`, `golang.org/x/image`) | Image → PDF, format conversion (PNG/JPG/WEBP), resizing | **NO (Native Go)** |
| **OpenXML Engine** | Go `archive/zip` + `encoding/xml` | Fast native text extraction from `.pptx`, `.docx`, `.xlsx` | **NO (Native Go)** |
| **Office Engine** | LibreOffice Headless (`soffice`) | Render `.pptx`, `.docx`, `.xlsx`, `.odt`, `.rtf` → PDF / Images | **YES (`soffice`)** |

---

# 3. Processing Pipeline

```
                     OMNIA CLI (omnia convert file.ext)
                                     │
                             MIME Detector
                                     │
             ┌───────────────────────┴───────────────────────┐
             │                                               │
    Office Documents                                  Images / PDFs
   (.pptx, .docx, .xlsx)                           (.png, .jpg, .pdf)
             │                                               │
             v                                               v
    LibreOffice Engine                              Native Go Engines
   (soffice --headless)                           (pdfcpu & image/*)
             │                                               │
             └───────────────────────┬───────────────────────┘
                                     │
                               Output: .pdf
```

---

# 4. Engine Interface

Every engine implements a uniform Go interface:

```go
type Engine interface {
    Name() string
    CanHandle(job jobs.Job) bool
    Execute(ctx context.Context, job jobs.Job) error
}
```

Implementations:
- `PDFCPUEngine` (Native Go)
- `ImageEngine` (Native Go)
- `OpenXMLEngine` (Native Go)
- `LibreOfficeEngine` (External `soffice` process invocation)

---

# 5. Project Structure

```text
cmd/
    root.go
    convert.go
    compress.go
    process.go
    doctor.go
    version.go
internal/
    scanner/
    router/
    planner/
    worker/
    jobs/
    config/
    logger/
    dependencies/
    progress/
    mime/
pkg/
    engines/
        engine.go
        pdfcpu.go
        image.go
        openxml.go
        libreoffice.go
        registry.go
tests/
    unit/
        config_test.go
        logger_test.go
        dependencies_test.go
main.go
```

---

# 6. Dependency Manager & `omnia doctor`

`omnia doctor` checks for the single external binary requirement:

```go
exec.LookPath("soffice")
```

Diagnostic Output:
- `✔ Native PDF Engine (pdfcpu) ....... Built-in (Native Go)`
- `✔ Native Image Engine (image/*) ..... Built-in (Native Go)`
- `✔ Native OpenXML Engine ............. Built-in (Native Go)`
- `✔ / ✗ LibreOffice Engine (soffice) .. Installed / Missing`

Installation instructions if `soffice` is missing:
- **Windows**: `winget install LibreOffice.LibreOffice`
- **macOS**: `brew install --cask libreoffice`
- **Linux**: `sudo apt install libreoffice`

---

# 7. Roadmap

- **Phase 1 (Completed ✔)**: Foundation CLI, Viper Config, `log/slog` Logger, `internal/dependencies` (`soffice` checker), `omnia doctor` & `version`.
- **Phase 2**: Engine wrappers (`pdfcpu.go`, `image.go`, `openxml.go`, `libreoffice.go`, `registry.go`).
- **Phase 3**: Scanner, MIME detection (`mimetype`), Router & Planner.
- **Phase 4**: Goroutine Worker Pool & `pterm` Progress UI.
- **Phase 5**: Commands integration & Cross-platform release builds.
