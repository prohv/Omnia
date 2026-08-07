package dependencies

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

// Dependency represents an external executable or native engine requirement.
type Dependency struct {
	Name         string
	BinaryName   string
	Purpose      string
	IsNative     bool
	Installed    bool
	Path         string
	InstallGuide string
}

// CheckDependencies inspects system dependencies and native engine availability.
func CheckDependencies() []Dependency {
	results := []Dependency{}

	// 1. Native PDF Engine (pdfcpu)
	results = append(results, Dependency{
		Name:         "Native PDF Engine (pdfcpu)",
		BinaryName:   "pdfcpu",
		Purpose:      "PDF Compression, Optimization, Metadata & Operations",
		IsNative:     true,
		Installed:    true,
		Path:         "Built-in (Native Go)",
		InstallGuide: "Native Go library integrated into binary",
	})

	// 2. Native Image Engine (image/*)
	results = append(results, Dependency{
		Name:         "Native Image Engine (image/*)",
		BinaryName:   "image/*",
		Purpose:      "Image Processing, Format Conversion & PDF Encoding",
		IsNative:     true,
		Installed:    true,
		Path:         "Built-in (Native Go)",
		InstallGuide: "Native Go standard library integrated into binary",
	})

	// 3. Native OpenXML Engine (zip + xml)
	results = append(results, Dependency{
		Name:         "Native OpenXML Engine (zip+xml)",
		BinaryName:   "zip+xml",
		Purpose:      "High-speed Text Extraction (.pptx, .docx, .xlsx)",
		IsNative:     true,
		Installed:    true,
		Path:         "Built-in (Native Go)",
		InstallGuide: "Native Go standard library integrated into binary",
	})

	// 4. External Office Engine (LibreOffice / soffice)
	officeDep := checkOfficeDependency()
	results = append(results, officeDep)

	return results
}

func checkOfficeDependency() Dependency {
	dep := Dependency{
		Name:       "LibreOffice Engine (soffice)",
		BinaryName: "soffice",
		Purpose:    "Office Documents (.docx, .pptx, .xlsx, etc.) -> PDF",
		IsNative:   false,
	}

	// Candidates to check via LookPath
	candidates := []string{"soffice", "libreoffice"}
	var foundPath string

	for _, cand := range candidates {
		if path, err := exec.LookPath(cand); err == nil {
			foundPath = path
			break
		}
	}

	// Fallback platform-specific paths if LookPath didn't catch soffice
	if foundPath == "" && runtime.GOOS == "windows" {
		winPaths := []string{
			`C:\Program Files\LibreOffice\program\soffice.exe`,
			`C:\Program Files (x86)\LibreOffice\program\soffice.exe`,
		}
		for _, p := range winPaths {
			if cleanPath, err := filepath.Abs(p); err == nil {
				if _, err := exec.LookPath(cleanPath); err == nil {
					foundPath = cleanPath
					break
				}
			}
		}
	}

	if foundPath != "" {
		dep.Installed = true
		dep.Path = foundPath
		dep.InstallGuide = "Installed and ready"
	} else {
		dep.Installed = false
		dep.Path = "Not Found"
		dep.InstallGuide = getOfficeInstallGuide()
	}

	return dep
}

func getOfficeInstallGuide() string {
	switch runtime.GOOS {
	case "windows":
		return "Run: winget install TheDocumentFoundation.LibreOffice OR download from https://www.libreoffice.org"
	case "darwin":
		return "Run: brew install --cask libreoffice"
	default:
		return "Run: sudo apt update && sudo apt install libreoffice (or package manager equivalent)"
	}
}
