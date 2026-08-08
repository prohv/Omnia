# Changelog

All notable changes to the **Omnia** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-08-08

### Added
- **Core CLI Infrastructure**:
  - Root command CLI parser with Cobra (`cmd/root.go`).
  - Configuration system powered by Viper (`internal/config`) supporting `~/.omnia/config.yaml`, env variables (`OMNIA_*`), and CLI flags.
  - Structured logger (`internal/logger`) powered by Go standard `log/slog` supporting `default`, `verbose`, `quiet`, and `json` output modes.
- **Engine Abstraction Layer**:
  - Uniform `Engine` interface and thread-safe `Registry` (`pkg/engines/`).
  - Native PDF Engine (`pkg/engines/pdfcpu.go`) for compression, optimization, split, merge, and metadata.
  - Native Image Engine (`pkg/engines/image.go`) for image format conversion (PNG, JPG, WEBP, BMP, GIF), quality compression, and Image-to-PDF generation.
  - Native OpenXML Engine (`pkg/engines/openxml.go`) for high-speed text extraction from `.pptx`, `.docx`, `.xlsx`, `.txt`, `.md`.
  - External LibreOffice Engine (`pkg/engines/libreoffice.go`) for visual rendering of Office files (`.pptx`, `.docx`, `.xlsx`, `.odt`, `.rtf`) to PDF and slide images.
- **Processing Pipeline**:
  - Magic-number file MIME detection (`internal/mime`) using `gabriel-vasile/mimetype`.
  - Recursive directory scanner (`internal/scanner`) skipping hidden files and temporary artifacts.
  - Smart router (`internal/router`) resolving target formats to the optimal engine.
  - Pipeline planner (`internal/planner`) generating execution plans.
- **Concurrency & User Experience**:
  - Bounded Goroutine Worker Pool (`internal/worker`) defaulting to `min(runtime.NumCPU(), 6)` workers.
  - Live progress tracking and batch summary tables (`internal/progress`) using `pterm`.
  - Multi-file and multi-directory arguments support for `omnia process` and `omnia convert`.
- **Diagnostics & Automated Setup**:
  - `omnia doctor` command for engine verification and path diagnostics.
  - `omnia setup` command for automated cross-platform installation of missing dependencies via `winget`, `brew`, or `apt`.
  - `omnia info` command for inspecting magic-number MIME signatures, file sizes, and routing details.
  - `omnia version` command for displaying build information and OS/Arch details.
- **Testing & Packaging**:
  - Dedicated unit testing suite in `tests/unit/` (13 test suites passing).
  - Executable size reduction via `-ldflags="-s -w"` and UPX compression (reduced binary size down to ~4.6 MB).
